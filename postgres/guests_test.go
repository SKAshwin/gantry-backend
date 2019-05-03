package postgres_test

//DONT RUN THESE TESTS IN PARALLEL
//SOME MODIFY DATABASE STATE AND CLEAN UP BEFORE EXITING
import (
	"checkin"
	"checkin/mock"
	"checkin/postgres"
	"checkin/test"
	"errors"
	"testing"
)

func hashFnGenerator(err error) func(string) (string, error) {
	return func(pwd string) (string, error) {
		if err != nil {
			return "", err
		}
		//take last char and put it in front
		return string(pwd[len(pwd)-1]) + pwd[:len(pwd)-1], nil
	}
}

func compareHashAndPasswordGenerator() func(string, string) bool {
	return func(hash string, pwd string) bool {
		return hash == (string(pwd[len(pwd)-1]) + pwd[:len(pwd)-1])
	}
}

func TestGuests(t *testing.T) {
	var hm mock.HashMethod
	gs := postgres.GuestService{DB: db, HM: &hm}

	names, err := gs.Guests("aa19239f-f9f5-4935-b1f7-0edfdceabba7", []string{})
	test.Ok(t, err)
	expectedNames := make([]string, 10)
	for i := range expectedNames { // see testData.sql for why
		//get letters of alphabet from A to J
		expectedNames[i] = string('A' + byte(i))
	}
	test.Equals(t, expectedNames, names)

	//nil or empty string array do the same thing
	names2, err := gs.Guests("aa19239f-f9f5-4935-b1f7-0edfdceabba7", nil)
	test.Ok(t, err)
	test.Equals(t, names, names2)

	//check that tag searching works as expected
	names, err = gs.Guests("aa19239f-f9f5-4935-b1f7-0edfdceabba7", []string{"VIP"})
	test.Ok(t, err)
	expectedNames = []string{"C", "D", "H", "I"}
	test.Equals(t, expectedNames, names)

	names, err = gs.Guests("aa19239f-f9f5-4935-b1f7-0edfdceabba7", []string{"ATTENDING"})
	test.Ok(t, err)
	expectedNames = []string{"C", "E", "H", "J"}
	test.Equals(t, expectedNames, names)

	names, err = gs.Guests("aa19239f-f9f5-4935-b1f7-0edfdceabba7", []string{"ATTENDING", "VIP"})
	test.Ok(t, err)
	expectedNames = []string{"C", "H"}
	test.Equals(t, expectedNames, names)

	//empty array, not nil, if no guests fetched
	names, err = gs.Guests("aa19239f-f9f5-4935-b1f7-0edfdceabba7", []string{"UNKNOWNTAG"})
	test.Ok(t, err)
	expectedNames = []string{}
	test.Equals(t, expectedNames, names)

	names, err = gs.Guests("3820a980-a207-4738-b82b-45808fe7aba8", []string{})
	test.Ok(t, err)
	expectedNames = []string{}
	test.Equals(t, expectedNames, names)
}

func TestGuestsCheckedIn(t *testing.T) {
	var hm mock.HashMethod
	gs := postgres.GuestService{DB: db, HM: &hm}

	names, err := gs.GuestsCheckedIn("aa19239f-f9f5-4935-b1f7-0edfdceabba7", []string{})
	test.Ok(t, err)
	expectedNames := []string{"F", "G", "H", "I", "J"}
	test.Equals(t, expectedNames, names)

	//nil or empty string array do the same thing
	names2, err := gs.GuestsCheckedIn("aa19239f-f9f5-4935-b1f7-0edfdceabba7", nil)
	test.Ok(t, err)
	test.Equals(t, names, names2)

	//check that tag searching works as expected
	names, err = gs.GuestsCheckedIn("aa19239f-f9f5-4935-b1f7-0edfdceabba7", []string{"VIP"})
	test.Ok(t, err)
	expectedNames = []string{"H", "I"}
	test.Equals(t, expectedNames, names)

	names, err = gs.GuestsCheckedIn("aa19239f-f9f5-4935-b1f7-0edfdceabba7", []string{"ATTENDING"})
	test.Ok(t, err)
	expectedNames = []string{"H", "J"}
	test.Equals(t, expectedNames, names)

	names, err = gs.GuestsCheckedIn("aa19239f-f9f5-4935-b1f7-0edfdceabba7", []string{"ATTENDING", "VIP"})
	test.Ok(t, err)
	expectedNames = []string{"H"}
	test.Equals(t, expectedNames, names)

	//empty array, not nil, if no guests fetched
	names, err = gs.GuestsCheckedIn("aa19239f-f9f5-4935-b1f7-0edfdceabba7", []string{"UNKNOWNTAG"})
	test.Ok(t, err)
	expectedNames = []string{}
	test.Equals(t, expectedNames, names)

	//this event has no people straight up
	names, err = gs.GuestsCheckedIn("3820a980-a207-4738-b82b-45808fe7aba8", []string{})
	test.Ok(t, err)
	expectedNames = []string{}
	test.Equals(t, expectedNames, names)

	//this event has no checked in people
	names, err = gs.GuestsCheckedIn("03293b3b-df83-407e-b836-fb7d4a3c4966", []string{})
	test.Ok(t, err)
	expectedNames = []string{}
	test.Equals(t, expectedNames, names)
}

func TestGuestsNotCheckedIn(t *testing.T) {
	var hm mock.HashMethod
	gs := postgres.GuestService{DB: db, HM: &hm}

	names, err := gs.GuestsNotCheckedIn("aa19239f-f9f5-4935-b1f7-0edfdceabba7", []string{})
	test.Ok(t, err)
	expectedNames := []string{"A", "B", "C", "D", "E"}
	test.Equals(t, expectedNames, names)

	//nil or empty string array do the same thing
	names2, err := gs.GuestsNotCheckedIn("aa19239f-f9f5-4935-b1f7-0edfdceabba7", nil)
	test.Ok(t, err)
	test.Equals(t, names, names2)

	//check that tag searching works as expected
	names, err = gs.GuestsNotCheckedIn("aa19239f-f9f5-4935-b1f7-0edfdceabba7", []string{"VIP"})
	test.Ok(t, err)
	expectedNames = []string{"C", "D"}
	test.Equals(t, expectedNames, names)

	names, err = gs.GuestsNotCheckedIn("aa19239f-f9f5-4935-b1f7-0edfdceabba7", []string{"ATTENDING"})
	test.Ok(t, err)
	expectedNames = []string{"C", "E"}
	test.Equals(t, expectedNames, names)

	names, err = gs.GuestsNotCheckedIn("aa19239f-f9f5-4935-b1f7-0edfdceabba7", []string{"ATTENDING", "VIP"})
	test.Ok(t, err)
	expectedNames = []string{"C"}
	test.Equals(t, expectedNames, names)

	//empty array, not nil, if no guests fetched
	names, err = gs.GuestsNotCheckedIn("aa19239f-f9f5-4935-b1f7-0edfdceabba7", []string{"UNKNOWNTAG"})
	test.Ok(t, err)
	expectedNames = []string{}
	test.Equals(t, expectedNames, names)

	//this event has no guests
	names, err = gs.GuestsNotCheckedIn("3820a980-a207-4738-b82b-45808fe7aba8", []string{})
	test.Ok(t, err)
	expectedNames = []string{}
	test.Equals(t, expectedNames, names)
}

func TestCheckInStats(t *testing.T) {
	var hm mock.HashMethod
	gs := postgres.GuestService{DB: db, HM: &hm}

	stats, err := gs.CheckInStats("aa19239f-f9f5-4935-b1f7-0edfdceabba7", []string{})
	test.Ok(t, err)
	expectedStats := checkin.GuestStats{
		TotalGuests:      10,
		CheckedIn:        5,
		PercentCheckedIn: 0.5,
	}
	test.Equals(t, expectedStats, stats)

	//nil or empty string array do the same thing
	stats2, err := gs.CheckInStats("aa19239f-f9f5-4935-b1f7-0edfdceabba7", nil)
	test.Ok(t, err)
	test.Equals(t, stats, stats2)

	//check that tag searching works as expected
	stats, err = gs.CheckInStats("aa19239f-f9f5-4935-b1f7-0edfdceabba7", []string{"VIP"})
	test.Ok(t, err)
	expectedStats = checkin.GuestStats{
		TotalGuests:      4,
		CheckedIn:        2,
		PercentCheckedIn: 0.5,
	}
	test.Equals(t, expectedStats, stats)

	stats, err = gs.CheckInStats("aa19239f-f9f5-4935-b1f7-0edfdceabba7", []string{"ATTENDING"})
	test.Ok(t, err)
	expectedStats = checkin.GuestStats{
		TotalGuests:      4,
		CheckedIn:        2,
		PercentCheckedIn: 0.5,
	}
	test.Equals(t, expectedStats, stats)

	stats, err = gs.CheckInStats("aa19239f-f9f5-4935-b1f7-0edfdceabba7", []string{"ATTENDING", "VIP"})
	test.Ok(t, err)
	expectedStats = checkin.GuestStats{
		TotalGuests:      2,
		CheckedIn:        1,
		PercentCheckedIn: 0.5,
	}
	test.Equals(t, expectedStats, stats)

	//empty stats, not nil, if no guests fetched
	stats, err = gs.CheckInStats("aa19239f-f9f5-4935-b1f7-0edfdceabba7", []string{"UNKNOWNTAG"})
	test.Ok(t, err)
	expectedStats = checkin.GuestStats{
		TotalGuests:      0,
		CheckedIn:        0,
		PercentCheckedIn: 0,
	}
	test.Equals(t, expectedStats, stats)

	//this event has no people straight up
	stats, err = gs.CheckInStats("3820a980-a207-4738-b82b-45808fe7aba8", []string{})
	test.Ok(t, err)
	expectedStats = checkin.GuestStats{
		TotalGuests:      0,
		CheckedIn:        0,
		PercentCheckedIn: 0,
	}
	test.Equals(t, expectedStats, stats)

	//this event has no checked in people
	stats, err = gs.CheckInStats("03293b3b-df83-407e-b836-fb7d4a3c4966", []string{})
	test.Ok(t, err)
	expectedStats = checkin.GuestStats{
		TotalGuests:      1,
		CheckedIn:        0,
		PercentCheckedIn: 0,
	}
	test.Equals(t, expectedStats, stats)
}

func TestRegisterGuest(t *testing.T) {
	var hm mock.HashMethod
	gs := postgres.GuestService{DB: db, HM: &hm}

	hm.HashAndSaltFn = hashFnGenerator(nil)
	hm.CompareHashAndPasswordFn = compareHashAndPasswordGenerator()

	err := gs.RegisterGuest("3820a980-a207-4738-b82b-45808fe7aba8", checkin.Guest{NRIC: "1234A", Name: "Jim Bob", Tags: []string{"NEWLYREGISTERED"}})
	test.Ok(t, err)
	names, err := gs.Guests("3820a980-a207-4738-b82b-45808fe7aba8", []string{"NEWLYREGISTERED"})
	test.Ok(t, err)
	test.Equals(t, []string{"Jim Bob"}, names)
	err = gs.RemoveGuest("3820a980-a207-4738-b82b-45808fe7aba8", "1234A")
	test.Ok(t, err)

	//check salting fails
	hm.HashAndSaltFn = hashFnGenerator(errors.New("An error"))
	err = gs.RegisterGuest("3820a980-a207-4738-b82b-45808fe7aba8", checkin.Guest{NRIC: "1234A", Name: "Jim Bob", Tags: []string{"NEWLYREGISTERED"}})
	test.Assert(t, err != nil, "Failed hashing does not throw an error")
	hm.HashAndSaltFn = hashFnGenerator(nil)

	//check event does not even exist
	err = gs.RegisterGuest("a6db3963-5389-4dbe-8fc6-bbd7f7ce66b8", checkin.Guest{NRIC: "1234A", Name: "Jim Bob", Tags: []string{"NEWLYREGISTERED"}})
	test.Assert(t, err != nil, "Registering guest for non-existent event does not throw an error")

	//check nil tag and empty tag do the same thing
	err = gs.RegisterGuest("3820a980-a207-4738-b82b-45808fe7aba8", checkin.Guest{NRIC: "1234A", Name: "Jim Bob", Tags: nil})
	test.Ok(t, err)
	names, err = gs.Guests("3820a980-a207-4738-b82b-45808fe7aba8", []string{})
	test.Ok(t, err)
	test.Equals(t, []string{"Jim Bob"}, names)
	err = gs.RemoveGuest("3820a980-a207-4738-b82b-45808fe7aba8", "1234A")
	test.Ok(t, err)

	err = gs.RegisterGuest("3820a980-a207-4738-b82b-45808fe7aba8", checkin.Guest{NRIC: "1234A", Name: "Jim Bob", Tags: []string{}})
	test.Ok(t, err)
	names, err = gs.Guests("3820a980-a207-4738-b82b-45808fe7aba8", []string{})
	test.Ok(t, err)
	test.Equals(t, []string{"Jim Bob"}, names)
	err = gs.RemoveGuest("3820a980-a207-4738-b82b-45808fe7aba8", "1234A")
	test.Ok(t, err)

	//check case insensitivity of NRIC
	err = gs.RegisterGuest("3820a980-a207-4738-b82b-45808fe7aba8", checkin.Guest{NRIC: "1234A", Name: "Jim Bob", Tags: []string{}})
	test.Ok(t, err)
	err = gs.RegisterGuest("3820a980-a207-4738-b82b-45808fe7aba8", checkin.Guest{NRIC: "1234a", Name: "Other name", Tags: []string{}})
	test.Assert(t, err != nil, "Registering same guest but with different last char did not throw an error (no case insensitivity)")
	err = gs.RemoveGuest("3820a980-a207-4738-b82b-45808fe7aba8", "1234A")
	test.Ok(t, err)

	//check that its empty now, before moving on to next test
	names, err = gs.Guests("3820a980-a207-4738-b82b-45808fe7aba8", []string{})
	test.Ok(t, err)
	test.Equals(t, []string{}, names)
}

func TestTags(t *testing.T) {
	var hm mock.HashMethod
	gs := postgres.GuestService{DB: db, HM: &hm}

	hm.CompareHashAndPasswordFn = compareHashAndPasswordGenerator()

	//normal functionality
	tags, err := gs.Tags("aa19239f-f9f5-4935-b1f7-0edfdceabba7", "2346C")
	test.Ok(t, err)
	test.Equals(t, []string{"VIP", "ATTENDING"}, tags)
	test.Assert(t, hm.CompareHashAndPasswordInvoked, "No expected comparison of hashes in function call")

	//guest doesn't exist
	tags, err = gs.Tags("aa19239f-f9f5-4935-b1f7-0edfdceabba7", "6433G")
	test.Assert(t, err != nil, "No error when fetching tags of non-existent guests")

	//event does not exist
	tags, err = gs.Tags("a6db3963-5389-4dbe-8fc6-bbd7f7ce66b8", "6433G")
	test.Assert(t, err != nil, "No error when fetching tags of guest of non-existent event")

	tags, err = gs.Tags("notevenauuid", "6433G")
	test.Assert(t, err != nil, "No error when fetching tags of guest of non-existent event (invalid UUID)")

	//no tags should be empty array not nil
	tags, err = gs.Tags("aa19239f-f9f5-4935-b1f7-0edfdceabba7", "1234A")
	test.Ok(t, err)
	test.Equals(t, []string{}, tags)

}

func TestSetTags(t *testing.T) {
	var hm mock.HashMethod
	gs := postgres.GuestService{DB: db, HM: &hm}

	hm.CompareHashAndPasswordFn = compareHashAndPasswordGenerator()

	unaffectedTags, err := gs.Tags("aa19239f-f9f5-4935-b1f7-0edfdceabba7", "5678B")
	test.Ok(t, err)

	//test normal functionality
	err = gs.SetTags("aa19239f-f9f5-4935-b1f7-0edfdceabba7", "1234A", []string{"HELLO", "WORLD"})
	test.Ok(t, err)
	tags, err := gs.Tags("aa19239f-f9f5-4935-b1f7-0edfdceabba7", "1234A")
	test.Ok(t, err)
	test.Equals(t, []string{"HELLO", "WORLD"}, tags)

	//test if the guest already has some tags (new tags should overwrite old)
	err = gs.SetTags("aa19239f-f9f5-4935-b1f7-0edfdceabba7", "2346C", []string{"HELLO", "WORLD"})
	test.Ok(t, err)
	tags, err = gs.Tags("aa19239f-f9f5-4935-b1f7-0edfdceabba7", "2346C")
	test.Ok(t, err)
	test.Equals(t, []string{"HELLO", "WORLD"}, tags)

	//make sure unrelated guests not affected
	newUnaffectedTags, err := gs.Tags("aa19239f-f9f5-4935-b1f7-0edfdceabba7", "5678B")
	test.Ok(t, err)
	test.Equals(t, unaffectedTags, newUnaffectedTags)

	//test nil and empty array tags
	//nils should set tags to empty array
	err = gs.SetTags("aa19239f-f9f5-4935-b1f7-0edfdceabba7", "2346c", nil) //also random case sensitivity check - should be case insensitive
	test.Ok(t, err)
	tags, err = gs.Tags("aa19239f-f9f5-4935-b1f7-0edfdceabba7", "2346C")
	test.Ok(t, err)
	test.Equals(t, []string{}, tags)

	err = gs.SetTags("aa19239f-f9f5-4935-b1f7-0edfdceabba7", "2346C", []string{})
	test.Ok(t, err)
	tags, err = gs.Tags("aa19239f-f9f5-4935-b1f7-0edfdceabba7", "2346C")
	test.Ok(t, err)
	test.Equals(t, []string{}, tags)

	//test guest or event does not exist
	err = gs.SetTags("aa19239f-f9f5-4935-b1f7-0edfdceabba7", "1111B", []string{})
	test.Assert(t, err != nil, "No error returned for nonexistent guest")
	err = gs.SetTags("a6db3963-5389-4dbe-8fc6-bbd7f7ce66b8", "2346C", []string{})
	test.Assert(t, err != nil, "No error returned for nonexistent event")

}