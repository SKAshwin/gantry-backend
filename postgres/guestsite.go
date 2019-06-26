package postgres

import (
	"checkin"
	"errors"

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
	return checkin.GuestSite{}, errors.New("Not implemented")
}

//CreateGuestSite creates a website attached to the event specified by the
//eventID
//if the eventID does not point to an existing event, throws an error
func (gss *GuestSiteService) CreateGuestSite(eventID string, site checkin.GuestSite) error {
	return errors.New("Not implemented")

}

//UpdateGuestSite overwrites an existing eventID's guest site
//returns an error if the event does not exist
//returns an error if there is no website currently associated with the event
func (gss *GuestSiteService) UpdateGuestSite(eventID string, site checkin.GuestSite) error {
	return errors.New("Not implemented")

}

//DeleteGuestSite will delete the guest site associated with that event, if it exists
//If there is no guest site, will not return an error, will merely delete nothing
func (gss *GuestSiteService) DeleteGuestSite(eventID string) error {
	return errors.New("Not implemented")
}

//GuestSiteExists returns true if a guest site exists for that event, will return false otherwise
//Returns an error if theres an issue accessing the database
func (gss *GuestSiteService) GuestSiteExists(eventID string) (bool, error) {
	return false, errors.New("Not implemented")
}
