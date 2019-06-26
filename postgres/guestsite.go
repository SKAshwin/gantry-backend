package postgres

import (
	"checkin"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

//GuestSiteService Implementation of a guest site service
//Needs to be supplied with a database connection as well as a hashing method
type GuestSiteService struct {
	DB *sqlx.DB
}

//GuestSite returns the guest site of that event, if it exists
//If it does not exist, returns an error
//If said event does not exist, returns an error
func (gss *GuestSiteService) GuestSite(eventID string) (checkin.GuestSite, error) {
	var dataString string
	err := gss.DB.QueryRow("SELECT data from website where eventID = $1", eventID).Scan(&dataString)
	if err != nil {
		return checkin.GuestSite{}, errors.New("Error fetching guest site: " + err.Error())
	}

	var siteData checkin.GuestSite
	err = json.Unmarshal([]byte(dataString), &siteData)
	if err != nil {
		return checkin.GuestSite{}, errors.New("Error marshalling guest site data into struct: " + err.Error())
	}
	return siteData, nil
}

//CreateGuestSite creates a website attached to the event specified by the
//eventID
//if the eventID does not point to an existing event, throws an error
func (gss *GuestSiteService) CreateGuestSite(eventID string, site checkin.GuestSite) error {
	dataString, err := json.Marshal(site)
	if err != nil {
		return errors.New("Error marshalling site into dataString: " + err.Error())
	}
	_, err = gss.DB.Exec("INSERT into website (eventID, data) VALUES ($1, $2)", eventID, dataString)
	return err
}

//UpdateGuestSite overwrites an existing eventID's guest site
//returns an error if the event does not exist
//returns an error if there is no website currently associated with the event
func (gss *GuestSiteService) UpdateGuestSite(eventID string, site checkin.GuestSite) error {
	dataString, err := json.Marshal(site)
	if err != nil {
		return errors.New("Error marshalling site into dataString: " + err.Error())
	}
	res, err := gss.DB.Exec("UPDATE website SET data = $1 where eventID = $2", dataString, eventID)
	if err != nil {
		return errors.New("Error in update query: " + err.Error())
	} else if rows, err := res.RowsAffected(); err != nil {
		return errors.New("Error checking whether update affected any rows: " + err.Error())
	} else if rows == 0 {
		return errors.New("No such event or website exists to update/No rows affected by update")
	}
	return nil

}

//DeleteGuestSite will delete the guest site associated with that event, if it exists
//If there is no guest site (or no such event), will not return an error, will merely delete nothing
func (gss *GuestSiteService) DeleteGuestSite(eventID string) error {
	if _, err := uuid.Parse(eventID); err != nil {
		return nil //do nothing, event/website does not exist if UUID is invalid
	}
	_, err := gss.DB.Exec("DELETE from website where eventID = $1", eventID)
	return err
}

//GuestSiteExists returns true if a guest site exists for that event, will return false otherwise
//Will return false (no error) if event does not exist
//Returns an error if theres an issue accessing the database
func (gss *GuestSiteService) GuestSiteExists(eventID string) (bool, error) {
	if _, err := uuid.Parse(eventID); err != nil {
		return false, nil
	}
	var numsite int
	err := gss.DB.QueryRow("SELECT count(*) from website where eventID = $1", eventID).Scan(&numsite)
	return numsite == 1, err
}
