package http

import (
	"net/http"
)

//GuestMessenger allows for the server to communicate with guests
type GuestMessenger interface {
	//OpenConnection opens up a connection to a guest, given a request that came from them
	//Must be called before any other functions can be called communicating to that guest
	//guestID must be some identifier guaranteed to be unique for every guest
	OpenConnection(guestID string, w http.ResponseWriter, r *http.Request) error

	//Sends a message to the given guest
	Send(guestID string, msg GuestMessage) error

	//Closes a connection to a guest, preventing any further communications
	CloseConnection(guestID string) error
}

//GuestMessage is a message to be sent to a guest
type GuestMessage struct {
	Title   string      `json:"title"`
	Content interface{} `json:"content"`
}
