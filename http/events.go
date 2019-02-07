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
//Uses the given EventService and the given Logger
//Call NewEventHandler to initialize an EventHandler with the correct routes
type EventHandler struct {
	*mux.Router
	EventService checkin.EventService
	Logger       *log.Logger
}

//NewEventHandler Creates a new event handler using gorilla/mux for routing
//and the default Logger.
//UserService needs to be set by the calling function
//API endpoint changes happen here, as well as changes to the routing library and logger to be used
func NewEventHandler() *EventHandler {
	h := &EventHandler{
		Router: mux.NewRouter(),
		Logger: log.New(os.Stderr, "", log.LstdFlags),
	}

	//ALL OF THESE NEED VALIDATION CHECKS LMAO
	h.Handle("/api/events/v0", http.HandlerFunc(h.handleEventsBy)).Methods("GET")
	h.Handle("/api/events/v0", http.HandlerFunc(h.handleCreateEvent)).Methods("POST")
	h.Handle("/api/events/v0/exists/{eventURL}", http.HandlerFunc(h.handleURLExists)).Methods("GET")
	//TODO add check to these methods to see if the event in question exists
	h.Handle("/api/events/v0/{eventID}", http.HandlerFunc(h.handleEvent)).Methods("GET")
	h.Handle("/api/events/v0/{eventID}", http.HandlerFunc(h.handleUpdateEvent)).Methods("PUT")
	h.Handle("/api/events/v0/{eventID}", http.HandlerFunc(h.handleDeleteEvent)).Methods("DELETE")

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
	username := r.Header.Get("username")
	events, err := h.EventService.EventsBy(username)
	if err != nil {
		h.Logger.Println("Error in GetUsersEvents: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error fetching user's events", w)
		return
	}
	reply, _ := json.Marshal(events)
	w.Write(reply)
}

//handleEvent is a handler, which given a username in the http request and a eventID in the URL
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
	err = h.EventService.CreateEvent(eventData, r.Header.Get("username"))
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
