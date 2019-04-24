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

// GuestClient is a wrapper around the connection to the guest, to server as a middle man
type GuestClient struct {
	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan http.GuestMessage

	// channel of errors, in response
	errChannel chan error
}

//newGuestClient creates a new guest client with the given connection
//upon the connection closing (with a close message or abruptly), the onCloseFunc will be run
func newGuestClient(conn *websocket.Conn, onCloseFunc func()) *GuestClient {
	go func() {
		for {
			messageType, _, err := conn.ReadMessage() //stalls until the connection is closed, then performs clean up
			if messageType == websocket.CloseMessage || err != nil {//either closed properly or abruptly
				onCloseFunc()
				break
			}
		}
	}()
	return &GuestClient {
		conn: conn,
		send: make(chan http.GuestMessage),
		errChannel: make(chan error),
	}
}

//neatly closes the connection and all relevant channels
func (gc *GuestClient) close() error {
	close(gc.send)
	close(gc.errChannel)
	return gc.conn.WriteMessage(websocket.CloseMessage, []byte{})
}

//all writes through the send channel through this pump
//as the write message function cannot be called concurrently
func (gc *GuestClient) writePump() {
	ticker := time.NewTicker(pingPeriod) //used to ping/pong to see if connection is open
	defer func() {
		ticker.Stop()
		gc.conn.Close()
	}()
	for {
		select {
		case message, ok := <-gc.send:
			gc.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				// close connection
				gc.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			err := gc.conn.WriteJSON(message)
			gc.errChannel <- err
		case <-ticker.C:
			gc.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := gc.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
