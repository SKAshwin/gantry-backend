package checkin

import (
	"time"

	"github.com/guregu/null"
)

//User represents a user of the website creator - i.e., hosts of events, who want to run an event
//check-in page.
//This is distinguished from guests, who attend events
type User struct {
	Username          string  `json:"username,omitempty" db:"username"`
	PasswordPlaintext *string `json:"password,omitempty"` //using a string pointer so this can be nil
	//null.String doesn't work will omitempty, but a nil string pointer will be omitted
	PasswordHash string    `json:"-" db:"passwordhash"` //always omitted upon JSON marshalling
	Name         string    `json:"name,omitempty" db:"name"`
	CreatedAt    time.Time `json:"createdAt,omitempty"`
	UpdatedAt    time.Time `json:"updatedAt,omitempty"`
	LastLoggedIn null.Time `json:"lastLoggedIn,omitempty"`
}

//UserService An interface for functions that modify/fetch user data in the database
type UserService interface {
	User(username string) (User, error)
	Users() ([]User, error)
	CreateUser(u User) error
	DeleteUser(username string) error
	UpdateUser(originalUsername string, newUser User) error
	CheckIfExists(username string) (bool, error)
	UpdateLastLoggedIn(username string) error
}

//Event represents an event which will have an associated website
type Event struct {
	ID        string     `json:"eventId" db:"id"`
	Name      string     `json:"name" db:"name"`
	Release   null.Time  `json:"releaseDateTime" db:"release"`
	Start     null.Time  `json:"startDateTime" db:"start"`
	End       null.Time  `json:"endDateTime" db:"end"`
	Lat       null.Float `json:"lat" db:"lat"`
	Long      null.Float `json:"long" db:"long"`
	Radius    null.Float `json:"radius" db:"radius"` //in km
	URL       string     `json:"url" db:"url"`
	UpdatedAt time.Time  `json:"updatedAt" db:"updatedat"`
	CreatedAt time.Time  `json:"createdAt" db:"createdat"`
}

//FeedbackFormItem represents a question/answer pair in a feedback form
type FeedbackFormItem struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

//FeedbackForm is a collection of questions (and their answers) and the NRIC (either a hash or literal) of the submitter
//which conceptually can be an empty string for anonymous submissions
type FeedbackForm struct {
	ID         string             `json:"id" db:"id"`
	Name       string             `json:"name" db:"name"`
	Survey     []FeedbackFormItem `json:"survey" db:"survey"`
	SubmitTime time.Time          `json:"submitTime" db:"submittime"`
}

//Released returns true if the current time in Singapore is beyond
//the release time in UTC
func (event *Event) Released() bool {
	now := time.Now().UTC()

	if !event.Release.Valid {
		//if no release time set, return true
		return true
	}
	return event.Release.Time.UTC().Before(now)
}

//EventService An interface for functions that modify/fetch event data in the database
type EventService interface {
	Event(ID string) (Event, error)
	EventsBy(username string) ([]Event, error)
	Events() ([]Event, error)
	CreateEvent(e Event, hostUsername string) error
	DeleteEvent(ID string) error
	UpdateEvent(e Event) error
	URLExists(url string) (bool, error)
	CheckIfExists(id string) (bool, error)
	AddHost(eventID string, username string) error
	CheckHost(username string, eventID string) (bool, error)
	FeedbackForms(ID string) ([]FeedbackForm, error)
	SubmitFeedback(ID string, ff FeedbackForm) error
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
	Name string   `json:"name,omitempty"`
	NRIC string   `json:"nric,omitempty" db:"nrichash"`
	Tags []string `json:"tags,omitempty" db:"tags"`
}

//IsEmpty checks if this is an empty Guest struct
//i.e. "" name, "" nric and nil for Tags (NOT an empty array)
func (g *Guest) IsEmpty() bool {
	return g.Name == "" && g.NRIC == "" && g.Tags == nil
}

//GuestService is for checking in guests at a specific event
type GuestService interface {
	CheckIn(eventID string, nric string) (string, error)
	MarkAbsent(eventID string, nric string) error
	Guests(eventID string, tags []string) ([]string, error)
	GuestsCheckedIn(eventID string, tags []string) ([]string, error)
	GuestsNotCheckedIn(eventID string, tags []string) ([]string, error)
	GuestExists(eventID string, nric string) (bool, error)
	RegisterGuest(eventID string, guest Guest) error
	RegisterGuests(eventID string, guests []Guest) error
	Tags(eventID string, nric string) ([]string, error)
	SetTags(eventID string, nric string, tags []string) error
	RemoveGuest(eventID string, nric string) error
	CheckInStats(eventID string, tags []string) (GuestStats, error)
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
