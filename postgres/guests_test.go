package postgres_test

//DONT RUN THESE TESTS IN PARALLEL
//SOME MODIFY DATABASE STATE AND CLEAN UP BEFORE EXITING
import (
	"checkin"
	"checkin/bcrypt"
	"checkin/mock"
	"checkin/postgres"
	"checkin/test"
	"log"
	"os"
	"sort"
	"strconv"
	"testing"
	"time"
)

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

	err := gs.RegisterGuest("3820a980-a207-4738-b82b-45808fe7aba8", checkin.Guest{NRIC: "1234A", Name: "Jim Bob", Tags: []string{"NEWLYREGISTERED"}})
	test.Ok(t, err)
	names, err := gs.Guests("3820a980-a207-4738-b82b-45808fe7aba8", []string{"NEWLYREGISTERED"})
	test.Ok(t, err)
	test.Equals(t, []string{"Jim Bob"}, names)
	err = gs.RemoveGuest("3820a980-a207-4738-b82b-45808fe7aba8", "1234A")
	test.Ok(t, err)

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

func TestRegisterGuests(t *testing.T) {
	var hm mock.HashMethod
	gs := postgres.GuestService{DB: db, HM: &hm}

	//tests all kinds of valid inputs
	guests := []checkin.Guest{
		checkin.Guest{NRIC: "1234A", Name: "Jim Bob", Tags: []string{"NEWLYREGISTERED"}},
		checkin.Guest{NRIC: "1234B", Name: "Mayank", Tags: nil},
		checkin.Guest{NRIC: "1234C", Name: "Eugene", Tags: []string{}},
		checkin.Guest{NRIC: "1234D", Name: "Ya wei", Tags: []string{"SOME", "THING"}},
	}

	//test normal functionality
	err := gs.RegisterGuests("3820a980-a207-4738-b82b-45808fe7aba8", guests)
	test.Ok(t, err)
	names, err := gs.Guests("3820a980-a207-4738-b82b-45808fe7aba8", nil)
	sort.Strings(names)
	test.Equals(t, []string{"Eugene", "Jim Bob", "Mayank", "Ya wei"}, names)
	for _, guest := range guests {
		err := gs.RemoveGuest("3820a980-a207-4738-b82b-45808fe7aba8", guest.NRIC)
		test.Ok(t, err)
	}

	//check empty or nil slice of guests - should fail
	err = gs.RegisterGuests("3820a980-a207-4738-b82b-45808fe7aba8", []checkin.Guest{})
	test.Assert(t, err != nil, "No error thrown when attempting to register empty slice of guests")
	err = gs.RegisterGuests("3820a980-a207-4738-b82b-45808fe7aba8", nil)
	test.Assert(t, err != nil, "No error thrown when attempting to register nil guest slice")

	//check event does not even exist
	err = gs.RegisterGuests("a6db3963-5389-4dbe-8fc6-bbd7f7ce66b8", guests)
	test.Assert(t, err != nil, "Registering guests for non-existent event does not throw an error")

	//check case insensitivity of NRIC
	guests = []checkin.Guest{
		checkin.Guest{NRIC: "1234A", Name: "Jim Bob", Tags: []string{"NEWLYREGISTERED"}},
		checkin.Guest{NRIC: "1234a", Name: "Mayank", Tags: nil},
	}
	err = gs.RegisterGuests("a6db3963-5389-4dbe-8fc6-bbd7f7ce66b8", guests)
	test.Assert(t, err != nil, "Registering two identical guests (with only NRIC case differing) for non-existent event does not throw an error")
	names, err = gs.Guests("a6db3963-5389-4dbe-8fc6-bbd7f7ce66b8", nil)
	test.Ok(t, err)
	test.Equals(t, []string{}, names) //test that an error in registering one guest, for example 1234a, means neither are registered

	//check that its empty now, before moving on to next test
	names, err = gs.Guests("3820a980-a207-4738-b82b-45808fe7aba8", []string{})
	test.Ok(t, err)
	test.Equals(t, []string{}, names)

}

func TestTags(t *testing.T) {
	var hm mock.HashMethod
	gs := postgres.GuestService{DB: db, HM: &hm}

	//normal functionality
	tags, err := gs.Tags("aa19239f-f9f5-4935-b1f7-0edfdceabba7", "2346C")
	test.Ok(t, err)
	test.Equals(t, []string{"VIP", "ATTENDING"}, tags)

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

func TestAllTags(t *testing.T) {
	var hm mock.HashMethod
	gs := postgres.GuestService{DB: db, HM: &hm}

	//check normal functionality (multiple tags returns)
	tags, err := gs.AllTags("c14a592c-950d-44ba-b173-bbb9e4f5c8b4")
	test.Ok(t, err)
	sort.Strings(tags)
	test.Equals(t, []string{"ATTENDING", "OFFICER", "VIP"}, tags)

	//test no tags (return empty array)
	tags, err = gs.AllTags("03293b3b-df83-407e-b836-fb7d4a3c4966")
	test.Ok(t, err)
	test.Equals(t, []string{}, tags)

	//test event with no guests (return empty array)
	tags, err = gs.AllTags("3820a980-a207-4738-b82b-45808fe7aba8")
	test.Ok(t, err)
	test.Equals(t, []string{}, tags)

	//test event does not exist (should also return empty array)
	tags, err = gs.AllTags("1f73ed02-9427-41e4-9469-c9c4ac515f8d")
	test.Ok(t, err)
	test.Equals(t, []string{}, tags)

	//test event does not exist, invalid UUID (should also return an empty array)
	tags, err = gs.AllTags("ayylmao")
	test.Ok(t, err)
	test.Equals(t, []string{}, tags)
	tags, err = gs.AllTags("")
	test.Ok(t, err)
	test.Equals(t, []string{}, tags)
}

func TestGuestExists(t *testing.T) {
	var hm mock.HashMethod
	gs := postgres.GuestService{DB: db, HM: &hm}

	//event exists, guest does
	exists, err := gs.GuestExists("aa19239f-f9f5-4935-b1f7-0edfdceabba7", "5678B")
	test.Ok(t, err)
	test.Equals(t, true, exists)

	//event exists, guest does not
	exists, err = gs.GuestExists("aa19239f-f9f5-4935-b1f7-0edfdceabba7", "4384S")
	test.Ok(t, err)
	test.Equals(t, false, exists)

	//guest with that NRIC exists, but for another event
	exists, err = gs.GuestExists("2c59b54d-3422-4bdb-824c-4125775b44c8", "5678B")
	test.Ok(t, err)
	test.Equals(t, false, exists)

	//guest with that nric exists for both this event and another
	exists, err = gs.GuestExists("03293b3b-df83-407e-b836-fb7d4a3c4966", "1234A")
	test.Ok(t, err)
	test.Equals(t, true, exists)

	//event does not exist
	exists, err = gs.GuestExists("f64c91f8-f46a-4ef8-a4a6-cfbc4093e21c", "1234A")
	test.Ok(t, err)
	test.Equals(t, false, exists)

	//invalid UUID
	exists, err = gs.GuestExists("hello!", "1234A")
	test.Ok(t, err)
	test.Equals(t, false, exists)

	//empty string NRIC
	exists, err = gs.GuestExists("03293b3b-df83-407e-b836-fb7d4a3c4966", "")
	test.Ok(t, err)
	test.Equals(t, false, exists)

	//empty string UUID
	exists, err = gs.GuestExists("", "1234A")
	test.Ok(t, err)
	test.Equals(t, false, exists)
}

func TestCheckIn(t *testing.T) {
	var hm mock.HashMethod
	gs := postgres.GuestService{HM: &hm, DB: db}

	//test normal functionality
	name, err := gs.CheckIn("03293b3b-df83-407e-b836-fb7d4a3c4966", "1234A")
	test.Ok(t, err)
	test.Equals(t, "A", name)
	names, err := gs.GuestsCheckedIn("03293b3b-df83-407e-b836-fb7d4a3c4966", nil)
	test.Ok(t, err)
	test.Equals(t, []string{"A"}, names)

	//test guest already checked in (should work fine)
	name, err = gs.CheckIn("03293b3b-df83-407e-b836-fb7d4a3c4966", "1234A")
	test.Ok(t, err)
	test.Equals(t, "A", name)
	names, err = gs.GuestsCheckedIn("03293b3b-df83-407e-b836-fb7d4a3c4966", nil)
	test.Ok(t, err)
	test.Equals(t, []string{"A"}, names)

	err = gs.MarkAbsent("03293b3b-df83-407e-b836-fb7d4a3c4966", "1234A")
	test.Ok(t, err)

	//test guest does not exist (eventID for different event) (should throw error)
	name, err = gs.CheckIn("2c59b54d-3422-4bdb-824c-4125775b44c8", "1234A")
	test.Assert(t, err != nil, "No error thrown when check in called with non-existent guest")

	//test guest does not exist (NRIC wrong) (should throw error)
	name, err = gs.CheckIn("03293b3b-df83-407e-b836-fb7d4a3c4966", "3118B")
	test.Assert(t, err != nil, "No error thrown when check in called with non-existent guest")

	//test invalid UUID eventID (should throw error)
	name, err = gs.CheckIn("1312312312", "3118B")
	test.Assert(t, err != nil, "No error thrown when check in called with non-existent guest (invalid UUID)")
}

func TestGuestExistsTime(t *testing.T) {
	if os.Getenv("BENCH") == "" {
		t.Skip("Skipping testing since no BENCH argument provided")
	}
	hm := bcrypt.HashMethod{HashCost: 5}
	gs := postgres.GuestService{HM: &hm, DB: db}
	eventID := "3820a980-a207-4738-b82b-45808fe7aba8"

	for i := range [1000]int{} {
		err := gs.RegisterGuest(eventID, checkin.Guest{NRIC: strconv.Itoa(i), Name: strconv.Itoa(i)})
		test.Ok(t, err)
	}

	start := time.Now()
	res, err := gs.GuestExists(eventID, strconv.Itoa(1000))
	timeTaken := time.Since(start).Seconds()
	log.Println(timeTaken)
	test.Ok(t, err)
	test.Equals(t, false, res)

	res, err = gs.GuestExists(eventID, strconv.Itoa(999))
	test.Ok(t, err)
	test.Equals(t, true, res)

	res, err = gs.GuestExists(eventID, "lol fuck your logic")
	test.Ok(t, err)
	test.Equals(t, false, res)

	res, err = gs.GuestExists(eventID, strconv.Itoa(19))
	test.Ok(t, err)
	test.Equals(t, true, res)

	res, err = gs.GuestExists(eventID, strconv.Itoa(0))
	test.Ok(t, err)
	test.Equals(t, true, res)

}
