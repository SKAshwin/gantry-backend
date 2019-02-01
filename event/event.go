package event

import (
	"database/sql"
	"errors"
	"log"
	"registration-app/config"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/guregu/null"
)

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

//maps from json:db
var updateSchemaTranslator = map[string]string{"name": "name", "startDateTime": "start", "endDateTime": "end", "lat": "lat", "long": "long", "radius": "radius", "url": "url"}

//GetAll Returns all events, by every user.
func GetAll() ([]Event, error) {
	rows, err := config.DB.Queryx("SELECT ID, name, \"start\", \"end\", lat, long, radius, url, createdAt, updatedAt from event")
	if err != nil {
		return nil, errors.New("Error fetching all events: " + err.Error())
	}
	numEvents, err := getNumberOfEvents()
	if err != nil {
		return nil, errors.New("Error fetching number of events:" + err.Error())
	}

	events, err := scanRowsIntoEvents(rows, numEvents)
	if err == sql.ErrNoRows {
		log.Println("No rows, returned empty event")
		return make([]Event, 0), nil
	} else if err != nil {
		return nil, errors.New("Error scanning rows into events:" + err.Error())
	}

	return events, nil
}

//GetAllBy Given a username as an argument
//Returns an array of all the events hosted by that user
//Will return an empty array (with no error) if that user hosts no events
func GetAllBy(username string) ([]Event, error) {
	rows, err := config.DB.Queryx("SELECT ID, name, \"start\", \"end\", lat, long, radius, url, createdAt, updatedAt from event, hosts where hosts.username = $1 and hosts.eventID = event.ID",
		username)
	if err != nil {
		return nil, errors.New("Error fetching all events for user: " + err.Error())
	}
	numEvents, err := getNumberOfEventsBy(username)
	if err != nil {
		return nil, errors.New("Error fetching number of events for user:" + err.Error())
	}

	events, err := scanRowsIntoEvents(rows, numEvents)
	if err == sql.ErrNoRows {
		log.Println("No rows, returned empty event")
		return make([]Event, 0), nil
	} else if err != nil {
		return nil, errors.New("Error scanning rows into events:" + err.Error())
	}

	return events, nil
}

//Get returns an Event object corresponding to the given eventID
func Get(eventID string) (Event, error) {
	var event Event
	err := config.DB.QueryRowx("SELECT ID, name, \"start\", \"end\", lat, long, radius, url, createdAt, updatedAt from event where ID = $1", eventID).StructScan(&event)
	if err != nil {
		return Event{}, errors.New("Error fetching event details: " + err.Error())
	}
	return event, nil
}

//Delete removes an event (if it exists) from the database
func Delete(eventID string) error {
	_, err := config.DB.Exec("DELETE FROM event where ID = $1", eventID)
	return err
}

//Create creates a new event in the database given its contents
//Also creates a host relationship, given the host's username
func (eventData Event) Create(hostUsername string) error {
	tx, err := config.DB.Beginx()
	if err != nil {
		return errors.New("Error opening transaction:" + err.Error())
	}

	defer func() {
		if r := recover(); r != nil {
			log.Println("event.Create entered panic, recovered to rollback, with error: ", r)
			if rollBackErr := tx.Rollback(); rollBackErr != nil {
				log.Println("Could not rollback: " + rollBackErr.Error())
			}
			panic("Event.Create panicked")
		}
	}()

	_, err = tx.NamedExec("INSERT INTO event(id, name, url,start, \"end\", lat, long, radius) VALUES (:id, :name, :url, :start, :end, :lat, :long, :radius)", eventData)
	if err != nil {
		tx.Rollback()
		return errors.New("Error inserting event data: " + err.Error())
	}

	_, err = tx.Exec("INSERT into hosts(eventID, username) VALUES ($1, $2)", eventData.ID, hostUsername)
	if err != nil {
		tx.Rollback()
		return errors.New("Error creating host relationship: " + err.Error())
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return errors.New("Error committing changes to database: " + err.Error())
	}

	return nil
}

//Update updates a particular event given their username, and a map of attributes to new values
//Returns a boolean flag indicating if the arguments were valid
//Returns a non-nil error if there was an error updating the event
func Update(eventID string, updateFields map[string]string) (bool, error) {
	//check if the update fields are valid
	//this sanitizes the input for later
	if !IsUpdateRequestValid(updateFields) {
		return false, nil
	}

	tx, err := config.DB.Begin()
	if err != nil {
		return false, errors.New("Error opening transaction:" + err.Error())
	}

	defer func() {
		if r := recover(); r != nil {
			log.Println("event.Update entered panic, recovered to rollback, with error: ", r)
			if rollBackErr := tx.Rollback(); rollBackErr != nil {
				log.Println("Could not rollback: " + rollBackErr.Error())
			}
			panic("Event.Update panicked")
		}
	}()

	for attribute, newValue := range updateFields {
		_, err := tx.Exec("UPDATE event SET \""+updateSchemaTranslator[attribute]+"\" = $1 where ID = $2", newValue, eventID)
		if err != nil {
			tx.Rollback()
			return false, errors.New("Error while updating database: " + err.Error())
		}
	}

	_, err = tx.Exec("UPDATE event SET updatedAt = NOW() where ID = $1", eventID)
	if err != nil {
		tx.Rollback()
		return false, errors.New("Error when updating updated field in event: " + err.Error())
	}

	err = tx.Commit()
	if err != nil {
		return false, errors.New("Error committing changes to database: " + err.Error())
	}

	return true, nil
}

//URLExists checks if the given URL is already used
//Returns true if it is already used
//Returns false otherwise
func URLExists(url string) (bool, error) {
	var numURL int
	err := config.DB.QueryRow("SELECT count(*) from event where url = $1", url).Scan(&numURL)
	return numURL == 1, err
}

//CheckIfExists checks if an event exists with that eventID
//Returns a boolean flag indicating if the event exists
//Return a non-nil error if there is an error in querying the database
func CheckIfExists(id string) (bool, error) {
	var num int
	err := config.DB.QueryRow("SELECT count(*) from event where id = $1", id).Scan(&num)
	return num == 1, err
}

//AddHost creates a new host relationship between a user and an event
func (eventData Event) AddHost(username string) error {
	_, err := config.DB.Exec("INSERT INTO hosts(eventID, username) VALUES ($1, $2)", eventData.ID, username)
	return err
}

//CheckHost returns true if that user is a host of the given event
func CheckHost(username string, eventID string) (bool, error) {
	var numHosts int
	err := config.DB.QueryRow("SELECT count(*) from hosts where hosts.eventID = $1 and hosts.username = $2", eventID, username).Scan(&numHosts)
	if err != nil {
		return false, errors.New("Error checking if host relationship exists: " + err.Error())
	}
	return numHosts == 1, nil
}

//IsUpdateRequestValid checks if the fields provided in an update request
//are allowed. Only specific columns are allowed to be updated
func IsUpdateRequestValid(updateFields map[string]string) bool {
	for attribute := range updateFields {
		if _, exist := updateSchemaTranslator[attribute]; !exist {
			return false
		}
	}
	return true
}

func getNumberOfEvents() (int, error) {
	var numEvents int
	err := config.DB.QueryRow("SELECT count(*) from event").Scan(&numEvents)

	if err != nil {
		return 0, errors.New("Cannot fetch event count: " + err.Error())
	}

	return numEvents, nil
}

func getNumberOfEventsBy(username string) (int, error) {
	var numEvents int
	err := config.DB.QueryRow("SELECT count(*) from event, hosts where hosts.username = $1 and hosts.eventID = event.ID", username).Scan(&numEvents)

	if err != nil {
		return 0, errors.New("Cannot fetch event count for user: " + err.Error())
	}

	return numEvents, nil
}

func scanRowsIntoEvents(rows *sqlx.Rows, numRows int) ([]Event, error) {
	events := make([]Event, numRows)

	index := 0
	for thereAreMore := rows.Next(); thereAreMore; thereAreMore = rows.Next() {
		var event Event
		err := rows.StructScan(&event)
		if err != nil {
			return nil, errors.New("Could not extract event: " + err.Error())
		}
		events[index] = event
		index++
	}

	return events, nil
}
