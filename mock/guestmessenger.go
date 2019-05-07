package mock

import (
	myhttp "checkin/http"
	"net/http"
)

//GuestMessenger is a mock implementation of http.GuestMessenger, which takes mock functions as attributes
//and calls them/marks them as invoked when a http.GuestMessenger function is called
type GuestMessenger struct {
	OpenConnectionFn       func(guestID string, w http.ResponseWriter, r *http.Request) error
	OpenConnectionInvoked  bool
	SendFn                 func(guestID string, msg myhttp.GuestMessage) error
	SendInvoked            bool
	HasConnectionFn        func(guestID string) bool
	HasConnectionInvoked   bool
	CloseConnectionFn      func(guestID string) error
	CloseConnectionInvoked bool
}

//OpenConnection calls the mock function attribute (part of the struct) and marks it as invoked
func (gm *GuestMessenger) OpenConnection(guestID string, w http.ResponseWriter, r *http.Request) error {
	gm.OpenConnectionInvoked = true
	return gm.OpenConnectionFn(guestID, w, r)
}

//Send calls the mock function attribute (part of the struct) and marks it as invoked
func (gm *GuestMessenger) Send(guestID string, msg myhttp.GuestMessage) error {
	gm.SendInvoked = true
	return gm.SendFn(guestID, msg)
}

//HasConnection calls the mock function attribute (part of the struct) and marks it as invoked
func (gm *GuestMessenger) HasConnection(guestID string) bool {
	gm.HasConnectionInvoked = true
	return gm.HasConnectionFn(guestID)
}

//CloseConnection calls the mock function attribute (part of the struct) and marks it as invoked
func (gm *GuestMessenger) CloseConnection(guestID string) error {
	gm.CloseConnectionInvoked = true
	return gm.CloseConnectionFn(guestID)
}
