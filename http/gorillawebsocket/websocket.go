package gorillawebsocket

import (
	myhttp "checkin/http"
	"errors"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

//GuestMessenger is an implementation of http.GuestMessenger which uses gorilla/websocket to set up
//a websocket connection with a guest, and allows for communication with the guest
type GuestMessenger struct {
	connections map[string][]*GuestConnection
	upgrader    websocket.Upgrader
	lock        sync.RWMutex
}

//NewGuestMessenger creates a GuestMessenger with a given read/write buffer size
//in each websocket connection with each guest
//Set both to 1024 if you don't know what you want
func NewGuestMessenger(readBufferSize int, writeBufferSize int) *GuestMessenger {
	return &GuestMessenger{
		connections: make(map[string][]*GuestConnection),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  readBufferSize,
			WriteBufferSize: writeBufferSize,
			CheckOrigin: func(r *http.Request) bool {
				return true //CORS handles origins for us
				//GuestMessenger offers no guarantee on origin checking
			},
		},
		lock: sync.RWMutex{},
	}
}

//OpenConnection opens a websocket connection with the guest, saving that connection under the given
//guest ID, given a responsewriter and request from said guest.
func (gm *GuestMessenger) OpenConnection(guestID string, w http.ResponseWriter, r *http.Request) error {
	gm.lock.Lock()
	defer gm.lock.Unlock()
	conn, err := gm.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return errors.New("Error upgrading request to websocket connection: " + err.Error())
	}

	newConn := newGuestConnection(conn, func() {
		gm.lock.Lock()
		defer gm.lock.Unlock()
		delete(gm.connections, guestID)
	})

	if conns, ok := gm.connections[guestID]; !ok {
		//initialize it if it does not exist
		conns := make([]*GuestConnection, 1, 10)
		conns[0] = newConn
		gm.connections[guestID] = conns
	} else {
		gm.connections[guestID] = append(conns, newConn)
	}

	return nil
}

//Send sends the provided guest data over the websocket connections marked by the given guestID
//Must have called OpenConnection with that guestID beforehand
func (gm *GuestMessenger) Send(guestID string, data myhttp.GuestMessage) error {
	gm.lock.RLock()
	defer gm.lock.RUnlock()
	if clients, ok := gm.connections[guestID]; ok {
		response := make(chan error)
		errorList := make([]error, len(clients))
		anyerr := false
		for i, client := range clients {
			client.send <- SendTask{message: data, response: response}
			errorList[i] = <-response
			if errorList[i] != nil {
				anyerr = true
			}
		}

		if anyerr {
			errString := "Error(s) occured when sending message: "
			for _, err := range errorList {
				if err != nil {
					errString += err.Error() + "\n"
				}
			}
			return errors.New(errString)
		}

		return nil
	}
	return errors.New("No such guest ID")
}

//HasConnection returns true if there is at least one active connection with the given guest ID
func (gm *GuestMessenger) HasConnection(guestID string) bool {
	gm.lock.RLock()
	defer gm.lock.RUnlock()
	_, ok := gm.connections[guestID]
	return ok
}

//CloseConnection closes ALL the websocket connections marked with the given guestID
func (gm *GuestMessenger) CloseConnection(guestID string) error {
	gm.lock.Lock()
	defer gm.lock.Unlock()
	if clients, ok := gm.connections[guestID]; ok {
		for _, client := range clients {
			client.close()
		}
		delete(gm.connections, guestID)
		return nil
	}
	return errors.New("No such connection with that guest ID exists")
}
