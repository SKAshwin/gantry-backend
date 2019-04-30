package postgres_test

import (
	"checkin"
	"checkin/mock"
	"checkin/postgres"
	"checkin/test"
	"testing"

	_ "github.com/lib/pq" //Loads postgres as our database of choice
)

func hashFnGenerator(err error) func(string) (string, error) {
	return func(pwd string) (string, error) {
		if err != nil {
			return "", err
		}
		return pwd, nil
	}
}

func compareHashAndPasswordGenerator(t *testing.T, expectedHash string, expectedPassword string) func(string, string) bool {
	return func(hash string, pwd string) bool {
		if hash != expectedHash || pwd != expectedPassword {
			t.Fatal("Expected hash and password of " + expectedHash + " and " + expectedPassword + " respectively, but got " + hash + " and " + pwd)
		}
		return hash == pwd
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

	//event does not exist
	names, err = gs.Guests("notevenavaliduuidlol", []string{})
	test.Assert(t, err != nil, "No error thrown when event does not exist")

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

	//event does not exist
	names, err = gs.GuestsCheckedIn("notevenavaliduuidlol", []string{})
	test.Assert(t, err != nil, "No error thrown when event does not exist")
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

	//event does not exist
	names, err = gs.GuestsNotCheckedIn("notevenavaliduuidlol", []string{})
	test.Assert(t, err != nil, "No error thrown when event does not exist")
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
