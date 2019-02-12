package mock

import (
	"checkin"
)

//GuestService represents a mock implementation of the checkin.GuestService interface
type GuestService struct {
	CheckInFn      func(eventID string, nric string) (string, error)
	CheckInInvoked bool

	GuestsFn      func(eventID string) ([]string, error)
	GuestsInvoked bool

	GuestsCheckedInFn      func(eventID string) ([]string, error)
	GuestsCheckedInInvoked bool

	GuestsNotCheckedInFn      func(eventID string) ([]string, error)
	GuestsNotCheckedInInvoked bool

	RegisterGuestFn      func(nric string, name string) error
	RegisterGuestInvoked bool

	RemoveGuestFn      func(nric string) error
	RemoveGuestInvoked bool

	CheckInStatsFn      func() (checkin.AttendanceStats, error)
	CheckInStatsInvoked bool
}

//CheckIn invokes the mock implementation and marks the function as invoked
func (as *GuestService) CheckIn(eventID string, nric string) (string, error) {
	as.CheckInInvoked = true
	return as.CheckInFn(eventID, nric)
}

//Guests invokes the mock implementation and marks the function as invoked
func (as *GuestService) Guests(eventID string) ([]string, error) {
	as.GuestsInvoked = true
	return as.GuestsFn(eventID)
}

//GuestsCheckedIn invokes the mock implementation and marks the function as invoked
func (as *GuestService) GuestsCheckedIn(eventID string) ([]string, error) {
	as.GuestsCheckedInInvoked = true
	return as.GuestsCheckedInFn(eventID)
}

//GuestsNotCheckedIn invokes the mock implementation and marks the function as invoked
func (as *GuestService) GuestsNotCheckedIn(eventID string) ([]string, error) {
	as.GuestsNotCheckedInInvoked = true
	return as.GuestsNotCheckedInFn(eventID)
}

//RegisterGuest invokes the mock implementation and marks the function as invoked
func (as *GuestService) RegisterGuest(nric string, name string) error {
	as.RegisterGuestInvoked = true
	return as.RegisterGuestFn(nric, name)
}

//RemoveGuest invokes the mock implementation and marks the function as invoked
func (as *GuestService) RemoveGuest(nric string) error {
	as.RemoveGuestInvoked = true
	return as.RemoveGuestFn(nric)
}

//CheckInStats invokes the mock implementation and marks the function as invoked
func (as *GuestService) CheckInStats() (checkin.AttendanceStats, error) {
	as.CheckInInvoked = true
	return as.CheckInStats()
}
