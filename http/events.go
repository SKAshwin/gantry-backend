package http

import (
	"bytes"
	"checkin"
	"encoding/csv"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/gorilla/mux"
)

//EventHandler An extension of mux.Router which handles all event-related requests
//Uses the given EventService, the given Logger, and a given Authenticator to check if
//requests are valid
//Also contains a GuestHandler to handle all the subset of event-related requests
//that deal with guests
//Call NewEventHandler to initialize an EventHandler with the correct routes
type EventHandler struct {
	*mux.Router
	GuestHandler     *GuestHandler
	EventService     checkin.EventService
	Logger           *log.Logger
	Authenticator    Authenticator
	MaxLengthName    int
	MaxLengthURL     int
	MaxLengthTimeTag int
}

//NewEventHandler Creates a new event handler using gorilla/mux for routing
//and the default Logger.
//GuestHandler, EventService, Authenticator needs to be set by the calling function
//API endpoint changes happen here, as well as changes to the routing library and logger to be used
//and type of authenticator
func NewEventHandler(es checkin.EventService, auth Authenticator, gh *GuestHandler, maxLengthName, maxLengthURL, maxLengthTimeTag int) *EventHandler {
	h := &EventHandler{
		Router:           mux.NewRouter(),
		Logger:           log.New(os.Stderr, "", log.LstdFlags),
		Authenticator:    auth,
		EventService:     es,
		GuestHandler:     gh,
		MaxLengthName:    maxLengthName,
		MaxLengthURL:     maxLengthURL,
		MaxLengthTimeTag: maxLengthTimeTag,
	}
	//Adapters to check if handler should serve the request
	tokenCheck := checkAuth(auth, h.Logger)
	credentialsCheck := isAdminOrHost(auth, es, "eventID", h.Logger)
	existCheck := eventExists(es, "eventID", h.Logger)

	h.Handle("/api/v1-3/events", Adapt(http.HandlerFunc(h.handleEventsBy),
		tokenCheck, correctTimezonesOutput, jsonSelector)).Methods("GET")
	h.Handle("/api/v1-3/events", Adapt(http.HandlerFunc(h.handleCreateEvent),
		tokenCheck, correctTimezonesInput)).Methods("POST")
	h.Handle("/api/v0/events/takenurls/{eventURL}", Adapt(http.HandlerFunc(h.handleURLTaken),
		tokenCheck)).Methods("GET")
	h.Handle("/api/v1-3/events/{eventID}", Adapt(http.HandlerFunc(h.handleEvent),
		tokenCheck, existCheck, credentialsCheck, correctTimezonesOutput, jsonSelector)).Methods("GET")
	h.Handle("/api/v1-3/events/{eventID}", Adapt(http.HandlerFunc(h.handleUpdateEvent),
		tokenCheck, existCheck, credentialsCheck, correctTimezonesInput)).Methods("PATCH")
	h.Handle("/api/v0/events/{eventID}", Adapt(http.HandlerFunc(h.handleDeleteEvent),
		tokenCheck, existCheck, credentialsCheck)).Methods("DELETE")
	h.Handle("/api/v0/events/{eventID}/released", Adapt(http.HandlerFunc(h.handleReleased),
		existCheck)).Methods("GET")
	h.Handle("/api/v1-3/events/{eventID}/triggers/{triggername}", Adapt(http.HandlerFunc(h.handleGetTimeTag),
		existCheck, correctTimezonesOutput)).Methods("GET")
	h.Handle("/api/v1-3/events/{eventID}/triggers/{triggername}/occurred", Adapt(http.HandlerFunc(h.handleTimeTagOccurred),
		existCheck)).Methods("GET")
	h.Handle("/api/v1-2/events/{eventID}/feedback", Adapt(http.HandlerFunc(h.handleSubmitForm),
		existCheck)).Methods("POST")
	h.Handle("/api/v1-2/events/{eventID}/feedback/report", Adapt(http.HandlerFunc(h.handleFeedbackReport),
		tokenCheck, existCheck, credentialsCheck)).Methods("GET")
	//route all guest-related requests to the guest handler
	h.PathPrefix("/api/{versionNumber}/events/{eventID}/guests").Handler(gh)

	return h
}

func (h *EventHandler) handleGetTimeTag(w http.ResponseWriter, r *http.Request) {
	event, err := h.EventService.Event(mux.Vars(r)["eventID"])
	if err != nil {
		h.Logger.Println("Error fetching event details: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Could not fetch event information due to internal server issue", w)
		return
	}
	tag := strings.ToLower(mux.Vars(r)["triggername"])
	if val, ok := event.TimeTags[tag]; !ok {
		WriteMessage(http.StatusNotFound, "No such time tag found", w)
	} else {
		reply, _ := json.Marshal(val)
		w.Write(reply)
	}
}

func (h *EventHandler) handleTimeTagOccurred(w http.ResponseWriter, r *http.Request) {
	event, err := h.EventService.Event(mux.Vars(r)["eventID"])
	if err != nil {
		h.Logger.Println("Error fetching event details: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Could not fetch event information due to internal server issue", w)
		return
	}
	tag := strings.ToLower(mux.Vars(r)["triggername"])
	if val, ok := event.TimeTags[tag]; !ok {
		WriteMessage(http.StatusNotFound, "No such time tag found", w)
	} else {
		reply, _ := json.Marshal(val.Before(time.Now()))
		w.Write(reply)
	}
}

//Takes a feedback form encoded in JSON, anonymous or otherwise, and writes it into the
//permanent storage
func (h *EventHandler) handleSubmitForm(w http.ResponseWriter, r *http.Request) {
	var ff checkin.FeedbackForm
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(&ff)
	if err != nil {
		h.Logger.Println("Error parsing JSON body in SubmitForm: " + err.Error())
		WriteMessage(http.StatusBadRequest, "JSON could not be decoded: must be in the format {name, survey:[{question, answer},...]}", w)
		return
	}

	if ff.Survey == nil || len(ff.Survey) == 0 {
		WriteMessage(http.StatusBadRequest, "Feedback form cannot have null or empty survey", w)
		return
	}

	err = h.EventService.SubmitFeedback(mux.Vars(r)["eventID"], ff)
	if err != nil {
		h.Logger.Println("Error submitting feedback: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error writing feedback form into database", w)
	} else {
		WriteOKMessage("Form submitted successfully", w)
	}
}

func (h *EventHandler) handleFeedbackReport(w http.ResponseWriter, r *http.Request) {
	forms, err := h.EventService.FeedbackForms(mux.Vars(r)["eventID"])
	if err != nil {
		h.Logger.Println("Error fetching feedback forms: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error fetching feedback forms", w)
		return
	}

	b := &bytes.Buffer{}
	wr := csv.NewWriter(b)
	if len(forms) != 0 {
		//get all the unique questions (different forms might have different questions in the same event)
		//for this event
		questions := h.uniqueQuestions(forms)
		wr.Write(append([]string{"Name"}, questions...))
		for _, form := range forms {
			row := make([]string, len(questions)+1)
			row[0] = form.Name
			//for every question in each form
			//check the current question headers
			//see if that question was asked in this form
			//if not, answer for this form should be ""
			//if it was, put that answer in the csv cell
			for i, question := range questions {
				formHasQuestion := false
				for _, formItem := range form.Survey {
					if formItem.Question == question {
						row[i+1] = formItem.Answer
						formHasQuestion = true
					}
				}
				if !formHasQuestion {
					row[i+1] = ""
				}
			}
			wr.Write(row)

		}
	}
	wr.Flush()

	w.Header().Set("Content-Type", "text/csv")
	//set the file name here
	w.Header().Set("Content-Disposition", "attachment;filename=Feedback.csv")
	w.Write(b.Bytes())
}

func (h *EventHandler) uniqueQuestions(forms []checkin.FeedbackForm) []string {
	questions := make([]string, 0, 20)
	for _, form := range forms {
		for _, formItem := range form.Survey {
			unique := true
			for _, question := range questions {
				if question == formItem.Question {
					unique = false
				}
			}
			if unique == true {
				questions = append(questions, formItem.Question)
			}
		}
	}

	return questions
}

func (h *EventHandler) handleReleased(w http.ResponseWriter, r *http.Request) {
	event, err := h.EventService.Event(mux.Vars(r)["eventID"])
	if err != nil {
		h.Logger.Println("Error fetching event info: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error in fetching event info", w)
		return
	}

	reply, _ := json.Marshal(event.TimeTags["release"].Before(time.Now()))
	w.Write(reply)
}

//handleEventsBy is a handler which, given a username in the http request
//Returns all the information regarding the events belonging to that user
func (h *EventHandler) handleEventsBy(w http.ResponseWriter, r *http.Request) {
	authInfo, err := h.Authenticator.GetAuthInfo(r)
	if err != nil {
		h.Logger.Println("Error fetching authorization info: " + err.Error())
		WriteMessage(http.StatusBadRequest, "Error in fetching authorization info", w)
		return
	}

	events, err := h.EventService.EventsBy(authInfo.Username)
	if err != nil {
		h.Logger.Println("Error in GetUsersEvents: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error fetching user's events", w)
		return
	}
	reply, _ := json.Marshal(events)
	w.Write(reply)
}

//handleEvent is a handler, which given a eventID in the URL, writes that event's details
func (h *EventHandler) handleEvent(w http.ResponseWriter, r *http.Request) {
	eventID := mux.Vars(r)["eventID"]
	ev, err := h.EventService.Event(eventID)
	if err != nil {
		h.Logger.Println("Error fetching event: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error fetching event", w)
	} else {
		reply, _ := json.Marshal(ev)
		w.Write(reply)
	}
}

//handleDeleteEvent deletes the event given by the eventID provided in the endpoint
func (h *EventHandler) handleDeleteEvent(w http.ResponseWriter, r *http.Request) {
	eventID := mux.Vars(r)["eventID"]
	err := h.EventService.DeleteEvent(eventID)
	if err != nil {
		h.Logger.Println("Error deleting event: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error deleting user", w)
	} else {
		WriteOKMessage("Successfully deleted event", w)
	}
}

//handleCreateEvent creates an event
func (h *EventHandler) handleCreateEvent(w http.ResponseWriter, r *http.Request) {
	var eventData checkin.Event
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(&eventData)
	if err != nil {
		h.Logger.Println("Error decoding event JSON: " + err.Error())
		WriteMessage(http.StatusBadRequest, "Badly formatted JSON in event (Possibly invalid time format or invalid fields)", w)
		return
	}
	if !h.validCreateEventInputs(eventData) {
		WriteMessage(http.StatusBadRequest, "Invalid arguments to create event", w)
		return
	}

	if exists, err := h.EventService.URLExists(eventData.URL.String); err != nil {
		//check if the URL provided is available
		h.Logger.Println("Error checking if URL already taken: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error checking if URL is available", w)
		return
	} else if exists {
		WriteMessage(http.StatusConflict, "URL already used by another event", w)
		return
	}

	eventData.ID = uuid.New().String()
	authInfo, err := h.Authenticator.GetAuthInfo(r)
	if err != nil {
		h.Logger.Println("Error fetching authorization info: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error in fetching authorization info", w)
		return
	}

	err = h.EventService.CreateEvent(eventData, authInfo.Username)
	if err != nil {
		h.Logger.Println("Error in creating event: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error in creating event", w)
	} else {
		WriteMessage(http.StatusCreated, "Event created", w)
	}
}

//make sure the event creation data is valid
func (h *EventHandler) validCreateEventInputs(event checkin.Event) bool {
	for key := range event.TimeTags { //check that all time tags of create input aren't too long in label
		if len(key) > h.MaxLengthTimeTag || key == "" {
			return false
		}
	}
	return !(event.URL.String == "" && event.URL.Valid) && event.Name != "" && event.UpdatedAt == time.Time{} && event.CreatedAt == time.Time{} && len(event.URL.String) <= h.MaxLengthURL && len(event.Name) <= h.MaxLengthName
}

//checks that no empty string or too long strings are involved in update data
func (h *EventHandler) validUpdateEventInputs(event checkin.Event) bool {
	for key := range event.TimeTags { //check that all time tags of create input aren't too long in label
		if len(key) > h.MaxLengthTimeTag || key == "" {
			return false
		}
	}
	//don't allow empty string URLs
	//but allow null URLs
	return !(event.URL.String == "" && event.URL.Valid) && event.Name != "" && len(event.URL.String) <= h.MaxLengthURL && len(event.Name) <= h.MaxLengthName
}

//handleUpdateEvent updates the event given by the eventID provided in the endpoint
//using the fields provided in the body of the request
//Only need to supply the fields that need updating
func (h *EventHandler) handleUpdateEvent(w http.ResponseWriter, r *http.Request) {
	//Load original event, marshal JSON into it
	//This updates only the fields that were supplied
	event, err := h.EventService.Event(mux.Vars(r)["eventID"])
	if err != nil {
		h.Logger.Println("Error fetching original event: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Could not fetch original event", w)
		return
	}

	originalURL, originalCreatedAt, originalUpdatedAt := event.URL, event.CreatedAt, event.UpdatedAt

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	h.Logger.Println("Before", event.URL)
	err = dec.Decode(&event)
	if err != nil {
		h.Logger.Println("Error when decoding update fields: " + err.Error())
		WriteMessage(http.StatusBadRequest, "JSON could not be decoded (Possibly invalid time format or unknown fields)", w)
		return
	}
	h.Logger.Println(event.URL)

	//validate inputs
	if (event.ID != mux.Vars(r)["eventID"]) || (event.UpdatedAt != originalUpdatedAt) ||
		(event.CreatedAt != originalCreatedAt) {
		//if caller trying to update these non-updatable fields
		WriteMessage(http.StatusBadRequest, "Cannot update ID or update and create times", w)
		return
	} else if !h.validUpdateEventInputs(event) { //otherwise check that the object you have is valid
		WriteMessage(http.StatusBadRequest, "Cannot set name or URL or timetag label to empty string, or longer than 64 bytes", w)
		return
	}

	if event.URL != originalURL { //if the caller is attempting to update the url
		if ok, err := h.EventService.URLExists(event.URL.String); err != nil {
			h.Logger.Println("Error checking if URL taken: " + err.Error())
			WriteMessage(http.StatusInternalServerError, "Error checking if URL already taken", w)
			return
		} else if ok {
			WriteMessage(http.StatusConflict, "URL already exists", w)
			return
		}
	}

	err = h.EventService.UpdateEvent(event)
	if err != nil {
		h.Logger.Println("Error updating user: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error updating event", w)
	} else {
		WriteOKMessage("Event updated", w)
	}
}

//handleURLExists Checks if the eventURL provided in the endpoint is already used
func (h *EventHandler) handleURLTaken(w http.ResponseWriter, r *http.Request) {
	url := mux.Vars(r)["eventURL"]
	if exists, err := h.EventService.URLExists(url); err != nil {
		h.Logger.Println("Error checking if URL already taken: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error checking if URL exists", w)
	} else {
		reply, _ := json.Marshal(exists)
		w.Write(reply)
	}
}

//An Adapter generator which produces an adapter which checks if
//an event exists before allowing the handler to execute
//Returns a 404 otherwise (or a 500 if an error occurred when checking
//if event exists)
func eventExists(es checkin.EventService, eventIDKey string, logger *log.Logger) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			eventID := mux.Vars(r)[eventIDKey]
			ok, err := es.CheckIfExists(eventID)
			if err != nil {
				logger.Println("Error checking that event exists: " + err.Error())
				WriteMessage(http.StatusInternalServerError, "Error checking if event exists", w)
			} else if ok {
				h.ServeHTTP(w, r)
			} else {
				WriteMessage(http.StatusNotFound, "Event does not exist with that ID", w)
			}
		})
	}
}

func eventReleased(es checkin.EventService, eventIDKey string, logger *log.Logger) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			eventID := mux.Vars(r)[eventIDKey]
			event, err := es.Event(eventID)
			if err != nil {
				logger.Println("Error fetching event data in eventReleased: " + err.Error())
				WriteMessage(http.StatusInternalServerError, "Error fetching event data", w)
			} else if event.TimeTags["release"].Before(time.Now()) {
				h.ServeHTTP(w, r)
			} else {
				WriteMessage(http.StatusForbidden, "Event has not been released yet", w)
			}
		})
	}
}

//CURRENTLY UNUSED/UNTESTED (SHOULD BE FINE THOUGH)
//handleEvents is a handler which returns all information pertaining to all events
//func (h *EventHandler) handleEvents(w http.ResponseWriter, r *http.Request) {
//	events, err := h.EventService.Events()
//	if err != nil {
//		h.Logger.Println("Error in GetAllEvents: " + err.Error())
//		WriteMessage(http.StatusInternalServerError, "Error fetching all events", w)
//		return
//	}
//	reply, _ := json.Marshal(events)
//	w.Write(reply)
//}
