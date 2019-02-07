package http

import (
	"encoding/json"
	"log"
	"net/http"
)

//response is the standard json message format of {"message":"contents"}
type response struct {
	Message string `json:"message"`
}

//WriteMessage writes a message in the standard format using the given
//ResponseWriter, with the given status code
//If the status code is 400 and above, an error message will also be logged
func WriteMessage(statusCode int, message string, w http.ResponseWriter) {
	if statusCode >= 400 { //if the message is an error message
		log.Println("writeError:", message)
	}
	w.WriteHeader(statusCode)
	reply, _ := json.Marshal(response{Message: message})
	w.Write(reply)
}

//WriteOKMessage does the same thing as WriteMessage, but the
//status code of the response is set to 200 (Status OK)
func WriteOKMessage(message string, w http.ResponseWriter) {
	WriteMessage(http.StatusOK, message, w)
}
