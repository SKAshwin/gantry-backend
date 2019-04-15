package http

import (
	"checkin"
	"net/http"
)

//GuestMessenger allows for the server to communicate with guests
type GuestMessenger interface {
	//OpenConnection opens up a connection to a guest, given a request that came from them
	//Must be called before any other functions can be called communicating to that guest
	//guestID must be some identifier guaranteed to be unique for every guest
	OpenConnection(guestID string, w http.ResponseWriter, r *http.Request) error

	//NotifyCheckIn notifies a given guest that they have been checked in
	//and passes them the guest data
	NotifyCheckIn(guestID string, data checkin.Guest) error

	//Closes a connection to a guest, preventing any further communications
	CloseConnection(guestID string) error
}
