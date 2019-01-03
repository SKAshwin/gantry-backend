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

//ListEvents is a handler which, given a username in the http request
//Returns all the information regarding the events belonging to that user
var ListEvents = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get(auth.JWTUsername)
	events, err := event.GetAll(username)
	if err != nil {
		log.Println("Error in ListEvents: " + err.Error())
		response.WriteMessage(http.StatusInternalServerError, "Error fetching user's events", w)
		return
	}
	type Events []event.Event
	reply, _ := json.Marshal(map[string]Events{"events": events})
	w.Write(reply)
})

//GetEvent is a handler, which given a username in the http request and a eventID in the URL
var GetEvent = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//TODO handle authorization in an accesscontrol fashion
	username := r.Header.Get(auth.JWTUsername)
	eventID := mux.Vars(r)["eventID"]
	if ok, err := event.CheckHost(username, eventID); err != nil {
		log.Println("Error checking if request is authorized: " + err.Error())
		response.WriteMessage(http.StatusInternalServerError, "Error checking if request is authorized", w)
		return
	} else if !ok {
		response.WriteMessage(http.StatusForbidden, "User is not host of the stated event (or event does not exist)", w)
		return
	} else {
		ev, err := event.Get(eventID)
		if err != nil {
			log.Println("Error fetching event: " + err.Error())
			response.WriteMessage(http.StatusInternalServerError, "Error fetching event", w)
		} else {
			reply, _ := json.Marshal(map[string]event.Event{"event": ev})
			w.Write(reply)
		}
	}
})

//CreateEvent creates an event
var CreateEvent = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var eventData event.Event
	err := json.NewDecoder(r.Body).Decode(&eventData)
	eventData.ID = uuid.New().String()
	if err != nil {
		log.Println("Error decoding event JSON: " + err.Error())
		response.WriteMessage(http.StatusBadRequest, "Invalid arguments to create event", w)
		return
	}
	err = eventData.Create(r.Header.Get(auth.JWTUsername))
	if err != nil {
		log.Println("Error in creating event: " + err.Error())
		response.WriteMessage(http.StatusInternalServerError, "Error in creating event", w)
	} else {
		response.WriteOKMessage("Event created", w)
	}
})
