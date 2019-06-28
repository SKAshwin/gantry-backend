package postgres

import (
	"checkin"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/jmoiron/sqlx"
)

//EventService Implementation of an event service
//Needs to be supplied with a database connection
type EventService struct {
	DB *sqlx.DB
}

//rawEvent has all the fields of an event, but reads the timetags as a JSON byte array
//corresponds to the DB representation of an Event (which does not have a TimeTag type, and therefore
//store TimeTags as JSON
type rawEvent struct {
	checkin.Event
	TimetagJSON []byte `db:"timetags"`
}

//Event Fetches the details of an event, given its ID
func (es *EventService) Event(eventID string) (checkin.Event, error) {
	var rEvent rawEvent
	err := es.DB.QueryRowx(
		"SELECT * from event where ID = $1",
		eventID).StructScan(&rEvent)
	if err != nil {
		return checkin.Event{}, errors.New("Error fetching event details: " + err.Error())
	}
	event, err := es.unmarshalEvent(rEvent)
	if err != nil {
		return checkin.Event{}, errors.New("Error unmarshalling timetag in event: " + err.Error())
	}
	return event, nil
}

//EventByURL Fetches the details of an event, given its URL
//Obviously, this cannot fetch events which have null event IDs
func (es *EventService) EventByURL(url string) (checkin.Event, error) {
	var rEvent rawEvent
	err := es.DB.QueryRowx(
		"SELECT * from event where url = $1",
		url).StructScan(&rEvent)
	if err != nil {
		return checkin.Event{}, errors.New("Error fetching event details: " + err.Error())
	}
	event, err := es.unmarshalEvent(rEvent)
	if err != nil {
		return checkin.Event{}, errors.New("Error unmarshalling timetag in event: " + err.Error())
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
	defer rows.Close()
	numEvents, err := es.getNumberOfEvents()
	if err != nil {
		return nil, errors.New("Error fetching number of events:" + err.Error())
	}

	events, err := es.scanRowsIntoEvents(rows, numEvents)
	if err == sql.ErrNoRows {
		return make([]checkin.Event, 0), nil
	} else if err != nil {
		return nil, errors.New("Error scanning rows into events:" + err.Error())
	}

	return events, nil
}

//EventsBy Given a username as an argument
//Returns an array of all the events hosted by that user
//Will return an empty array (with no error) if that user hosts no events
//If the user does not exist, will return an empty array (with no error)
//as it is not the job of an EventService to perform user validation
//Error only if there are issues fetching events from the database or scanning them
//into structs
func (es *EventService) EventsBy(username string) ([]checkin.Event, error) {
	//need to list out columns instead of * as hosts is used in the query
	rows, err := es.DB.Queryx("SELECT id, name, \"start\", \"end\", lat, long, radius, url, updatedat, createdat, timetags, website from event, hosts where hosts.username = $1 and hosts.eventID = event.ID",
		username)
	if err != nil {
		return nil, errors.New("Error fetching all events for user: " + err.Error())
	}
	defer rows.Close()
	numEvents, err := es.getNumberOfEventsBy(username)
	if err != nil {
		return nil, errors.New("Error fetching number of events for user:" + err.Error())
	}

	events, err := es.scanRowsIntoEvents(rows, numEvents)
	if err == sql.ErrNoRows {
		return make([]checkin.Event, 0), nil
	} else if err != nil {
		return nil, errors.New("Error scanning rows into events:" + err.Error())
	}

	return events, nil
}

//CreateEvent creates a new event in the database given its contents
//Also creates a host relationship, given the host's username
//e.ID needs to be a valid UUID
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

	rawEvent := es.marshalEvent(e)

	_, err = tx.NamedExec("INSERT INTO event(id, name, url, start, \"end\", timetags, lat, long, radius, website) VALUES (:id, :name, :url, :start, :end, :timetags,:lat, :long, :radius, :website)", rawEvent)
	if err != nil {
		tx.Rollback()
		return errors.New("Error inserting event data: " + err.Error())
	}

	_, err = tx.Exec("INSERT into hosts(eventID, username) VALUES ($1, $2)", rawEvent.ID, hostUsername)
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
//Except for the ID, createdAt and updatedAt fields, which are not editable
//Returns an error if error in executing update, or no event with that ID exists
func (es *EventService) UpdateEvent(event checkin.Event) error {
	rawEvent := es.marshalEvent(event)
	res, err := es.DB.NamedExec("UPDATE event SET name = :name, timetags = :timetags, \"start\" = :start, "+
		"\"end\" = :end, lat = :lat, long= :long, radius = :radius, url = :url, website = :website, updatedAt = (NOW() at time zone 'utc') where id = :id",
		&rawEvent)
	if err != nil {
		return errors.New("Error when updating event: " + err.Error())
	}
	if rows, err := res.RowsAffected(); err != nil {
		return errors.New("Error checking if rows were affected: " + err.Error())
	} else if rows == 0 {
		return errors.New("No event exists with that UUID")
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
	if _, err := uuid.Parse(id); err != nil {
		//if id is not even a valid UUID, event does not exist
		//this check is necessary, because otherwise an SQL error will be thrown in the next statement
		//so this function will return an error instead of false for the event not existing
		return false, nil
	}
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

//SubmitFeedback adds a feedback form to the database
//Returns error if the feedback form has a nil or empty survey (no questions in it)
func (es *EventService) SubmitFeedback(eventID string, ff checkin.FeedbackForm) error {
	if ff.Survey == nil || len(ff.Survey) == 0 {
		return errors.New("Cannot submit nil or empty survey")
	}

	j, err := json.Marshal(ff.Survey)
	if err != nil {
		return errors.New("Error marshalling form items into JSON: " + err.Error())
	}

	_, err = es.DB.Exec("INSERT INTO form(name, survey, eventID) VALUES($1, $2, $3)", ff.Name, j, eventID)
	if err != nil {
		return errors.New("Error inserting new form: " + err.Error())
	}
	return nil
}

//FeedbackForms return an array of forms that have been submitted for a particular event
//Returns an error if the event does not exist (or if checking existence caused failure)
//or if there is an error fetching/parsing the forms
func (es *EventService) FeedbackForms(eventID string) ([]checkin.FeedbackForm, error) {
	if exists, err := es.CheckIfExists(eventID); !exists {
		return nil, errors.New("No such event exists to get feedback forms for")
	} else if err != nil {
		return nil, errors.New("Error checking if event exists: " + err.Error())
	}

	rows, err := es.DB.Queryx("SELECT ID, name, survey, submitTime from form where eventID = $1", eventID)
	if err != nil {
		return nil, errors.New("Error fetching all forms for event: " + err.Error())
	}
	defer rows.Close()
	numForms, err := es.getNumberOfFeedbackForms(eventID)
	if err != nil {
		return nil, errors.New("Error fetching number of forms for event:" + err.Error())
	}

	forms, err := es.scanRowsIntoForms(rows, numForms)
	if err != nil {
		return nil, errors.New("Error scanning rows into forms:" + err.Error())
	}

	return forms, nil
}

func (es *EventService) scanRowsIntoForms(rows *sqlx.Rows, numRows int) ([]checkin.FeedbackForm, error) {
	forms := make([]checkin.FeedbackForm, numRows)

	index := 0
	for thereAreMore := rows.Next(); thereAreMore; thereAreMore = rows.Next() {
		var form checkin.FeedbackForm
		var surveyJSON []byte
		err := rows.Scan(&form.ID, &form.Name, &surveyJSON, &form.SubmitTime)
		if err != nil {
			return nil, errors.New("Could not extract form: " + err.Error())
		}
		err = json.Unmarshal(surveyJSON, &form.Survey)
		if err != nil {
			return nil, errors.New("Could not unmarshal form item data from JSON: " + err.Error())
		}
		form.SubmitTime = form.SubmitTime.UTC() //make sure all times are in UTC (postgres has them in a +0:00 timezone)
		forms[index] = form
		index++
	}

	return forms, nil
}

func (es *EventService) getNumberOfFeedbackForms(eventID string) (int, error) {
	var numForms int
	err := es.DB.QueryRow("SELECT count(*) from form where eventID = $1", eventID).Scan(&numForms)

	if err != nil {
		return 0, errors.New("Cannot fetch form count for event: " + err.Error())
	}

	return numForms, nil
}

func (es *EventService) scanRowsIntoEvents(rows *sqlx.Rows, numRows int) ([]checkin.Event, error) {
	events := make([]checkin.Event, numRows)

	index := 0
	for thereAreMore := rows.Next(); thereAreMore; thereAreMore = rows.Next() {
		var rawEvent rawEvent
		err := rows.StructScan(&rawEvent)
		if err != nil {
			return nil, errors.New("Could not extract event: " + err.Error())
		}
		event, err := es.unmarshalEvent(rawEvent)
		if err != nil {
			return nil, errors.New("Could not unmarshal time tag data from JSON: " + err.Error())
		}
		events[index] = event
		index++
	}

	return events, nil
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

//Converts an event to its raw (DB-storable) form, by marshalling
//its timetag array into a JSON
//also makes sure that all times are in UTC, as postgres will lose all timezone information
func (es *EventService) marshalEvent(event checkin.Event) rawEvent {
	if event.TimeTags == nil {
		event.TimeTags = make(map[string]time.Time, 0)
	}
	for key, val := range event.TimeTags {
		delete(event.TimeTags, key)
		event.TimeTags[strings.ToLower(key)] = val.In(time.UTC)
	}
	event.Start.Time = event.Start.Time.In(time.UTC)
	event.End.Time = event.End.Time.In(time.UTC)
	event.CreatedAt = event.CreatedAt.In(time.UTC)
	event.UpdatedAt = event.UpdatedAt.In(time.UTC)
	timetags, _ := json.Marshal(event.TimeTags)
	return rawEvent{Event: event, TimetagJSON: timetags}
}

//unmarshals an event from its raw (DB-storable) form, by unmarshalling
//its timetag JSON into a timetag array
func (es *EventService) unmarshalEvent(re rawEvent) (checkin.Event, error) {
	event := re.Event //create an actual event object from the rawEvent, parse the timetagJSON into TimeTags
	err := json.Unmarshal(re.TimetagJSON, &event.TimeTags)

	for label, tt := range event.TimeTags {
		event.TimeTags[label] = tt.In(time.UTC)
	}
	event.Start.Time = event.Start.Time.In(time.UTC)
	event.End.Time = event.End.Time.In(time.UTC)
	event.CreatedAt = event.CreatedAt.In(time.UTC)
	event.UpdatedAt = event.UpdatedAt.In(time.UTC)

	return event, err
}
