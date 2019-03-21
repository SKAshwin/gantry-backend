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
	UpdateUser(username string, updateFields map[string]string) (bool, error)
	CheckIfExists(username string) (bool, error)
	UpdateLastLoggedIn(username string) error
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
	CheckIfExists(id string) (bool, error)
	AddHost(eventID string, username string) error
	CheckHost(username string, eventID string) (bool, error)
}

//HashMethod An interface allowing you to hash a string, and confirm if a string matches a given hash
type HashMethod interface {
	HashAndSalt(pwd string) (string, error)
	CompareHashAndPassword(hash string, pwd string) bool
}

//AuthenticationService An interface for functions to perform authentication
type AuthenticationService interface {
	Authenticate(username string, pwdPlaintext string, isAdmin bool) (bool, error)
}

//GuestStats are statistics relating to attendance of the event
type GuestStats struct {
	TotalGuests      int     `json:"total"`
	CheckedIn        int     `json:"checkedIn"`
	PercentCheckedIn float64 `json:"percentCheckedIn"`
}

//Guest is all the information related to a particular guest
type Guest struct {
	Name string `json:"name,omitempty"`
	NRIC string `json:"nric,omitempty" db:"nrichash"`
}

//GuestService is for checking in guests at a specific event
type GuestService interface {
	CheckIn(eventID string, nric string) (string, error)
	Guests(eventID string) ([]string, error)
	GuestsCheckedIn(eventID string) ([]string, error)
	GuestsNotCheckedIn(eventID string) ([]string, error)
	GuestExists(eventID string, nric string) (bool, error)
	RegisterGuest(eventID string, nric string, name string) error
	RemoveGuest(eventID string, nric string) error
	CheckInStats(eventID string) (GuestStats, error)
}

//AuthorizationInfo stores critical information about a particular request's authorizations
//It provides the username and admin status of the user
type AuthorizationInfo struct {
	Username string
	IsAdmin  bool
}

//QRGenerator generates a QR code given a message and size.
type QRGenerator interface {
	//Encode encodes a message in a QR code with minimum padding
	//and a given size
	//The QR Code can be a byte array of any image type
	//up to caller to check
	Encode(msg string, size int) ([]byte, error)
}
