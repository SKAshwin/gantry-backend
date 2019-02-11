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

	UpdateEventFn      func(ID string, updateFields map[string]string) (bool, error)
	UpdateEventInvoked bool

	URLExistsFn      func(url string) (bool, error)
	URLExistsInvoked bool

	CheckIfExistsFn      func(id string) (bool, error)
	CheckIfExistsInvoked bool

	AddHostFn      func(eventID string, username string) error
	AddHostInvoked bool

	CheckHostFn      func(username string, eventID string) (bool, error)
	CheckHostInvoked bool
}

func (es *EventService) Event(ID string) (checkin.Event, error) {

}
func (es *EventService) EventsBy(username string) ([]checkin.Event, error) {

}
func (es *EventService) Events() ([]checkin.Event, error) {

}
func (es *EventService) CreateEvent(e checkin.Event, hostUsername string) error {

}
func (es *EventService) DeleteEvent(ID string) error {

}
func (es *EventService) UpdateEvent(ID string, updateFields map[string]string) (bool, error) {

}
func (es *EventService) URLExists(url string) (bool, error) {

}
func (es *EventService) CheckIfExists(id string) (bool, error) {

}
func (es *EventService) AddHost(eventID string, username string) error {

}
func (es *EventService) CheckHost(username string, eventID string) (bool, error) {

}
