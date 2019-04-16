package gorillawebsocket

import (
	"checkin"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/websocket"
)

//GuestMessenger is an implementation of http.GuestMessenger which uses gorilla/websocket to set up
//a websocket connection with a guest, and allows for communication with the guest
type GuestMessenger struct {
	connections map[string]*websocket.Conn
	upgrader    websocket.Upgrader
}

//NewGuestMessenger creates a GuestMessenger with a given read/write buffer size
//in each websocket connection with each guest
//Set both to 1024 if you don't know what you want
func NewGuestMessenger(readBufferSize int, writeBufferSize int) *GuestMessenger {
	return &GuestMessenger{
		connections: make(map[string]*websocket.Conn),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  readBufferSize,
			WriteBufferSize: writeBufferSize,
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

	gm.connections[guestID] = conn
	return nil
}

//NotifyCheckIn sends the provided guest data over the websocket connection marked by the given guestID
//Must have called OpenConnection with that guestID beforehand
func (gm *GuestMessenger) NotifyCheckIn(guestID string, data checkin.Guest) error {
	msg, err := json.Marshal(data)
	if err != nil {
		return errors.New("Error marshalling guest data into JSON: " + err.Error())
	}

	if conn, ok := gm.connections[guestID]; ok {
		err = conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			return errors.New("Error writing message: " + err.Error())
		}

		return nil
	}

	return errors.New("No such guest ID")
}

//CloseConnection closes the websocket connection marked with the given guestID
func (gm *GuestMessenger) CloseConnection(guestID string) error {
	if conn, ok := gm.connections[guestID]; ok {
		err := conn.Close()
		delete(gm.connections, guestID)
		return err
	}
	return errors.New("No such connection with that guest ID exists")
}
