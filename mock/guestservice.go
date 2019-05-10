package mock

import (
	"checkin"
)

//GuestService represents a mock implementation of the checkin.GuestService interface
type GuestService struct {
	CheckInFn      func(eventID string, nric string) (string, error)
	CheckInInvoked bool

	MarkAbsentFn      func(eventID string, nric string) error
	MarkAbsentInvoked bool

	GuestsFn      func(eventID string, tags []string) ([]string, error)
	GuestsInvoked bool

	GuestsCheckedInFn      func(eventID string, tags []string) ([]string, error)
	GuestsCheckedInInvoked bool

	GuestsNotCheckedInFn      func(eventID string, tags []string) ([]string, error)
	GuestsNotCheckedInInvoked bool

	GuestExistsFn      func(eventID string, nric string) (bool, error)
	GuestExistsInvoked bool

	RegisterGuestFn      func(eventID string, guest checkin.Guest) error
	RegisterGuestInvoked bool

	RegisterGuestsFn      func(eventID string, guest []checkin.Guest) error
	RegisterGuestsInvoked bool

	RemoveGuestFn      func(eventID string, nric string) error
	RemoveGuestInvoked bool

	CheckInStatsFn      func(eventID string, tags []string) (checkin.GuestStats, error)
	CheckInStatsInvoked bool

	TagsFn      func(eventID string, nric string) ([]string, error)
	TagsInvoked bool

	SetTagsFn      func(eventID string, nric string, tags []string) error
	SetTagsInvoked bool
}

//CheckIn invokes the mock implementation and marks the function as invoked
func (as *GuestService) CheckIn(eventID string, nric string) (string, error) {
	as.CheckInInvoked = true
	return as.CheckInFn(eventID, nric)
}

//MarkAbsent invokes the mock implementation and marks the function as invoked
func (as *GuestService) MarkAbsent(eventID string, nric string) error {
	as.MarkAbsentInvoked = true
	return as.MarkAbsentFn(eventID, nric)
}

//Guests invokes the mock implementation and marks the function as invoked
func (as *GuestService) Guests(eventID string, tags []string) ([]string, error) {
	as.GuestsInvoked = true
	return as.GuestsFn(eventID, tags)
}

//GuestsCheckedIn invokes the mock implementation and marks the function as invoked
func (as *GuestService) GuestsCheckedIn(eventID string, tags []string) ([]string, error) {
	as.GuestsCheckedInInvoked = true
	return as.GuestsCheckedInFn(eventID, tags)
}

//GuestsNotCheckedIn invokes the mock implementation and marks the function as invoked
func (as *GuestService) GuestsNotCheckedIn(eventID string, tags []string) ([]string, error) {
	as.GuestsNotCheckedInInvoked = true
	return as.GuestsNotCheckedInFn(eventID, tags)
}

//GuestExists invokes the mock implementation and marks the function as invoked
func (as *GuestService) GuestExists(eventID string, nric string) (bool, error) {
	as.GuestExistsInvoked = true
	return as.GuestExistsFn(eventID, nric)
}

//RegisterGuest invokes the mock implementation and marks the function as invoked
func (as *GuestService) RegisterGuest(eventID string, guest checkin.Guest) error {
	as.RegisterGuestInvoked = true
	return as.RegisterGuestFn(eventID, guest)
}

func (as *GuestService) RegisterGuests(eventID string, guests []checkin.Guest) error {
	as.RegisterGuestsInvoked = true
	return as.RegisterGuestsFn(eventID, guests)
}

//RemoveGuest invokes the mock implementation and marks the function as invoked
func (as *GuestService) RemoveGuest(eventID string, nric string) error {
	as.RemoveGuestInvoked = true
	return as.RemoveGuestFn(eventID, nric)
}

//CheckInStats invokes the mock implementation and marks the function as invoked
func (as *GuestService) CheckInStats(eventID string, tags []string) (checkin.GuestStats, error) {
	as.CheckInInvoked = true
	return as.CheckInStatsFn(eventID, tags)
}

//Tags invokes the mock implementation and marks the function as invoked
func (as *GuestService) Tags(eventID string, nric string) ([]string, error) {
	as.TagsInvoked = true
	return as.TagsFn(eventID, nric)
}

//SetTags invokes the mock implementation and marks the function as invoked
func (as *GuestService) SetTags(eventID string, nric string, tags []string) error {
	as.SetTagsInvoked = true
	return as.SetTagsFn(eventID, nric, tags)
}
