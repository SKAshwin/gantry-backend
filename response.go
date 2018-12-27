package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type Response struct {
	Message string `json:"message"`
}

func WriteMessage(statusCode int, message string, w http.ResponseWriter) {
	if statusCode >= 400 { //if the message is an error message
		log.Println("writeError:", message)
	}
	w.WriteHeader(statusCode)
	reply, _ := json.Marshal(Response{Message: message})
	w.Write(reply)
}

func WriteOKMessage(message string, w http.ResponseWriter) {
	WriteMessage(http.StatusOK, message, w)
}
