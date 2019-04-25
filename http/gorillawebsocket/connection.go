package gorillawebsocket

import (
	"checkin/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

// SendTask is a combination of a http.GuestMessage as well as a response channel for the results
// of sending a message to the guests (a possible error)
type SendTask struct {
	message http.GuestMessage
	response chan error
}

// GuestConnection is a wrapper around the connection to the guest, to serve as a middle man
// To enable control flow - as the writing on the connection cannot be done concurrently
// in order to write to the guest, pass the message and a response channel (wrapped in a SendTask) to the send channel
// listen in on the response channel for the results
type GuestConnection struct {
	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of tasks, consisting of messages to be sent out and response channels.
	send chan SendTask
}

// newGuestConnection creates a new guest connection wrapping the connection
// upon the connection closing (with a close message or abruptly), the onCloseFunc will be run
// starts the writePump, so will start accepting tasks from the send channel
func newGuestConnection(conn *websocket.Conn, onCloseFunc func()) *GuestConnection {
	go func() {
		for {
			messageType, _, err := conn.ReadMessage() //stalls until the connection is closed, then performs clean up
			if messageType == websocket.CloseMessage || err != nil {//either closed properly or abruptly
				onCloseFunc()
				break
			}
		}
	}()
	gc := &GuestConnection {
		conn: conn,
		send: make(chan SendTask),
	}

	go gc.writePump()

	return gc
}

// neatly closes the connection and all relevant channels
func (gc *GuestConnection) close() {
	close(gc.send)
}

// writePump is run in a separate goroutine
// waits on the send channel, will execute any SendTasks it gets
// also handles pinging, writePump routine will close when connection closes
func (gc *GuestConnection) writePump() {
	ticker := time.NewTicker(pingPeriod) //used to ping/pong to see if connection is open
	defer func() {
		ticker.Stop()
		gc.conn.Close()
	}()
	for {
		select {
		case sendTask, ok := <-gc.send:
			gc.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Channel was closed
				// close connection
				gc.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			err := gc.conn.WriteJSON(sendTask.message)
			sendTask.response <- err
		case <-ticker.C:
			gc.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := gc.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				// if there's an error pinging the other side, the connection has been closed
				// so end the writePump
				return
			}
		}
	}
}
