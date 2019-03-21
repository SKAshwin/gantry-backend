package mock

import (
	"checkin"
)

//GuestService represents a mock implementation of the checkin.GuestService interface
type GuestService struct {
	CheckInFn      func(eventID string, nric string) (string, error)
	CheckInInvoked bool

	AbsentFn      func(eventID string, nric string) error
	AbsentInvoked bool

	GuestsFn      func(eventID string) ([]string, error)
	GuestsInvoked bool

	GuestsCheckedInFn      func(eventID string) ([]string, error)
	GuestsCheckedInInvoked bool

	GuestsNotCheckedInFn      func(eventID string) ([]string, error)
	GuestsNotCheckedInInvoked bool

	GuestExistsFn      func(eventID string, nric string) (bool, error)
	GuestExistsInvoked bool

	RegisterGuestFn      func(eventID string, nric string, name string) error
	RegisterGuestInvoked bool

	RemoveGuestFn      func(eventID string, nric string) error
	RemoveGuestInvoked bool

	CheckInStatsFn      func(eventID string) (checkin.GuestStats, error)
	CheckInStatsInvoked bool
}

//CheckIn invokes the mock implementation and marks the function as invoked
func (as *GuestService) CheckIn(eventID string, nric string) (string, error) {
	as.CheckInInvoked = true
	return as.CheckInFn(eventID, nric)
}

//Absent invokes the mock implementation and marks the function as invoked
func (as *GuestService) Absent(eventID string, nric string) (string, error) {
	as.AbsentInvoked = true
	return as.Absent(eventID, nric)
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

//GuestExists invokes the mock implementation and marks the function as invoked
func (as *GuestService) GuestExists(eventID string, nric string) (bool, error) {
	as.GuestExistsInvoked = true
	return as.GuestExistsFn(eventID, nric)
}

//RegisterGuest invokes the mock implementation and marks the function as invoked
func (as *GuestService) RegisterGuest(eventID string, nric string, name string) error {
	as.RegisterGuestInvoked = true
	return as.RegisterGuestFn(eventID, nric, name)
}

//RemoveGuest invokes the mock implementation and marks the function as invoked
func (as *GuestService) RemoveGuest(eventID string, nric string) error {
	as.RemoveGuestInvoked = true
	return as.RemoveGuestFn(eventID, nric)
}

//CheckInStats invokes the mock implementation and marks the function as invoked
func (as *GuestService) CheckInStats(eventID string) (checkin.GuestStats, error) {
	as.CheckInInvoked = true
	return as.CheckInStatsFn(eventID)
}
