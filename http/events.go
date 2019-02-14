package http

import (
	"checkin"
	"encoding/json"
	"log"
	"net/http"
	"os"

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
	GuestHandler  *GuestHandler
	EventService  checkin.EventService
	Logger        *log.Logger
	Authenticator Authenticator
}

//NewEventHandler Creates a new event handler using gorilla/mux for routing
//and the default Logger.
//GuestHandler, EventService, Authenticator needs to be set by the calling function
//API endpoint changes happen here, as well as changes to the routing library and logger to be used
//and type of authenticator
func NewEventHandler(es checkin.EventService, auth Authenticator, gh *GuestHandler) *EventHandler {
	h := &EventHandler{
		Router:        mux.NewRouter(),
		Logger:        log.New(os.Stderr, "", log.LstdFlags),
		Authenticator: auth,
		EventService:  es,
		GuestHandler:  gh,
	}
	//Adapters to check if handler should serve the request
	tokenCheck := checkAuth(auth)
	credentialsCheck := isAdminOrHost(auth, es, "eventID")
	existCheck := eventExists(es, "eventID")

	h.Handle("/api/v0/events", Adapt(http.HandlerFunc(h.handleEventsBy),
		tokenCheck)).Methods("GET")
	h.Handle("/api/v0/events", Adapt(http.HandlerFunc(h.handleCreateEvent),
		tokenCheck)).Methods("POST")
	h.Handle("/api/v0/events/exists/{eventURL}", Adapt(http.HandlerFunc(h.handleURLExists),
		tokenCheck)).Methods("GET")
	h.Handle("/api/v0/events/{eventID}", Adapt(http.HandlerFunc(h.handleEvent),
		tokenCheck, credentialsCheck, existCheck)).Methods("GET")
	h.Handle("/api/v0/events/{eventID}", Adapt(http.HandlerFunc(h.handleUpdateEvent),
		tokenCheck, credentialsCheck, existCheck)).Methods("PUT")
	h.Handle("/api/v0/events/{eventID}", Adapt(http.HandlerFunc(h.handleDeleteEvent),
		tokenCheck, credentialsCheck, existCheck)).Methods("DELETE")
	//route all guest-related requests to the guest handler
	h.PathPrefix("/api/v0/events/{eventID}/guests").Handler(gh)

	return h
}

//handleEvents is a handler which returns all information pertaining to all events
func (h *EventHandler) handleEvents(w http.ResponseWriter, r *http.Request) {
	events, err := h.EventService.Events()
	if err != nil {
		h.Logger.Println("Error in GetAllEvents: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error fetching all events", w)
		return
	}
	reply, _ := json.Marshal(events)
	w.Write(reply)
}

//handleEventsBy is a handler which, given a username in the http request
//Returns all the information regarding the events belonging to that user
func (h *EventHandler) handleEventsBy(w http.ResponseWriter, r *http.Request) {
	authInfo, err := h.Authenticator.GetAuthInfo(r)
	if err != nil {
		h.Logger.Println("Error fetching authorization info: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error in fetching authorization info", w)
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
	err := json.NewDecoder(r.Body).Decode(&eventData)
	if err != nil {
		h.Logger.Println("Error decoding event JSON: " + err.Error())
		WriteMessage(http.StatusBadRequest, "Invalid arguments to create event", w)
		return
	}

	if exists, err := h.EventService.URLExists(eventData.URL); err != nil {
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
	}

	err = h.EventService.CreateEvent(eventData, authInfo.Username)
	if err != nil {
		h.Logger.Println("Error in creating event: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error in creating event", w)
	} else {
		WriteOKMessage("Event created", w)
	}
}

//handleUpdateEvent updates the event given by the eventID provided in the endpoint
//using the fields provided in the body of the request
//Only need to supply the fields that need updating
func (h *EventHandler) handleUpdateEvent(w http.ResponseWriter, r *http.Request) {
	var updatedFields map[string]string
	err := json.NewDecoder(r.Body).Decode(&updatedFields)
	if err != nil {
		h.Logger.Println("Error when decoding update fields: " + err.Error())
		WriteMessage(http.StatusBadRequest, "JSON could not be decoded", w)
		return
	}

	if val, ok := updatedFields["url"]; ok { //if the caller is attempting to update the url
		if ok, err := h.EventService.URLExists(val); err != nil {
			h.Logger.Println("Error checking if URL taken: " + err.Error())
			WriteMessage(http.StatusInternalServerError, "Error checking if URL already taken", w)
			return
		} else if ok {
			WriteMessage(http.StatusConflict, "URL already exists", w)
			return
		}
	}

	eventID := mux.Vars(r)["eventID"] //middleware already confirms event exists
	validRequest, err := h.EventService.UpdateEvent(eventID, updatedFields)

	if err != nil {
		h.Logger.Println("Error updating user: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error updating event", w)
	} else if !validRequest {
		WriteMessage(http.StatusBadRequest, "Incorrect fields for event update", w)
	} else {
		WriteOKMessage("Event updated", w)
	}
}

//handleURLExists Checks if the eventURL provided in the endpoint is already used
func (h *EventHandler) handleURLExists(w http.ResponseWriter, r *http.Request) {
	url := mux.Vars(r)["eventURL"]
	if exists, err := h.EventService.URLExists(url); err != nil {
		h.Logger.Println("Error checking if URL already taken: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error checking if URL exists", w)
	} else {
		reply, _ := json.Marshal(map[string]bool{"available": !exists})
		w.Write(reply)
	}
}

func eventExists(es checkin.EventService, eventIDKey string) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			eventID := mux.Vars(r)[eventIDKey]
			ok, err := es.CheckIfExists(eventID)
			if err != nil {
				log.Println("Error checking that event exists: " + err.Error())
				WriteMessage(http.StatusInternalServerError, "Error checking if event exists", w)
			} else if ok {
				h.ServeHTTP(w, r)
			} else {
				WriteMessage(http.StatusNotFound, "Event does not exist with that ID", w)
			}
		})
	}
}
