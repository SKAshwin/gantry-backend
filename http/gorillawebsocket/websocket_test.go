package gorillawebsocket_test

import (
	"checkin"
	myhttp "checkin/http"
	mywebsocket "checkin/http/gorillawebsocket"
	"checkin/test"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestGuestMessenger(t *testing.T) {
	gm := mywebsocket.NewGuestMessenger(2048, 2048)
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gm.OpenConnection(r.Header.Get("GuestID"), w, r)
	}))
	defer s.Close()

	url := "ws" + strings.TrimPrefix(s.URL, "http")
	header1 := make(http.Header)
	header1.Add("GuestID", "1234")
	ws, _, err := websocket.DefaultDialer.Dial(url, header1)
	test.Ok(t, err)

	//try notifying someone before a connection has been opened
	err = gm.Send("3000",
		myhttp.GuestMessage{Title: "Check in", Content: checkin.Guest{NRIC: "3000", Name: "Tim Smith"}})
	test.Assert(t, err != nil, "Communicating with non-existent guest fails to throw an error")

	//create the other connection
	header2 := make(http.Header)
	header2.Add("GuestID", "3000")
	ws2, _, err := websocket.DefaultDialer.Dial(url, header2)
	test.Ok(t, err)

	//Try notifying one websocket connection about a check in
	err = gm.Send("1234", myhttp.GuestMessage{Title: "Check in", Content: checkin.Guest{NRIC: "1234",
		Name: "Jim Bob"}})
	test.Ok(t, err)
	//the guest object should have been sent to this websocket connection
	type msg struct {
		Title   string        `json:"title"`
		Content checkin.Guest `json:"content"`
	}
	var guest msg
	err = ws.ReadJSON(&guest)
	test.Ok(t, err)
	test.Equals(t, msg{Title: "Check in", Content: checkin.Guest{NRIC: "1234", Name: "Jim Bob"}}, guest)

	//try communicating after closing a connection
	err = gm.CloseConnection("1234")
	test.Ok(t, err)
	err = gm.Send("1234", myhttp.GuestMessage{Title: "Check in", Content: checkin.Guest{NRIC: "1234", Name: "Jim Bob"}})
	test.Assert(t, err != nil, "Communicating after closed connection fails to throw an error")

	//You should be able to re-establish a connection to the same guest ID, and reuse the methods
	ws, _, err = websocket.DefaultDialer.Dial(url, header1)
	test.Ok(t, err)
	err = gm.Send("1234",
		myhttp.GuestMessage{Title: "Check in", Content: checkin.Guest{NRIC: "1234", Name: "Some other name"}})
	test.Ok(t, err)
	err = ws.ReadJSON(&guest)
	test.Ok(t, err)
	test.Equals(t, msg{Title: "Check in", Content: checkin.Guest{NRIC: "1234", Name: "Some other name"}}, guest)

	//See what happens if you close the connection from the other end abruptly
	//should not longer be listed as an active connection
	//given sufficient delay first
	ws.Close()
	time.Sleep(25 * time.Millisecond)
	test.Assert(t, !gm.HasConnection("1234"), "1234 still listed as a connection even though the websocket was closed by the other end")

	//Try the other connection
	gm.Send("3000", myhttp.GuestMessage{Title: "Check in", Content: checkin.Guest{NRIC: "3000", Name: "Tim Smith"}})
	//the guest object should have been sent to this websocket connection
	err = ws2.ReadJSON(&guest)
	test.Ok(t, err)
	test.Equals(t, msg{Title: "Check in", Content: checkin.Guest{NRIC: "3000", Name: "Tim Smith"}}, guest)

	//See what happens if you close the connection from the other end gracefully
	ws2.WriteMessage(websocket.CloseMessage, []byte{})
	time.Sleep(25 * time.Millisecond)
	test.Assert(t, !gm.HasConnection("3000"), "1234 still listed as a connection even though the websocket was closed by the other end")

	//shouldn't be able to close a connection that's already closed
	err = gm.CloseConnection("3000")
	test.Assert(t, err != nil, "Attempting to close already closed connection failed to throw an error")

}

func TestSendMultipleConnections(t *testing.T) {
	gm := mywebsocket.NewGuestMessenger(2048, 2048)
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := gm.OpenConnection(r.Header.Get("GuestID"), w, r)
		test.Ok(t, err)
	}))
	defer s.Close()

	type msg struct {
		Title   string        `json:"title"`
		Content checkin.Guest `json:"content"`
	}

	url := "ws" + strings.TrimPrefix(s.URL, "http")
	header1 := make(http.Header)
	header1.Add("GuestID", "1234")
	header2 := make(http.Header)
	header2.Add("GuestID", "2234")
	var wg sync.WaitGroup
	var reading sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		reading.Add(1)
		go func(n int) {
			defer wg.Done()
			var guest msg
			if n%2 == 0 {
				ws, _, err := websocket.DefaultDialer.Dial(url, header1)
				test.Ok(t, err)
				reading.Done()
				err = ws.ReadJSON(&guest)
				test.Ok(t, err)
				test.Equals(t, msg{Title: "Check in", Content: checkin.Guest{NRIC: "1234", Name: "Jim Bob"}}, guest)
			} else {
				ws, _, err := websocket.DefaultDialer.Dial(url, header2)
				test.Ok(t, err)
				reading.Done()
				err = ws.ReadJSON(&guest)
				test.Ok(t, err)
				test.Equals(t, msg{Title: "Check in", Content: checkin.Guest{NRIC: "2234", Name: "Jimothy Bob"}}, guest)
			}

		}(i)
	}
	reading.Wait()
	gm.Send("1234", myhttp.GuestMessage{Title: "Check in", Content: checkin.Guest{NRIC: "1234", Name: "Jim Bob"}})
	gm.Send("2234", myhttp.GuestMessage{Title: "Check in", Content: checkin.Guest{NRIC: "2234", Name: "Jimothy Bob"}})
	wg.Wait()
}
