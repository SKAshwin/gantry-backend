package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"registration-app/auth"
	"registration-app/event"
	"registration-app/response"
)

//ListEvents is a handler which, given a username in the http request
//Returns all the information regarding the events belonging to that user
var ListEvents = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//TODO check if user exists first?
	username := r.Header.Get(auth.JWTUsername)
	events, err := event.GetAll(username)
	if err != nil {
		log.Println("Error in ListEvents: " + err.Error())
		response.WriteMessage(http.StatusInternalServerError, "Error fetching user's events", w)
		return
	}
	type Events []event.Event
	reply, _ := json.Marshal(map[string]Events{"message": events})
	w.Write(reply)
})
