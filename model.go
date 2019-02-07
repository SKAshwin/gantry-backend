package checkin

import (
	"time"

	"github.com/guregu/null"
)

//User represents a user of the website creator - i.e., hosts of events, who want to run an event
//check-in page.
//This is distinguished from guests, who attend events
type User struct {
	Username          string    `json:"username,omitempty" db:"username"`
	PasswordPlaintext string    `json:"password,omitempty"`
	PasswordHash      string    `json:"-" db:"passwordHash"` //always omitted upon JSON marshalling
	Name              string    `json:"name,omitempty" db:"name"`
	CreatedAt         time.Time `json:"createdAt,omitempty"`
	UpdatedAt         time.Time `json:"updatedAt,omitempty"`
	LastLoggedIn      null.Time `json:"lastLoggedIn,omitempty"`
}

//UserService An interface for functions that modify/fetch user data in the database
type UserService interface {
	User(username string) (User, error)
	Users() ([]User, error)
	CreateUser(u User) error
	DeleteUser(username string) error
	UpdateUser(username string, updateFields map[string]string) error
	CheckIfExists(username string)
	UpdateLastLoggedIn(username string)
	AddHost(eventID string, username string) error
	CheckHost(username string, eventID string) (bool, error)
}

//Event represents an event which will have an associated website
type Event struct {
	ID        string     `json:"eventId" db:"id"`
	Name      string     `json:"name" db:"name"`
	Start     null.Time  `json:"startDateTime" db:"start"`
	End       null.Time  `json:"endDateTime" db:"end"`
	Lat       null.Float `json:"lat" db:"lat"`
	Long      null.Float `json:"long" db:"long"`
	Radius    null.Float `json:"radius" db:"radius"` //in km
	URL       string     `json:"url" db:"url"`
	UpdatedAt time.Time  `json:"updatedAt" db:"updatedat"`
	CreatedAt time.Time  `json:"createdAt" db:"createdat"`
}

//EventService An interface for functions that modify/fetch event data in the database
type EventService interface {
	Event(ID string) (Event, error)
	EventsBy(username string) ([]Event, error)
	Events() ([]Event, error)
	CreateEvent(e Event, hostUsername string) error
	DeleteEvent(ID string) error
	UpdateEvent(ID string, updateFields map[string]string) (bool, error)
	URLExists(url string) (bool, error)
}

//HashMethod An interface allowing you to hash a string, and confirm if a string matches a given hash
type HashMethod interface {
	HashAndSalt(pwd string) (string, error)
	CompareHashAndPassword(hash string, pwd string) bool
}
