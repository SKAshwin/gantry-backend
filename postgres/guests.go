package postgres

//UNTESTED
import (
	"checkin"
	"database/sql"
	"errors"

	"github.com/google/uuid"

	"strings"

	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

//GuestService an implementation of checkin.GuestService using postgres
//Needs a HashMethod as all NRICs are stored internally as hashes for
//security purposes
type GuestService struct {
	DB         *sqlx.DB
	HM         checkin.HashMethod
	guestCache map[string]checkin.Guest //a map of eventID + " " + nric to what guest it corresponds to
	//to speed up the execution of finding a guest
	cacheLock sync.RWMutex //RWMutex to use when reading/writing to the guest cache
}

//CheckIn marks a guest (indicated by the last 5 digits of the nric)
//of a particular event as having attended the event
//Returns the name of the guest who was checked in
//Will return an error if said guest does not exist, or event with
//that ID does not exist
//Will not throw an error if the guest is already checked in
//If any error occurs, check in status of the guest will not be edited
func (gs *GuestService) CheckIn(eventID string, nric string) (string, error) {
	guest, err := gs.getGuestWithNRIC(eventID, nric)
	if err != nil {
		return "", errors.New("Error getting guest with that NRIC: " + err.Error())
	}
	if guest.IsEmpty() {
		return "", errors.New("Guest with that NRIC does not exist: " + nric)
	}
	nricHash := guest.NRIC

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

//MarkAbsent marks a guest of a particular event as being absent, the opposite of check in
//Will return an error if said guest does not exist, or even with that
//ID does not exist
//Will not throw an error if the guest is already not checked in
func (gs *GuestService) MarkAbsent(eventID string, nric string) error {
	guest, err := gs.getGuestWithNRIC(eventID, nric)
	if err != nil {
		return errors.New("Error getting guest with that NRIC: " + err.Error())
	}
	if guest.IsEmpty() {
		return errors.New("Guest with that NRIC does not exist: " + nric)
	}
	nricHash := guest.NRIC

	_, err = gs.DB.Exec("UPDATE guest SET checkedIn = False, checkInTime = NOW() WHERE eventID = $1 and nricHash = $2",
		eventID, nricHash)
	return err
}

//Guests returns an array of names of the guests who are registered/signed up for
//an event given by the eventID
//Can filter the list down to guests which have *all* the tags specified in tags
//A nil tags, or empty string array, will fetch all guests
//No error thrown if event does not exist - just gives empty array, so check existence before calling method
func (gs *GuestService) Guests(eventID string, tags []string) ([]string, error) {
	if tags == nil {
		tags = []string{}
	}
	rows, err := gs.DB.Query("SELECT name from guest where eventID = $1 and $2 <@ tags", eventID, pq.Array(tags))
	if err != nil {
		return nil, errors.New("Cannot fetch guest names: " + err.Error())
	}
	defer rows.Close() //make sure this is after checking for an error, or this will be a nil pointer dereference
	numGuests, err := gs.getNumberOfGuests(eventID, tags)
	if err != nil {
		return nil, errors.New("Cannot fetch number of guests: " + err.Error())
	}

	return gs.scanRowsIntoStrings(rows, numGuests)
}

//GuestsCheckedIn return an array of names of the guests who have checked in
//to the event given by the eventID
//Can filter the list down to guests which have *all* the tags specified in tags
//A nil tags, or empty string array, will fetch all guests
//No error thrown if event does not exist - just gives empty array, so check existence before calling method
func (gs *GuestService) GuestsCheckedIn(eventID string, tags []string) ([]string, error) {
	if tags == nil {
		tags = []string{}
	}
	rows, err := gs.DB.Query("SELECT name from guest where eventID = $1 and checkedIn = TRUE and $2 <@ tags", eventID, pq.Array(tags))
	if err != nil {
		return nil, errors.New("Cannot fetch checked in guest names: " + err.Error())
	}
	defer rows.Close() //make sure this is after checking for an error, or this will be a nil pointer dereference
	numGuests, err := gs.getNumberOfGuestsCheckInStatus(eventID, true, tags)
	if err != nil {
		return nil, errors.New("Cannot fetch number of guests checked in: " + err.Error())
	}

	return gs.scanRowsIntoStrings(rows, numGuests)
}

//GuestsNotCheckedIn returns an array of guests who haven't checked into the
//event
//Can filter the list down to guests which have *all* the tags specified in tags
//A nil tags, or empty string array, will fetch all guests
//No error thrown if event does not exist - just gives empty array, so check existence before calling method
func (gs *GuestService) GuestsNotCheckedIn(eventID string, tags []string) ([]string, error) {
	if tags == nil {
		tags = []string{}
	}
	rows, err := gs.DB.Query("SELECT name from guest where eventID = $1 and checkedIn = FALSE and $2 <@ tags", eventID, pq.Array(tags))
	if err != nil {
		return nil, errors.New("Cannot fetch not checked in guest names: " + err.Error())
	}
	defer rows.Close() //make sure this is after checking for an error, or this will be a nil pointer dereference
	numGuests, err := gs.getNumberOfGuestsCheckInStatus(eventID, false, tags)
	if err != nil {
		return nil, errors.New("Cannot fetch number of guests not checked in: " + err.Error())
	}

	return gs.scanRowsIntoStrings(rows, numGuests)
}

//GuestExists returns true if a Guest with the given NRIC identifier (last 5 digits of NRIC)
//and attending the given event exists
//Returns false if the event does not exist in the first place (NOT an error), so check for event existence
//separately
//Returns an error only if there is an error *checking* if the guest exists (e.g. database connection error)
func (gs *GuestService) GuestExists(eventID string, nric string) (bool, error) {
	guest, err := gs.getGuestWithNRIC(eventID, nric)
	if err != nil {
		return false, errors.New("Error getting guests: " + err.Error())
	}
	return !guest.IsEmpty(), nil
}

//RegisterGuest adds a guest with the given nric, name and event that they're attending
//to the database, i.e. "registers" them for the event
func (gs *GuestService) RegisterGuest(eventID string, guest checkin.Guest) error {
	nricHash, err := gs.HM.HashAndSalt(strings.ToUpper(guest.NRIC))
	if guest.Tags == nil {
		guest.Tags = []string{} //no nils allowed
	}
	if err != nil {
		return errors.New("Error hashing NRIC: " + err.Error())
	}

	_, err = gs.DB.Exec("INSERT into guest(nricHash,eventID,name,tags,checkedIn) VALUES($1,$2,$3,$4,FALSE)",
		nricHash, eventID, guest.Name, pq.Array(guest.Tags))

	return err
}

//RegisterGuests does the same as RegisterGuest, but registers multiple guests, and if there's a failure on
//any one of them no guest will be added (for example, if one of them has already been registered)
//An empty or nil guest array will return an error, as this is presumably not expected input
//if the event provided does not exist, will return an error
func (gs *GuestService) RegisterGuests(eventID string, guests []checkin.Guest) error {
	if guests == nil || len(guests) == 0 {
		return errors.New("Cannot register nil or empty slice of guests")
	}

	tx, err := gs.DB.Beginx()
	if err != nil {
		return errors.New("Error opening transaction: " + err.Error())
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	for _, guest := range guests {
		nricHash, err := gs.HM.HashAndSalt(strings.ToUpper(guest.NRIC))
		if guest.Tags == nil {
			guest.Tags = []string{} //no nils allowed
		}
		if err != nil {
			tx.Rollback()
			return errors.New("Error hashing NRIC: " + err.Error())
		}

		_, err = gs.DB.Exec("INSERT into guest(nricHash,eventID,name,tags,checkedIn) VALUES($1,$2,$3,$4,FALSE)",
			nricHash, eventID, guest.Name, pq.Array(guest.Tags))
		if err != nil {
			tx.Rollback()
			return errors.New("Error inserting one of the guests: " + err.Error())
		}
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return errors.New("Error committing changes to database: " + err.Error())
	}
	return nil
}

//Tags returns the tags of a given guest
//Returns an empty array for no tags
//Returns error if guest does not exist, or there was an error fetching it
func (gs *GuestService) Tags(eventID string, nric string) ([]string, error) {
	guest, err := gs.getGuestWithNRIC(eventID, nric)
	if err != nil {
		return nil, errors.New("Error fetching guest with that NRIC")
	}
	if guest.IsEmpty() {
		return nil, errors.New("Guest does not exist")
	}

	var tags []string
	err = gs.DB.QueryRow("SELECT tags from guest where eventID = $1 and nricHash = $2", eventID, guest.NRIC).Scan(pq.Array(&tags))
	if err != nil {
		return nil, errors.New("Error getting tags: " + err.Error())
	}

	return tags, nil
}

//SetTags sets the tags of a given guest; it overwrites all previous tags on that guest
//nil tags treated as empty array tags
//Error if guest does not exist or error updating/fetching the guest
func (gs *GuestService) SetTags(eventID string, nric string, tags []string) error {
	guest, err := gs.getGuestWithNRIC(eventID, nric)
	if err != nil {
		return errors.New("Error fetching guest with that NRIC: " + err.Error())
	}
	if guest.IsEmpty() {
		return errors.New("Guest does not exist")
	}
	if tags == nil { //treat nils as empty arrays
		tags = []string{}
	}

	_, err = gs.DB.Exec("UPDATE guest SET tags = $1 where eventID = $2 and nricHash = $3", pq.Array(tags), eventID, guest.NRIC)
	return err
}

//AllTags returns all the unique tags for guests in a given event
//Returns an empty string array if there are no tags, or no guests, in a given event
//Also returns an empty string array if the event does not exist (or invalid UUID), NOT AN ERROR
//Returns an error only if there is an error fetching the list of guests from the database
func (gs *GuestService) AllTags(eventID string) ([]string, error) {
	if _, err := uuid.Parse(eventID); err != nil {
		return []string{}, nil
	}

	rows, err := gs.DB.Query("SELECT distinct unnest(tags) from guest where eventID = $1", eventID)
	if err != nil {
		return nil, errors.New("Error fetching tags for event: " + err.Error())
	}
	defer rows.Close()

	count, err := gs.getNumberOfUniqueTags(eventID)
	if err != nil {
		return nil, errors.New("Error fetching number of unique tags: " + err.Error())
	}

	uniqueTags, err := gs.scanRowsIntoStrings(rows, count)
	if err != nil {
		return nil, errors.New("Error scanning rows into a string array: " + err.Error())
	}

	return uniqueTags, nil
}

//RemoveGuest removes a given guest (indicated by nric) from the database
//will not return an error if guest does not exist, will merely delete no one
func (gs *GuestService) RemoveGuest(eventID string, nric string) error {
	guest, err := gs.getGuestWithNRIC(eventID, nric)
	if err != nil {
		return errors.New("Error getting guest with that NRIC: " + err.Error())
	}
	if guest.IsEmpty() {
		return nil
	}
	nricHash := guest.NRIC

	_, err = gs.DB.Exec("DELETE from guest where eventID = $1 and nricHash = $2",
		eventID, nricHash)

	return err
}

//CheckInStats returns statistics relating to the attendance of the given endedvent
//See checkin.CheckinStats for the exact information returned
//Can filter the stats down to counting only guests which have *all* the tags specified in tags
//A nil tags, or empty string array, will use all guests
//No error thrown if event does not exist - just gives empty stats, so check existence before calling method
func (gs *GuestService) CheckInStats(eventID string, tags []string) (checkin.GuestStats, error) {
	total, err := gs.getNumberOfGuests(eventID, tags)
	if err != nil {
		return checkin.GuestStats{}, errors.New("Error fetching total number of guests: " + err.Error())
	}

	checkedIn, err := gs.getNumberOfGuestsCheckInStatus(eventID, true, tags)
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

func (gs *GuestService) getNumberOfUniqueTags(eventID string) (int, error) {
	var i int
	err := gs.DB.QueryRow("SELECT count(distinct tag) from guest, unnest(guest.tags) as tag where eventID = $1", eventID).Scan(&i)

	return i, err
}

//if tags is nil OR an empty array, looks for all guests, ignoring tags
//in general, looks for guests who have all the tags specified in tags
//they could possibly have more
func (gs *GuestService) getNumberOfGuests(eventID string, tags []string) (int, error) {
	var i int
	var err error
	if tags != nil {
		err = gs.DB.QueryRow("SELECT count(*) from guest where eventID = $1 and $2 <@ tags",
			eventID, pq.Array(tags)).Scan(&i)
	} else {
		err = gs.DB.QueryRow("SELECT count(*) from guest where eventID = $1",
			eventID).Scan(&i)
	}

	if err != nil {
		return 0, errors.New("Cannot fetch guest count: " + err.Error())
	}

	return i, nil
}

func (gs *GuestService) getNumberOfGuestsCheckInStatus(eventID string, checkInStatus bool, tags []string) (int, error) {
	var i int
	var err error
	if tags != nil {
		err = gs.DB.QueryRow("SELECT count(*) from guest where eventID = $1 and checkedIn = $2 and $3 <@ tags",
			eventID, checkInStatus, pq.Array(tags)).Scan(&i)
	} else {
		err = gs.DB.QueryRow("SELECT count(*) from guest where eventID = $1 and checkedIn = $2",
			eventID, checkInStatus).Scan(&i)
	}
	if err != nil {
		return 0, errors.New("Cannot fetch guest count: " + err.Error())
	}

	return i, nil
}

func (gs *GuestService) scanRowsIntoStrings(rows *sql.Rows, rowCount int) ([]string, error) {
	strings := make([]string, rowCount)

	index := 0
	for ok := rows.Next(); ok; ok = rows.Next() {
		var str string
		err := rows.Scan(&str)
		if err != nil {
			return nil, errors.New("Could not extract guest name: " + err.Error())
		}
		strings[index] = str
		index++
	}

	return strings, nil
}

func (gs *GuestService) checkCache(eventID string, nric string) (checkin.Guest, bool) {
	gs.cacheLock.RLock()
	defer gs.cacheLock.RUnlock()
	if gs.guestCache == nil {
		gs.guestCache = make(map[string]checkin.Guest)
	}
	guest, ok := gs.guestCache[eventID+" "+nric]
	return guest, ok
}

func (gs *GuestService) addCache(eventID string, nric string, guest checkin.Guest) {
	gs.cacheLock.Lock()
	defer gs.cacheLock.Unlock()
	gs.guestCache[eventID+" "+nric] = guest
}

//Returns a guest and true if one could be found with that nric and eventID
//Returns an empty guest object (and no error) if the guest could not be found
//Returns an error if there is an error getting a guest
//caches
func (gs *GuestService) getGuestWithNRIC(eventID string, nric string) (checkin.Guest, error) {
	if guest, ok := gs.checkCache(eventID, nric); ok {
		return guest, nil
	}

	if _, err := uuid.Parse(eventID); err != nil {
		//attempting to search for a guest associated with an event with an invalid UUID will throw an error
		//since a guest with an invalid UUID will definitely not exist, return an empty guest object
		return checkin.Guest{}, nil
	}
	rows, err := gs.DB.Queryx("SELECT name, nricHash from guest where eventID = $1", eventID)
	if err != nil {
		return checkin.Guest{}, errors.New("Cannot fetch all guests: " + err.Error())
	}
	defer rows.Close()

	numGuests, err := gs.getNumberOfGuests(eventID, nil)
	if err != nil {
		return checkin.Guest{}, errors.New("Error fetching number of guests: " + err.Error())
	}
	guests, err := gs.scanRowsIntoGuests(rows, numGuests)
	if err != nil {
		return checkin.Guest{}, errors.New("Error reading guest data from database: " + err.Error())
	}

	guest := gs.findGuest(nric, guests)
	gs.addCache(eventID, nric, guest)
	return guest, nil
}

func (gs *GuestService) findGuest(nric string, hashedGuests []checkin.Guest) checkin.Guest {
	result := make(chan checkin.Guest) //channel to send a found guest
	quit := make(chan bool)            //channel to signal that all goroutines have finished execution
	found := make(chan bool)           //channel to signal that a guest has been found
	upperNRIC := strings.ToUpper(nric)
	wg := sync.WaitGroup{}
	for i := 0; i < len(hashedGuests); i += 20 {
		wg.Add(1)
		go func(guests []checkin.Guest) {
			defer wg.Done()
			for _, guest := range guests {
				select {
				case <-found:
					//guest found by another goroutine, end execution
					return
				default:
					if gs.HM.CompareHashAndPassword(guest.NRIC, upperNRIC) {
						result <- guest
						close(found)
						return
					}
				}

			}
		}(hashedGuests[i:min(len(hashedGuests), i+20)])
	}
	go func() {
		wg.Wait()
		close(quit)
	}()
	select {
	case guest := <-result: //one of the goroutines found a guest
		return guest
	case <-quit: //Wait finished executing, which means all threads closed, and as the other
		//case did not run, nothing was sent over the result channel
		return checkin.Guest{}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (gs *GuestService) scanRowsIntoGuests(rows *sqlx.Rows, rowCount int) ([]checkin.Guest, error) {
	guests := make([]checkin.Guest, rowCount)

	index := 0
	for thereAreMore := rows.Next(); thereAreMore; thereAreMore = rows.Next() {
		var guest checkin.Guest
		err := rows.StructScan(&guest)
		if err != nil {
			return nil, errors.New("Could not extract guest: " + err.Error())
		}
		guests[index] = guest
		index++
	}

	return guests, nil
}
