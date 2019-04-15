package gorillawebsocket

import (
	"checkin"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/websocket"
)

type GuestMessenger struct {
	connections map[string]*websocket.Conn
	upgrader    websocket.Upgrader
}

func NewGuestMessenger(readBufferSize int, writeBufferSize int) *GuestMessenger {
	return &GuestMessenger{
		connections: make(map[string]*websocket.Conn),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

func (gm *GuestMessenger) OpenConnection(guestID string, w http.ResponseWriter, r *http.Request) error {
	conn, err := gm.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return errors.New("Error upgrading request to websocket connection: " + err.Error())
	}

	gm.connections[guestID] = conn
	return nil
}

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

func (gm *GuestMessenger) CloseConnection(guestID string) error {
	return nil
}
