package gorillawebsocket

import (
	myhttp "checkin/http"
	"errors"
	"net/http"

	"github.com/gorilla/websocket"
)

//GuestMessenger is an implementation of http.GuestMessenger which uses gorilla/websocket to set up
//a websocket connection with a guest, and allows for communication with the guest
type GuestMessenger struct {
	connections map[string]*GuestClient
	upgrader    websocket.Upgrader
}

//NewGuestMessenger creates a GuestMessenger with a given read/write buffer size
//in each websocket connection with each guest
//Set both to 1024 if you don't know what you want
func NewGuestMessenger(readBufferSize int, writeBufferSize int) *GuestMessenger {
	return &GuestMessenger{
		connections: make(map[string]*GuestClient),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  readBufferSize,
			WriteBufferSize: writeBufferSize,
			CheckOrigin: func(r *http.Request) bool {
				return true //CORS handles origins for us
				//GuestMessenger offers no guarantee on origin checking
			},
		},
	}
}

//OpenConnection opens a websocket connection with the guest, saving that connection under the given
//guest ID, given a responsewriter and request from said guest.
func (gm *GuestMessenger) OpenConnection(guestID string, w http.ResponseWriter, r *http.Request) error {
	conn, err := gm.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return errors.New("Error upgrading request to websocket connection: " + err.Error())
	}

	gm.connections[guestID] = newGuestClient(conn, func(){
		delete(gm.connections, guestID)
	})
	go gm.connections[guestID].writePump() //start the write pump that reads from the client's send channel
	
	return nil
}

//Send sends the provided guest data over the websocket connection marked by the given guestID
//Must have called OpenConnection with that guestID beforehand
func (gm *GuestMessenger) Send(guestID string, data myhttp.GuestMessage) error {
	if client, ok := gm.connections[guestID]; ok {
		client.send <- data
		err := <- client.errChannel
		if err != nil {
			return errors.New("Error writing message: " + err.Error())
		}

		return nil
	}

	return errors.New("No such guest ID")
}

//HasConnection returns true if there is an active connection with the given guest ID
func (gm *GuestMessenger) HasConnection(guestID string) bool {
	_, ok := gm.connections[guestID]
	return ok
}

//CloseConnection closes the websocket connection marked with the given guestID
func (gm *GuestMessenger) CloseConnection(guestID string) error {
	if client, ok := gm.connections[guestID]; ok {
		err := client.close()
		delete(gm.connections, guestID)
		return err
	}
	return errors.New("No such connection with that guest ID exists")
}
