package mock

import (
	"checkin"
)

//EventService represents a mock implementation of the checkin.EventService interface
type EventService struct {
	EventFn      func(ID string) (checkin.Event, error)
	EventInvoked bool

	EventsByFn      func(username string) ([]checkin.Event, error)
	EventsByInvoked bool

	EventsFn      func() ([]checkin.Event, error)
	EventsInvoked bool

	CreateEventFn      func(e checkin.Event, hostUsername string) error
	CreateEventInvoked bool

	DeleteEventFn      func(ID string) error
	DeleteEventInvoked bool

	UpdateEventFn      func(event checkin.Event) error
	UpdateEventInvoked bool

	URLExistsFn      func(url string) (bool, error)
	URLExistsInvoked bool

	CheckIfExistsFn      func(id string) (bool, error)
	CheckIfExistsInvoked bool

	AddHostFn      func(eventID string, username string) error
	AddHostInvoked bool

	CheckHostFn      func(username string, eventID string) (bool, error)
	CheckHostInvoked bool

	FeedbackFormsFn func(ID string) ([]checkin.FeedbackForm, error)
	FeedbackFormsInvoked bool

	SubmitFeedbackFn func(ID string, ff checkin.FeedbackForm) error
	SubmitFeedbackInvoked bool
}

//Event invokes the mock implementation and marks the function as invoked
func (es *EventService) Event(ID string) (checkin.Event, error) {
	es.EventInvoked = true
	return es.EventFn(ID)
}

//EventsBy invokes the mock implementation and marks the function as invoked
func (es *EventService) EventsBy(username string) ([]checkin.Event, error) {
	es.EventsByInvoked = true
	return es.EventsByFn(username)
}

//Events invokes the mock implementation and marks the function as invoked
func (es *EventService) Events() ([]checkin.Event, error) {
	es.EventsInvoked = true
	return es.EventsFn()
}

//CreateEvent invokes the mock implementation and marks the function as invoked
func (es *EventService) CreateEvent(e checkin.Event, hostUsername string) error {
	es.CreateEventInvoked = true
	return es.CreateEventFn(e, hostUsername)
}

//DeleteEvent invokes the mock implementation and marks the function as invoked
func (es *EventService) DeleteEvent(ID string) error {
	es.DeleteEventInvoked = true
	return es.DeleteEventFn(ID)
}

//UpdateEvent invokes the mock implementation and marks the function as invoked
func (es *EventService) UpdateEvent(event checkin.Event) error {
	es.UpdateEventInvoked = true
	return es.UpdateEventFn(event)
}

//URLExists invokes the mock implementation and marks the function as invoked
func (es *EventService) URLExists(url string) (bool, error) {
	es.URLExistsInvoked = true
	return es.URLExistsFn(url)
}

//CheckIfExists invokes the mock implementation and marks the function as invoked
func (es *EventService) CheckIfExists(id string) (bool, error) {
	es.CheckIfExistsInvoked = true
	return es.CheckIfExistsFn(id)
}

//AddHost invokes the mock implementation and marks the function as invoked
func (es *EventService) AddHost(eventID string, username string) error {
	es.AddHostInvoked = true
	return es.AddHostFn(eventID, username)
}

//CheckHost invokes the mock implementation and marks the function as invoked
func (es *EventService) CheckHost(username string, eventID string) (bool, error) {
	es.CheckHostInvoked = true
	return es.CheckHostFn(username, eventID)
}

func (es *EventService) FeedbackForms(ID string) ([]checkin.FeedbackForm, error) {
	es.FeedbackFormsInvoked = true
	return es.FeedbackFormsFn(ID)
}

func (es *EventService) SubmitFeedback(ID string, ff checkin.FeedbackForm) error {
	es.SubmitFeedbackInvoked = true
	return es.SubmitFeedbackFn(ID, ff)
}
