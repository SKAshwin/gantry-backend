package postgres

//UNTESTED
import (
	"checkin"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

//GuestService an implementation of checkin.GuestService using postgres
//Needs a HashMethod as all NRICs are stored internally as hashes for
//security purposes
type GuestService struct {
	DB *sqlx.DB
	HM checkin.HashMethod
}

//CheckIn marks a guest (indicated by the last 5 digits of the nric)
//of a particular event as having attended the event
//Returns the name of the guest who was checked in
//Will return an error if said guest does not exist, or event with
//that ID does not exist
//Will not throw an error if the guest is already checked in
//If any error occurs, check in status of the guest will not be edited
func (gs *GuestService) CheckIn(eventID string, nric string) (string, error) {
	nricHash, err := gs.HM.HashAndSalt(nric)
	if err != nil {
		return "", errors.New("Error hashing NRIC: " + err.Error())
	}

	tx, err := gs.DB.Begin()
	if err != nil {
		return "", errors.New("Error starting transaction: " + err.Error())
	}

	_, err = tx.Exec("UPDATE guest SET checkedIn = TRUE, checkInTime = NOW() WHERE eventID = $1 and nricHash = $2",
		eventID, nricHash)
	if err != nil {
		tx.Rollback()
		return "", errors.New("Error updating check in status: " + err.Error())
	}

	var name string
	err = tx.QueryRow("SELECT name FROM guest WHERE eventID = $1 and nricHash = $2",
		eventID, nricHash).Scan(&name)
	if err != nil {
		tx.Rollback()
		return "", errors.New("Error updating fetching name: " + err.Error())
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return "", errors.New("Error committing changes to the database: " + err.Error())
	}

	return name, nil
}

//Guests returns an array of names of the guests who are registered/signed up for
//an event given by the eventID
func (gs *GuestService) Guests(eventID string) ([]string, error) {
	rows, err := gs.DB.Query("SELECT name from guest where eventID = $1", eventID)
	if err != nil {
		return nil, errors.New("Cannot fetch guest names: " + err.Error())
	}
	defer rows.Close() //make sure this is after checking for an error, or this will be a nil pointer dereference
	numGuests, err := gs.getNumberOfGuests(eventID)
	if err != nil {
		return nil, errors.New("Cannot fetch number of guests: " + err.Error())
	}

	return gs.scanRowsIntoNames(rows, numGuests)
}

//GuestsCheckedIn return an array of names of the guests who have checked in
//to the event given by the eventID
func (gs *GuestService) GuestsCheckedIn(eventID string) ([]string, error) {
	rows, err := gs.DB.Query("SELECT name from guest where eventID = $1 and checkedIn = TRUE", eventID)
	if err != nil {
		return nil, errors.New("Cannot fetch checked in guest names: " + err.Error())
	}
	defer rows.Close() //make sure this is after checking for an error, or this will be a nil pointer dereference
	numGuests, err := gs.getNumberOfGuestsCheckInStatus(eventID, true)
	if err != nil {
		return nil, errors.New("Cannot fetch number of guests checked in: " + err.Error())
	}

	return gs.scanRowsIntoNames(rows, numGuests)
}

//GuestsNotCheckedIn returns an array of guests who haven't checked into the
//event
func (gs *GuestService) GuestsNotCheckedIn(eventID string) ([]string, error) {
	rows, err := gs.DB.Query("SELECT name from guest where eventID = $1 and checkedIn = FALSE", eventID)
	if err != nil {
		return nil, errors.New("Cannot fetch not checked in guest names: " + err.Error())
	}
	defer rows.Close() //make sure this is after checking for an error, or this will be a nil pointer dereference
	numGuests, err := gs.getNumberOfGuestsCheckInStatus(eventID, false)
	if err != nil {
		return nil, errors.New("Cannot fetch number of guests not checked in: " + err.Error())
	}

	return gs.scanRowsIntoNames(rows, numGuests)
}

//GuestExists returns true if a Guest with the given NRIC identifier (last 5 digits of NRIC)
//and attending the given event exists
func (gs *GuestService) GuestExists(eventID string, nric string) (bool, error) {
	nricHash, err := gs.HM.HashAndSalt(nric)
	if err != nil {
		return false, errors.New("Error hashing NRIC: " + err.Error())
	}

	var numGuests int
	err = gs.DB.QueryRow("SELECT COUNT(*) from guest where eventID = $1 and nricHash = $2",
		eventID, nricHash).Scan(&numGuests)
	if err != nil {
		return false, err
	}
	return numGuests != 0, nil
}

//RegisterGuest adds a guest with the given nric, name and event that they're attending
//to the database, i.e. "registers" them for the event
func (gs *GuestService) RegisterGuest(eventID string, nric string, name string) error {
	nricHash, err := gs.HM.HashAndSalt(nric)
	if err != nil {
		return errors.New("Error hashing NRIC: " + err.Error())
	}

	_, err = gs.DB.Exec("INSERT into guest(nricHash,eventID,name,checkedIn) VALUES($1,$2,$3,FALSE)",
		nricHash, eventID, name)

	return err
}

//RemoveGuest removes a given guest (indicated by nric) from the database
func (gs *GuestService) RemoveGuest(eventID string, nric string) error {
	nricHash, err := gs.HM.HashAndSalt(nric)
	if err != nil {
		return errors.New("Error hashing NRIC: " + err.Error())
	}
	_, err = gs.DB.Exec("DELETE from guest where eventID = $1 and nricHash = $2",
		eventID, nricHash)

	return err
}

//CheckInStats returns statistics relating to the attendance of the given endedvent
//See checkin.CheckinStats for the exact information returned
func (gs *GuestService) CheckInStats(eventID string) (checkin.GuestStats, error) {
	total, err := gs.getNumberOfGuests(eventID)
	if err != nil {
		return checkin.GuestStats{}, errors.New("Error fetching total number of guests: " + err.Error())
	}

	checkedIn, err := gs.getNumberOfGuestsCheckInStatus(eventID, true)
	if err != nil {
		return checkin.GuestStats{}, errors.New("Error fetching checked in count:" + err.Error())
	}
	var percent float64
	if total == 0 {
		percent = 0
	} else {
		percent = float64(checkedIn) / float64(total)
	}
	return checkin.GuestStats{
		TotalGuests:      total,
		CheckedIn:        checkedIn,
		PercentCheckedIn: percent,
	}, nil
}

func (gs *GuestService) getNumberOfGuests(eventID string) (int, error) {
	var i int
	err := gs.DB.QueryRow("SELECT count(*) from guest where eventID = $1", eventID).Scan(&i)
	if err != nil {
		return 0, errors.New("Cannot fetch guest count: " + err.Error())
	}

	return i, nil
}

func (gs *GuestService) getNumberOfGuestsCheckInStatus(eventID string, checkInStatus bool) (int, error) {
	var i int
	err := gs.DB.QueryRow("SELECT count(*) from guest where eventID = $1 and checkedIn = $2",
		eventID, checkInStatus).Scan(&i)
	if err != nil {
		return 0, errors.New("Cannot fetch guest count: " + err.Error())
	}

	return i, nil
}

func (gs *GuestService) scanRowsIntoNames(rows *sql.Rows, rowCount int) ([]string, error) {
	names := make([]string, rowCount)

	index := 0
	for ok := rows.Next(); ok; ok = rows.Next() {
		var name string
		err := rows.Scan(&name)
		if err != nil {
			return nil, errors.New("Could not extract guest name: " + err.Error())
		}
		names[index] = name
		index++
	}

	return names, nil
}
