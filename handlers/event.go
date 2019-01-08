package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"registration-app/auth"
	"registration-app/event"
	"registration-app/response"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

//GetAllEvents is a handler which returns all information pertaining to all events
var GetAllEvents = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	events, err := event.GetAll()
	if err != nil {
		log.Println("Error in GetAllEvents: " + err.Error())
		response.WriteMessage(http.StatusInternalServerError, "Error fetching all events", w)
		return
	}
	type Events []event.Event
	reply, _ := json.Marshal(map[string]Events{"events": events})
	w.Write(reply)
})

//GetUsersEvents is a handler which, given a username in the http request
//Returns all the information regarding the events belonging to that user
var GetUsersEvents = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get(auth.JWTUsername)
	events, err := event.GetAllBy(username)
	if err != nil {
		log.Println("Error in GetUsersEvents: " + err.Error())
		response.WriteMessage(http.StatusInternalServerError, "Error fetching user's events", w)
		return
	}
	type Events []event.Event
	reply, _ := json.Marshal(map[string]Events{"events": events})
	w.Write(reply)
})

//GetEvent is a handler, which given a username in the http request and a eventID in the URL
var GetEvent = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	eventID := mux.Vars(r)["eventID"]

	ev, err := event.Get(eventID)
	if err != nil {
		log.Println("Error fetching event: " + err.Error())
		response.WriteMessage(http.StatusInternalServerError, "Error fetching event", w)
	} else {
		reply, _ := json.Marshal(map[string]event.Event{"event": ev})
		w.Write(reply)
	}
})

//CreateEvent creates an event
var CreateEvent = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var eventData event.Event
	err := json.NewDecoder(r.Body).Decode(&eventData)
	if err != nil {
		log.Println("Error decoding event JSON: " + err.Error())
		response.WriteMessage(http.StatusBadRequest, "Invalid arguments to create event", w)
		return
	}

	if exists, err := event.URLExists(eventData.URL); err != nil { //check if the URL provided is available
		log.Println("Error checking if URL already taken: " + err.Error())
		response.WriteMessage(http.StatusInternalServerError, "Error checking if URL is available", w)
		return
	} else if exists {
		response.WriteMessage(http.StatusConflict, "URL already used by another event", w)
		return
	}

	eventData.ID = uuid.New().String()
	err = eventData.Create(r.Header.Get(auth.JWTUsername))
	if err != nil {
		log.Println("Error in creating event: " + err.Error())
		response.WriteMessage(http.StatusInternalServerError, "Error in creating event", w)
	} else {
		response.WriteOKMessage("Event created", w)
	}
})

//DeleteEvent deletes the event given by the eventID provided in the endpoint
var DeleteEvent = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	eventID := mux.Vars(r)["eventID"]
	err := event.Delete(eventID)
	if err != nil {
		log.Println("Error deleting event: " + err.Error())
		response.WriteMessage(http.StatusInternalServerError, "Error deleting user", w)
	} else {
		response.WriteOKMessage("Successfully deleted event", w)
	}
})

//UpdateEvent updates the event given by the eventID provided in the endpoint
//using the fields provided in the body of the request
//Only need to supply the fields that need updating
var UpdateEvent = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var updatedFields map[string]string
	err := json.NewDecoder(r.Body).Decode(&updatedFields)
	if err != nil {
		log.Println("Error when decoding update fields: " + err.Error())
		response.WriteMessage(http.StatusBadRequest, "JSON could not be decoded", w)
		return
	}

	if val, ok := updatedFields["url"]; ok { //if the caller is attempting to update the url
		if ok, err := event.URLExists(val); err != nil {
			log.Println("Error checking if URL taken: " + err.Error())
			response.WriteMessage(http.StatusInternalServerError, "Error checking if URL already taken", w)
			return
		} else if ok {
			response.WriteMessage(http.StatusConflict, "URL already exists", w)
			return
		}
	}

	eventID := mux.Vars(r)["eventID"] //middleware already confirms event exists
	validRequest, err := event.Update(eventID, updatedFields)

	if err != nil {
		log.Println("Error updating user: " + err.Error())
		response.WriteMessage(http.StatusInternalServerError, "Error updating event", w)
	} else if !validRequest {
		response.WriteMessage(http.StatusBadRequest, "Incorrect fields for event update", w)
	} else {
		response.WriteOKMessage("Event updated", w)
	}
})

//EventURLAvailable Checks if the eventURL provided in the endpoint is already used
var EventURLAvailable = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	url := mux.Vars(r)["eventURL"]
	if exists, err := event.URLExists(url); err != nil {
		log.Println("Error checking if URL already taken: " + err.Error())
		response.WriteMessage(http.StatusInternalServerError, "Error checking if URL exists", w)
	} else {
		reply, _ := json.Marshal(map[string]bool{"available": !exists})
		w.Write(reply)
	}

})

//EventExists middleware which checks if the eventID in the
//URL points to an event which actually exists
//Outputs a 404 if the event does not exist
func EventExists(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		eventID := mux.Vars(r)["eventID"]
		if exists, err := event.CheckIfExists(eventID); err != nil {
			log.Println("Error checking if event exists" + err.Error())
			response.WriteMessage(http.StatusInternalServerError, "Error checking if event exists", w)
		} else if !exists {
			response.WriteMessage(http.StatusNotFound, "Event does not exist", w)
		} else {
			h.ServeHTTP(w, r)
		}
	})
}
