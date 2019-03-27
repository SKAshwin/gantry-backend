package postgres

import (
	"checkin"
	"database/sql"
	"errors"
	"log"

	"github.com/jmoiron/sqlx"
)

//EventService Implementation of an event service
//Needs to be supplied with a database connection
type EventService struct {
	DB *sqlx.DB
}

//Event Fetches the details of an event, given its ID
func (es *EventService) Event(eventID string) (checkin.Event, error) {
	var event checkin.Event
	err := es.DB.QueryRowx(
		"SELECT * from event where ID = $1",
		eventID).StructScan(&event)
	if err != nil {
		return checkin.Event{}, errors.New("Error fetching event details: " + err.Error())
	}
	return event, nil
}

//Events Returns every event, by every user
func (es *EventService) Events() ([]checkin.Event, error) {
	rows, err := es.DB.Queryx(
		"SELECT * from event")
	if err != nil {
		return nil, errors.New("Error fetching all events: " + err.Error())
	}
	numEvents, err := es.getNumberOfEvents()
	if err != nil {
		return nil, errors.New("Error fetching number of events:" + err.Error())
	}

	events, err := es.scanRowsIntoEvents(rows, numEvents)
	if err == sql.ErrNoRows {
		log.Println("No rows, returned empty event")
		return make([]checkin.Event, 0), nil
	} else if err != nil {
		return nil, errors.New("Error scanning rows into events:" + err.Error())
	}

	return events, nil
}

//EventsBy Given a username as an argument
//Returns an array of all the events hosted by that user
//Will return an empty array (with no error) if that user hosts no events
func (es *EventService) EventsBy(username string) ([]checkin.Event, error) {
	rows, err := es.DB.Queryx("SELECT ID, name, \"start\", \"end\", release, lat, long, radius, url, createdAt, updatedAt from event, hosts where hosts.username = $1 and hosts.eventID = event.ID",
		username)
	if err != nil {
		return nil, errors.New("Error fetching all events for user: " + err.Error())
	}
	numEvents, err := es.getNumberOfEventsBy(username)
	if err != nil {
		return nil, errors.New("Error fetching number of events for user:" + err.Error())
	}

	events, err := es.scanRowsIntoEvents(rows, numEvents)
	if err == sql.ErrNoRows {
		log.Println("No rows, returned empty event")
		return make([]checkin.Event, 0), nil
	} else if err != nil {
		return nil, errors.New("Error scanning rows into events:" + err.Error())
	}

	return events, nil
}

//CreateEvent creates a new event in the database given its contents
//Also creates a host relationship, given the host's username
func (es *EventService) CreateEvent(e checkin.Event, hostUsername string) error {
	tx, err := es.DB.Beginx()
	if err != nil {
		return errors.New("Error opening transaction:" + err.Error())
	}

	defer func() {
		if r := recover(); r != nil {
			log.Println("CreateEvent entered panic, recovered to rollback, with error: ", r)
			if rollBackErr := tx.Rollback(); rollBackErr != nil {
				log.Println("Could not rollback: " + rollBackErr.Error())
			}
			panic("CreateEvent panicked")
		}
	}()

	_, err = tx.NamedExec("INSERT INTO event(id, name, url,start, \"end\", lat, long, radius) VALUES (:id, :name, :url, :start, :end, :lat, :long, :radius)", e)
	if err != nil {
		tx.Rollback()
		return errors.New("Error inserting event data: " + err.Error())
	}

	_, err = tx.Exec("INSERT into hosts(eventID, username) VALUES ($1, $2)", e.ID, hostUsername)
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

//UpdateEvent updates a particular event given an event object encapsulating
//ALL THE NEW FIELDS of the object
//It will use the eventID as the key to know which row in the DB to update
//So note that eventID cannot be mutated
//All columns in the database will be set to the fields of the event object
//Except for the createdAt and updatedAt fields, which are not editable
func (es *EventService) UpdateEvent(event checkin.Event) error {
	tx, err := es.DB.Beginx()
	if err != nil {
		return errors.New("Error opening transaction:" + err.Error())
	}

	defer func() {
		if r := recover(); r != nil {
			log.Println("UpdateEvent entered panic, recovered to rollback, with error: ", r)
			if rollBackErr := tx.Rollback(); rollBackErr != nil {
				log.Println("Could not rollback: " + rollBackErr.Error())
			}
			panic("UpdateEvent panicked")
		}
	}()

	_, err = tx.NamedExec("UPDATE event SET name = :name, release = :release, \"start\" = :start, "+
		"\"end\" = :end, lat = :lat, long= :long, radius = :radius, url = :url where id = :id", &event)
	if err != nil {
		tx.Rollback()
		return errors.New("Error when updating event: " + err.Error())
	}

	_, err = tx.Exec("UPDATE event SET updatedAt = NOW() where ID = $1", event.ID)
	if err != nil {
		tx.Rollback()
		return errors.New("Error when updating updated field in event: " + err.Error())
	}

	err = tx.Commit()
	if err != nil {
		return errors.New("Error committing changes to database: " + err.Error())
	}

	return nil
}

//DeleteEvent removes an event (if it exists) from the database
func (es *EventService) DeleteEvent(eventID string) error {
	_, err := es.DB.Exec("DELETE FROM event where ID = $1", eventID)
	return err
}

//URLExists checks if the given URL is already used
//Returns true if it is already used
//Returns false otherwise
func (es *EventService) URLExists(url string) (bool, error) {
	var numURL int
	err := es.DB.QueryRow("SELECT count(*) from event where url = $1", url).Scan(&numURL)
	return numURL == 1, err
}

//CheckIfExists checks if an event exists with that eventID
//Returns a boolean flag indicating if the event exists
//Return a non-nil error if there is an error in querying the database
func (es *EventService) CheckIfExists(id string) (bool, error) {
	var num int
	err := es.DB.QueryRow("SELECT count(*) from event where id = $1", id).Scan(&num)
	return num == 1, err
}

//AddHost creates a new host relationship between a user and an event
func (es *EventService) AddHost(eventID string, username string) error {
	_, err := es.DB.Exec("INSERT INTO hosts(eventID, username) VALUES ($1, $2)", eventID, username)
	return err
}

//CheckHost returns true if that user is a host of the given event
func (es *EventService) CheckHost(username string, eventID string) (bool, error) {
	var numHosts int
	err := es.DB.QueryRow("SELECT count(*) from hosts where hosts.eventID = $1 and hosts.username = $2",
		eventID, username).Scan(&numHosts)
	if err != nil {
		return false, errors.New("Error checking if host relationship exists: " + err.Error())
	}
	return numHosts == 1, nil
}

func (es *EventService) getNumberOfEvents() (int, error) {
	var numEvents int
	err := es.DB.QueryRow("SELECT count(*) from event").Scan(&numEvents)

	if err != nil {
		return 0, errors.New("Cannot fetch event count: " + err.Error())
	}

	return numEvents, nil
}

func (es *EventService) getNumberOfEventsBy(username string) (int, error) {
	var numEvents int
	err := es.DB.QueryRow("SELECT count(*) from event, hosts where hosts.username = $1 and hosts.eventID = event.ID", username).Scan(&numEvents)

	if err != nil {
		return 0, errors.New("Cannot fetch event count for user: " + err.Error())
	}

	return numEvents, nil
}

func (es *EventService) scanRowsIntoEvents(rows *sqlx.Rows, numRows int) ([]checkin.Event, error) {
	events := make([]checkin.Event, numRows)

	index := 0
	for thereAreMore := rows.Next(); thereAreMore; thereAreMore = rows.Next() {
		var event checkin.Event
		err := rows.StructScan(&event)
		if err != nil {
			return nil, errors.New("Could not extract event: " + err.Error())
		}
		events[index] = event
		index++
	}

	return events, nil
}
