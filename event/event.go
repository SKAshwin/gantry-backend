package event

import (
	"database/sql"
	"errors"
	"log"
	"registration-app/config"

	"github.com/jmoiron/sqlx"

	"github.com/guregu/null"
)

//Event represents an event which will have an associated website
type Event struct {
	ID     string     `json:"id"`
	Name   string     `json:"name"`
	Start  null.Time  `json:"start"`
	End    null.Time  `json:"end"`
	Lat    null.Float `json:"lat"`
	Long   null.Float `json:"long"`
	Radius null.Float `json:"radius"` //in km
}

//GetAll Given a username as an argument
//Returns an array of all the events hosted by that user
//Will return an empty array (with no error) if that user hosts no events
func GetAll(username string) ([]Event, error) {
	rows, err := config.DB.Queryx("SELECT ID, name, \"start\", \"end\", lat, long, radius from event, hosts where hosts.username = $1 and hosts.eventID = event.ID", username)
	if err != nil {
		return nil, errors.New("Error fetching all events for user: " + err.Error())
	}
	numEvents, err := getNumberOfEvents(username)
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
	err := config.DB.QueryRowx("SELECT ID, name, \"start\", \"end\", lat, long, radius from event where ID = $1", eventID).StructScan(&event)
	if err != nil {
		return Event{}, errors.New("Error fetching event details " + err.Error())
	}
	return event, nil
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

func getNumberOfEvents(username string) (int, error) {
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
