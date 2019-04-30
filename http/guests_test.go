package http_test

import (
	"checkin"
	myhttp "checkin/http"
	"checkin/mock"
	"checkin/test"
	"encoding/csv"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/guregu/null"
)

//Generates a HasConnection mock function (for use in mock.GuestMessenger) that returns the
//result argument, and also checks that it is passed a guestID that matches the expected
func hasConnectionGenerator(t *testing.T, expectedID string, result bool) func(string) bool {
	return func(guestID string) bool {
		test.Equals(t, expectedID, guestID)
		return result
	}
}

//Generates a Send mock function (for use in mock.GuestMessenger) that returns the error value
//provided. Also tests if the guestID and message passed in matches what was expected
func sendGenerator(t *testing.T, err error, expectedID string,
	expectedMsg myhttp.GuestMessage) func(string, myhttp.GuestMessage) error {
	return func(guestID string, msg myhttp.GuestMessage) error {
		test.Equals(t, expectedID, guestID)
		test.Equals(t, expectedMsg, msg)
		return err
	}
}

func TestHandleGuests(t *testing.T) {
	// Inject our mock into our handler.
	var gs mock.GuestService
	var es mock.EventService
	var auth mock.Authenticator
	var gm mock.GuestMessenger
	h := myhttp.NewGuestHandler(&gs, &es, &gm, &auth)

	//mock the required calls
	es.CheckIfExistsFn = checkIfExistsGenerator("100", nil)
	es.CheckHostFn = checkHostGenerator("testing_username", "100", nil)
	auth.AuthenticateFn = authenticateGenerator(true, nil)
	auth.GetAuthInfoFn = getAuthInfoGenerator("testing_username", false, nil)
	guestsGenerator := func(names []string, err error) func(string, []string) ([]string, error) {
		return func(eventID string, tags []string) ([]string, error) {
			if eventID != "100" {
				t.Fatalf("unexpected id: %s", eventID)
			}
			if tags != nil && len(tags) != 0 {
				t.Fatal("Expected nil or empty tags, got ", tags)
			}
			if err != nil {
				return nil, err
			}
			return names, nil
		}
	}
	gs.GuestsFn = guestsGenerator([]string{"Bob", "Jim", "Jacob"}, nil)

	//Test normal behavior
	r := httptest.NewRequest("GET", "/api/v0/events/100/guests", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	var guests []string
	json.NewDecoder(w.Result().Body).Decode(&guests)
	test.Equals(t, []string{"Bob", "Jim", "Jacob"}, guests)

	//Test no guests
	gs.GuestsFn = guestsGenerator([]string{}, nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	json.NewDecoder(w.Result().Body).Decode(&guests)
	test.Equals(t, []string{}, guests)

	//Test error getting guests
	gs.GuestsFn = guestsGenerator(nil, errors.New("An error"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	gs.GuestsFn = guestsGenerator([]string{"Bob", "Jim", "Jacob"}, nil)

	//access restriction tests
	//Test access by another user
	nonHostAccessTest(t, r, h, &auth, &es, "unauthorized_person")

	//Test access by admin
	adminAccessTest(t, r, h, &auth, func(r *http.Response) {
		json.NewDecoder(r.Body).Decode(&guests)
		test.Equals(t, []string{"Bob", "Jim", "Jacob"}, guests)
	})

	//Test invalid token
	noValidTokenTest(t, r, h, &auth)

	//Test invalid eventID
	r = httptest.NewRequest("GET", "/api/v0/events/200/guests", nil)
	eventDoesNotExistTest(t, r, h, &es)
}

func TestHandleRegisterGuest(t *testing.T) {
	// Inject our mock into our handler.
	var gs mock.GuestService
	var es mock.EventService
	var auth mock.Authenticator
	var gm mock.GuestMessenger
	h := myhttp.NewGuestHandler(&gs, &es, &gm, &auth)

	//mock the required calls
	es.CheckIfExistsFn = checkIfExistsGenerator("300", nil)
	es.CheckHostFn = checkHostGenerator("testing_username", "300", nil)
	auth.AuthenticateFn = authenticateGenerator(true, nil)
	auth.GetAuthInfoFn = getAuthInfoGenerator("testing_username", false, nil)
	registerGuestGenerator := func(err error) func(string, checkin.Guest) error {
		return func(eventID string, guest checkin.Guest) error {
			test.Equals(t, "300", eventID)
			test.Equals(t, "5678F", guest.NRIC)
			test.Equals(t, "Jim", guest.Name)
			test.Equals(t, []string(nil), guest.Tags)
			return err
		}
	}
	gs.RegisterGuestFn = registerGuestGenerator(nil)
	guestExistsGenerator := func(err error) func(string, string) (bool, error) {
		return func(eventID string, nric string) (bool, error) {
			test.Equals(t, "300", eventID)
			if err != nil {
				return false, err
			}
			return nric == "1234F", nil
		}
	}
	gs.GuestExistsFn = guestExistsGenerator(nil)

	//Test normal behavior
	r := httptest.NewRequest("POST", "/api/v0/events/300/guests",
		strings.NewReader("{\"name\":\"Jim\", \"nric\":\"5678F\"}"))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusCreated, w.Result().StatusCode)

	//Test guest already exists with that nric
	r = httptest.NewRequest("POST", "/api/v0/events/300/guests",
		strings.NewReader("{\"name\":\"Jim\", \"nric\":\"1234F\"}"))
	gs.RegisterGuestInvoked = false
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusConflict, w.Result().StatusCode)
	test.Assert(t, !gs.RegisterGuestInvoked, "Register guest invoked even though could not tell if guest exists")

	//Test error checking if guest exists
	r = httptest.NewRequest("POST", "/api/v0/events/300/guests",
		strings.NewReader("{\"name\":\"Jim\", \"nric\":\"5678F\"}"))
	gs.RegisterGuestInvoked = false
	gs.GuestExistsFn = guestExistsGenerator(errors.New("An error"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	test.Assert(t, !gs.RegisterGuestInvoked, "Register guest invoked even though guest already exists")
	gs.GuestExistsFn = guestExistsGenerator(nil)

	//Test error registering guest
	gs.RegisterGuestFn = registerGuestGenerator(errors.New("An error"))
	r = httptest.NewRequest("POST", "/api/v0/events/300/guests",
		strings.NewReader("{\"name\":\"Jim\", \"nric\":\"5678F\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	gs.RegisterGuestFn = registerGuestGenerator(nil)

	//Test invalid JSON
	r = httptest.NewRequest("POST", "/api/v0/events/300/guests",
		strings.NewReader("{\"name\":\"Jim\", nric\":\"5678F\""))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusBadRequest, w.Result().StatusCode)

	//Test extra field
	r = httptest.NewRequest("POST", "/api/v0/events/300/guests",
		strings.NewReader("{\"name\":\"Jim\", \"nric\":\"5678F\", \"nricHash\":\"5678F\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusBadRequest, w.Result().StatusCode)

	//access restriction tests
	//Test access by another user
	r = httptest.NewRequest("POST", "/api/v0/events/300/guests",
		strings.NewReader("{\"name\":\"Jim\", \"nric\":\"5678F\"}"))
	nonHostAccessTest(t, r, h, &auth, &es, "unauthorized_person")

	//Test access by admin
	r = httptest.NewRequest("POST", "/api/v0/events/300/guests",
		strings.NewReader("{\"name\":\"Jim\", \"nric\":\"5678F\"}"))
	adminAccessTest(t, r, h, &auth, func(r *http.Response) {
		test.Equals(t, http.StatusCreated, r.StatusCode)
	})

	//Test invalid token
	r = httptest.NewRequest("POST", "/api/v0/events/300/guests",
		strings.NewReader("{\"name\":\"Jim\", \"nric\":\"5678F\"}"))
	noValidTokenTest(t, r, h, &auth)

	//Test invalid eventID
	r = httptest.NewRequest("POST", "/api/v0/events/100/guests",
		strings.NewReader("{\"name\":\"Jim\", \"nric\":\"5678F\"}"))
	eventDoesNotExistTest(t, r, h, &es)
}

func TestHandleRemoveGuest(t *testing.T) {
	// Inject our mock into our handler.
	var gs mock.GuestService
	var es mock.EventService
	var auth mock.Authenticator
	var gm mock.GuestMessenger
	h := myhttp.NewGuestHandler(&gs, &es, &gm, &auth)

	//mock the required calls
	es.CheckIfExistsFn = checkIfExistsGenerator("300", nil)
	es.CheckHostFn = checkHostGenerator("testing_username", "300", nil)
	auth.AuthenticateFn = authenticateGenerator(true, nil)
	auth.GetAuthInfoFn = getAuthInfoGenerator("testing_username", false, nil)
	removeGuestGenerator := func(err error) func(string, string) error {
		return func(eventID string, nric string) error {
			test.Equals(t, "300", eventID)
			test.Equals(t, "5678F", nric)
			return err
		}
	}
	gs.RemoveGuestFn = removeGuestGenerator(nil)

	//Test normal behavior
	r := httptest.NewRequest("DELETE", "/api/v0/events/300/guests",
		strings.NewReader("{\"nric\":\"5678F\"}"))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusOK, w.Result().StatusCode)

	//Test badly formatted JSON
	r = httptest.NewRequest("DELETE", "/api/v0/events/300/guests",
		strings.NewReader(""))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusBadRequest, w.Result().StatusCode)

	//Test extra field (name supplied in particular)
	r = httptest.NewRequest("DELETE", "/api/v0/events/300/guests",
		strings.NewReader("{\"nric\":\"5678F\", \"field\":\"amazing\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusBadRequest, w.Result().StatusCode)
	r = httptest.NewRequest("DELETE", "/api/v0/events/300/guests",
		strings.NewReader("{\"nric\":\"5678F\", \"name\":\"amazing\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusBadRequest, w.Result().StatusCode)

	//Test error removing guest
	gs.RemoveGuestFn = removeGuestGenerator(errors.New("An error"))
	r = httptest.NewRequest("DELETE", "/api/v0/events/300/guests",
		strings.NewReader("{\"nric\":\"5678F\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	gs.RemoveGuestFn = removeGuestGenerator(nil)

	//access restriction tests
	//Test access by another user
	r = httptest.NewRequest("DELETE", "/api/v0/events/300/guests",
		strings.NewReader("{\"nric\":\"5678F\"}"))
	nonHostAccessTest(t, r, h, &auth, &es, "unauthorized_person")

	//Test access by admin
	r = httptest.NewRequest("DELETE", "/api/v0/events/300/guests",
		strings.NewReader("{\"nric\":\"5678F\"}"))
	adminAccessTest(t, r, h, &auth, func(r *http.Response) {
		test.Equals(t, http.StatusOK, r.StatusCode)
	})

	//Test invalid token
	r = httptest.NewRequest("DELETE", "/api/v0/events/300/guests",
		strings.NewReader("{\"nric\":\"5678F\"}"))
	noValidTokenTest(t, r, h, &auth)

	//Test invalid eventID
	r = httptest.NewRequest("DELETE", "/api/v0/events/200/guests",
		strings.NewReader("{\"nric\":\"5678F\"}"))
	eventDoesNotExistTest(t, r, h, &es)
}

func TestHandleGuestsCheckedIn(t *testing.T) {
	// Inject our mock into our handler.
	var gs mock.GuestService
	var es mock.EventService
	var auth mock.Authenticator
	var gm mock.GuestMessenger
	h := myhttp.NewGuestHandler(&gs, &es, &gm, &auth)

	//mock the required calls
	es.CheckIfExistsFn = checkIfExistsGenerator("100", nil)
	es.CheckHostFn = checkHostGenerator("testing_username", "100", nil)
	auth.AuthenticateFn = authenticateGenerator(true, nil)
	auth.GetAuthInfoFn = getAuthInfoGenerator("testing_username", false, nil)
	guestsCheckedInGenerator := func(names []string, err error) func(string, []string) ([]string, error) {
		return func(eventID string, tags []string) ([]string, error) {
			if eventID != "100" {
				t.Fatalf("unexpected id: %s", eventID)
			}
			if tags != nil && len(tags) != 0 {
				t.Fatal("Expected nil/empty tags but got", tags)
			}
			return names, err
		}
	}
	gs.GuestsCheckedInFn = guestsCheckedInGenerator([]string{"Bob", "Jim", "Jacob"}, nil)

	r := httptest.NewRequest("GET", "/api/v0/events/100/guests/checkedin", nil)

	//Test normal behavior
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	var guests []string
	json.NewDecoder(w.Result().Body).Decode(&guests)
	test.Equals(t, []string{"Bob", "Jim", "Jacob"}, guests)

	//Test no guests checked in
	gs.GuestsCheckedInFn = guestsCheckedInGenerator([]string{}, nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	json.NewDecoder(w.Result().Body).Decode(&guests)
	test.Equals(t, []string{}, guests)

	//Test error getting checked in guests
	gs.GuestsCheckedInFn = guestsCheckedInGenerator([]string{}, errors.New("An error"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	gs.GuestsCheckedInFn = guestsCheckedInGenerator([]string{"Bob", "Jim", "Jacob"}, nil)

	//access restriction tests
	//Test access by another user
	nonHostAccessTest(t, r, h, &auth, &es, "unauthorized_person")

	//Test access by admin
	adminAccessTest(t, r, h, &auth, func(r *http.Response) {
		json.NewDecoder(r.Body).Decode(&guests)
		test.Equals(t, []string{"Bob", "Jim", "Jacob"}, guests)
	})

	//Test invalid token
	noValidTokenTest(t, r, h, &auth)

	//Test invalid eventID
	r = httptest.NewRequest("GET", "/api/v0/events/200/guests/checkedin", nil)
	eventDoesNotExistTest(t, r, h, &es)
}

func TestHandleCheckInGuest(t *testing.T) {
	// Inject our mock into our handler.
	var gs mock.GuestService
	var es mock.EventService
	var auth mock.Authenticator
	var gm mock.GuestMessenger
	h := myhttp.NewGuestHandler(&gs, &es, &gm, &auth)

	es.CheckIfExistsFn = checkIfExistsGenerator("300", nil)
	//mock the required calls
	eventFnGenerator := func(offset time.Duration, valid bool, err error) func(string) (checkin.Event, error) {
		return func(ID string) (checkin.Event, error) {
			if ID != "300" {
				t.Fatalf("unexpected id: %s", ID)
			}

			if err != nil {
				return checkin.Event{}, err
			}
			if !valid {
				return checkin.Event{Release: null.Time{}}, nil
			}
			return checkin.Event{
				Release: null.Time{Time: time.Now().UTC().Add(offset),
					Valid: true,
				},
			}, nil

		}
	}
	es.EventFn = eventFnGenerator(-1*time.Hour, true, nil) //to meet release check
	checkInFnGenerator := func(err error) func(string, string) (string, error) {
		return func(eventID string, nric string) (string, error) {
			test.Equals(t, "300", eventID)
			test.Equals(t, "1234F", nric)
			if err != nil {
				return "", err
			}
			return "Jim", nil
		}
	}
	gs.CheckInFn = checkInFnGenerator(nil)
	guestExistsFnGenerator := func(err error) func(string, string) (bool, error) {
		return func(eventID string, nric string) (bool, error) {
			test.Equals(t, "300", eventID)
			if err != nil {
				return false, err
			}
			return nric == "1234F", nil
		}
	}
	gs.GuestExistsFn = guestExistsFnGenerator(nil)
	gm.HasConnectionFn = hasConnectionGenerator(t, "300 1234F", false)

	//Test normal behavior
	r := httptest.NewRequest("POST", "/api/v0/events/300/guests/checkedin",
		strings.NewReader("{\"nric\":\"1234F\"}"))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	var name string
	json.NewDecoder(w.Result().Body).Decode(&name)
	test.Equals(t, "Jim", name)

	//Test guest messenger active
	gm.HasConnectionFn = hasConnectionGenerator(t, "300 1234F", true)
	gm.SendFn = sendGenerator(t, nil, "300 1234F", myhttp.GuestMessage{
		Title: "checkedin/1",
		Content: checkin.Guest{
			Name: "Jim",
			NRIC: "1234F",
		},
	})
	r = httptest.NewRequest("POST", "/api/v0/events/300/guests/checkedin",
		strings.NewReader("{\"nric\":\"1234F\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	json.NewDecoder(w.Result().Body).Decode(&name)
	test.Equals(t, "Jim", name)

	//Test guest messenger fails to send message; execution should still complete
	gm.SendFn = sendGenerator(t, errors.New("An error"), "300 1234F", myhttp.GuestMessage{
		Title: "checkedin/1",
		Content: checkin.Guest{
			Name: "Jim",
			NRIC: "1234F",
		},
	})
	r = httptest.NewRequest("POST", "/api/v0/events/300/guests/checkedin",
		strings.NewReader("{\"nric\":\"1234F\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	json.NewDecoder(w.Result().Body).Decode(&name)
	test.Equals(t, "Jim", name)
	gm.HasConnectionFn = hasConnectionGenerator(t, "300 1234F", false)
	gm.SendFn = nil

	//Test guest does not exist with that nric
	gs.CheckInInvoked = false
	r = httptest.NewRequest("POST", "/api/v0/events/300/guests/checkedin",
		strings.NewReader("{\"nric\":\"5678F\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusNotFound, w.Result().StatusCode)
	test.Assert(t, !gs.CheckInInvoked, "Check-in was invoked even though guest did not exist")

	//Test empty body
	r = httptest.NewRequest("POST", "/api/v0/events/300/guests/checkedin",
		strings.NewReader(""))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusBadRequest, w.Result().StatusCode)

	//Test extra field (name supplied in particular)
	r = httptest.NewRequest("POST", "/api/v0/events/300/guests/checkedin",
		strings.NewReader("{\"nric\":\"5678F\", \"field\":\"amazing\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusBadRequest, w.Result().StatusCode)
	r = httptest.NewRequest("POST", "/api/v0/events/300/guests/checkedin",
		strings.NewReader("{\"nric\":\"5678F\", \"name\":\"amazing\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusBadRequest, w.Result().StatusCode)

	//Test error checking if guest exists
	gs.GuestExistsFn = guestExistsFnGenerator(errors.New("An error"))
	gs.CheckInInvoked = false
	r = httptest.NewRequest("POST", "/api/v0/events/300/guests/checkedin",
		strings.NewReader("{\"nric\":\"1234F\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	test.Assert(t, !gs.CheckInInvoked, "Check-in was invoked even though guest existance not confirmed")
	gs.GuestExistsFn = guestExistsFnGenerator(nil)

	//Test error checking in guest
	gs.CheckInFn = checkInFnGenerator(errors.New("An error"))
	r = httptest.NewRequest("POST", "/api/v0/events/300/guests/checkedin",
		strings.NewReader("{\"nric\":\"1234F\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	gs.CheckInFn = checkInFnGenerator(nil)

	//Test invalid eventID
	r = httptest.NewRequest("POST", "/api/v0/events/200/guests/checkedin",
		strings.NewReader("{\"nric\":\"5678F\"}"))
	eventDoesNotExistTest(t, r, h, &es)

	//Test not yet released
	es.EventFn = eventFnGenerator(time.Hour, true, nil)
	r = httptest.NewRequest("POST", "/api/v0/events/300/guests/checkedin",
		strings.NewReader("{\"nric\":\"1234F\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusForbidden, w.Result().StatusCode)

	//Test no release date set
	es.EventFn = eventFnGenerator(0, false, nil)
	r = httptest.NewRequest("POST", "/api/v0/events/300/guests/checkedin",
		strings.NewReader("{\"nric\":\"1234F\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	json.NewDecoder(w.Result().Body).Decode(&name)
	test.Equals(t, "Jim", name)

	//Test error getting release date
	es.EventFn = eventFnGenerator(time.Hour, false, errors.New("An error"))
	r = httptest.NewRequest("POST", "/api/v0/events/300/guests/checkedin",
		strings.NewReader("{\"nric\":\"1234F\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)

}

func TestHandleMarkGuestAbsent(t *testing.T) {
	// Inject our mock into our handler.
	var gs mock.GuestService
	var es mock.EventService
	var auth mock.Authenticator
	var gm mock.GuestMessenger
	h := myhttp.NewGuestHandler(&gs, &es, &gm, &auth)

	//mock the required calls
	es.CheckIfExistsFn = checkIfExistsGenerator("300", nil)
	es.CheckHostFn = checkHostGenerator("testing_username", "300", nil)
	auth.AuthenticateFn = authenticateGenerator(true, nil)
	auth.GetAuthInfoFn = getAuthInfoGenerator("testing_username", false, nil)
	markAbsentFnGenerator := func(err error) func(string, string) error {
		return func(eventID string, nric string) error {
			test.Equals(t, "300", eventID)
			test.Equals(t, "1234F", nric)
			return err
		}
	}
	gs.MarkAbsentFn = markAbsentFnGenerator(nil)
	guestExistsFnGenerator := func(err error) func(string, string) (bool, error) {
		return func(eventID string, nric string) (bool, error) {
			test.Equals(t, "300", eventID)
			if err != nil {
				return false, err
			}
			return nric == "1234F", nil
		}
	}
	gs.GuestExistsFn = guestExistsFnGenerator(nil)
	gm.HasConnectionFn = hasConnectionGenerator(t, "300 1234F", false)

	//Test normal behavior
	r := httptest.NewRequest("DELETE", "/api/v0/events/300/guests/checkedin",
		strings.NewReader("{\"nric\":\"1234F\"}"))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusOK, w.Result().StatusCode)

	//Test listener active/need to send guest message
	gm.HasConnectionFn = hasConnectionGenerator(t, "300 1234F", true)
	gm.SendFn = sendGenerator(t, nil, "300 1234F", myhttp.GuestMessage{
		Title: "checkedin/0",
	})
	r = httptest.NewRequest("DELETE", "/api/v0/events/300/guests/checkedin",
		strings.NewReader("{\"nric\":\"1234F\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusOK, w.Result().StatusCode)

	//Test sending message to guest throws error; execution should still complete
	gm.SendFn = sendGenerator(t, errors.New("An error"), "300 1234F", myhttp.GuestMessage{
		Title: "checkedin/0",
	})
	r = httptest.NewRequest("DELETE", "/api/v0/events/300/guests/checkedin",
		strings.NewReader("{\"nric\":\"1234F\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusOK, w.Result().StatusCode)
	gm.HasConnectionFn = hasConnectionGenerator(t, "300 1234F", false)
	gm.SendFn = nil

	//Test guest does not exist with that nric
	r = httptest.NewRequest("DELETE", "/api/v0/events/300/guests/checkedin",
		strings.NewReader("{\"nric\":\"5678F\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusNotFound, w.Result().StatusCode)

	//Test badly formatted JSON
	r = httptest.NewRequest("DELETE", "/api/v0/events/300/guests/checkedin",
		strings.NewReader("{\"nric\":\"1234F\""))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusBadRequest, w.Result().StatusCode)

	//Test extra field (name supplied in particular)
	r = httptest.NewRequest("DELETE", "/api/v0/events/300/guests/checkedin",
		strings.NewReader("{\"nric\":\"5678F\", \"field\":\"amazing\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusBadRequest, w.Result().StatusCode)
	r = httptest.NewRequest("DELETE", "/api/v0/events/300/guests/checkedin",
		strings.NewReader("{\"nric\":\"5678F\", \"name\":\"amazing\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusBadRequest, w.Result().StatusCode)

	//Test error when marking as absent
	gs.MarkAbsentFn = markAbsentFnGenerator(errors.New("An error"))
	r = httptest.NewRequest("DELETE", "/api/v0/events/300/guests/checkedin",
		strings.NewReader("{\"nric\":\"1234F\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	gs.MarkAbsentFn = markAbsentFnGenerator(nil)

	//Test error when checking if guest exists
	gs.GuestExistsFn = guestExistsFnGenerator(errors.New("An error"))
	r = httptest.NewRequest("DELETE", "/api/v0/events/300/guests/checkedin",
		strings.NewReader("{\"nric\":\"1234F\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	gs.GuestExistsFn = guestExistsFnGenerator(nil)

	//access restriction tests
	//Test access by another user
	r = httptest.NewRequest("DELETE", "/api/v0/events/300/guests/checkedin",
		strings.NewReader("{\"nric\":\"1234F\"}"))
	nonHostAccessTest(t, r, h, &auth, &es, "unauthorized_person")

	//Test access by admin
	r = httptest.NewRequest("DELETE", "/api/v0/events/300/guests/checkedin",
		strings.NewReader("{\"nric\":\"1234F\"}"))
	adminAccessTest(t, r, h, &auth, func(r *http.Response) {
		test.Equals(t, http.StatusOK, r.StatusCode)
	})

	//Test invalid token
	r = httptest.NewRequest("DELETE", "/api/v0/events/300/guests/checkedin",
		strings.NewReader("{\"nric\":\"1234F\"}"))
	noValidTokenTest(t, r, h, &auth)

	//Test invalid eventID
	r = httptest.NewRequest("DELETE", "/api/v0/events/200/guests/checkedin",
		strings.NewReader("{\"nric\":\"1234F\"}"))
	eventDoesNotExistTest(t, r, h, &es)
}

func TestHandleCreateCheckInListener(t *testing.T) {
	// Inject our mock into our handler.
	var gs mock.GuestService
	var es mock.EventService
	var auth mock.Authenticator
	var gm mock.GuestMessenger
	h := myhttp.NewGuestHandler(&gs, &es, &gm, &auth)

	es.CheckIfExistsFn = checkIfExistsGenerator("300", nil)
	openConnectionGen := func(err error) func(string, http.ResponseWriter, *http.Request) error {
		return func(guestID string, w http.ResponseWriter, r *http.Request) error {
			test.Equals(t, "300 1234F", guestID)
			return err
		}
	}
	gm.OpenConnectionFn = openConnectionGen(nil)

	//Test normal behavior
	r := httptest.NewRequest("GET",
		"/api/v1-2/events/300/guests/checkedin/listener/1234F", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusOK, w.Result().StatusCode)

	//Test open connection fails
	gm.OpenConnectionFn = openConnectionGen(errors.New("An error"))
	r = httptest.NewRequest("GET",
		"/api/v1-2/events/300/guests/checkedin/listener/1234F", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	gm.OpenConnectionFn = openConnectionGen(nil)

	//Test invalid eventID
	r = httptest.NewRequest("GET",
		"/api/v1-2/events/200/guests/checkedin/listener/1234F", nil)
	eventDoesNotExistTest(t, r, h, &es)
}

func TestHandleGuestsNotCheckedIn(t *testing.T) {
	// Inject our mock into our handler.
	var gs mock.GuestService
	var es mock.EventService
	var auth mock.Authenticator
	var gm mock.GuestMessenger
	h := myhttp.NewGuestHandler(&gs, &es, &gm, &auth)

	//mock the required calls
	es.CheckIfExistsFn = checkIfExistsGenerator("100", nil)
	es.CheckHostFn = checkHostGenerator("testing_username", "100", nil)
	auth.AuthenticateFn = authenticateGenerator(true, nil)
	auth.GetAuthInfoFn = getAuthInfoGenerator("testing_username", false, nil)
	guestsNotCheckedInFnGenerator := func(names []string, err error) func(string, []string) ([]string, error) {
		return func(eventID string, tags []string) ([]string, error) {
			if eventID != "100" {
				t.Fatalf("unexpected id: %s", eventID)
			}
			if tags != nil && len(tags) != 0 {
				t.Fatal("Expected nil or empty tags, but got ", tags)
			}
			return names, err
		}
	}
	gs.GuestsNotCheckedInFn = guestsNotCheckedInFnGenerator([]string{"Bob", "Jim", "Jacob"}, nil)

	r := httptest.NewRequest("GET", "/api/v0/events/100/guests/notcheckedin", nil)

	//Test normal behavior
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	var guests []string
	json.NewDecoder(w.Result().Body).Decode(&guests)
	test.Equals(t, []string{"Bob", "Jim", "Jacob"}, guests)

	//Test nobody not checked in
	gs.GuestsNotCheckedInFn = guestsNotCheckedInFnGenerator([]string{}, nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	json.NewDecoder(w.Result().Body).Decode(&guests)
	test.Equals(t, []string{}, guests)
	gs.GuestsNotCheckedInFn = guestsNotCheckedInFnGenerator([]string{"Bob", "Jim", "Jacob"}, nil)

	//Test error getting those who have not checked in
	gs.GuestsNotCheckedInFn = guestsNotCheckedInFnGenerator([]string{}, errors.New("An error"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	gs.GuestsNotCheckedInFn = guestsNotCheckedInFnGenerator([]string{"Bob", "Jim", "Jacob"}, nil)

	//access restriction tests
	//Test access by another user
	nonHostAccessTest(t, r, h, &auth, &es, "unauthorized_person")

	//Test access by admin
	adminAccessTest(t, r, h, &auth, func(r *http.Response) {
		json.NewDecoder(r.Body).Decode(&guests)
		test.Equals(t, []string{"Bob", "Jim", "Jacob"}, guests)
	})

	//Test invalid token
	noValidTokenTest(t, r, h, &auth)

	//Test invalid eventID
	r = httptest.NewRequest("GET", "/api/v0/events/101/guests/notcheckedin", nil)
	eventDoesNotExistTest(t, r, h, &es)
}

func TestHandleStats(t *testing.T) {
	// Inject our mock into our handler.
	var gs mock.GuestService
	var es mock.EventService
	var auth mock.Authenticator
	var gm mock.GuestMessenger
	h := myhttp.NewGuestHandler(&gs, &es, &gm, &auth)

	//mock the required calls
	es.CheckIfExistsFn = checkIfExistsGenerator("100", nil)
	es.CheckHostFn = checkHostGenerator("testing_username", "100", nil)
	auth.AuthenticateFn = authenticateGenerator(true, nil)
	auth.GetAuthInfoFn = getAuthInfoGenerator("testing_username", false, nil)
	checkInStatsFnGenerator := func(err error) func(string, []string) (checkin.GuestStats, error) {
		return func(eventID string, tags []string) (checkin.GuestStats, error) {
			if eventID != "100" {
				t.Fatalf("unexpected id: %s", eventID)
			}
			if tags != nil && len(tags) != 0 {
				t.Fatal("Expected nil or empty tags but got ", tags)
			}
			if err != nil {
				return checkin.GuestStats{}, err
			}
			return checkin.GuestStats{
				TotalGuests:      10,
				CheckedIn:        5,
				PercentCheckedIn: 0.5,
			}, nil
		}
	}
	gs.CheckInStatsFn = checkInStatsFnGenerator(nil)

	r := httptest.NewRequest("GET", "/api/v0/events/100/guests/stats", nil)

	//Test normal behavior
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	var stats checkin.GuestStats
	json.NewDecoder(w.Result().Body).Decode(&stats)
	test.Equals(t, checkin.GuestStats{
		TotalGuests:      10,
		CheckedIn:        5,
		PercentCheckedIn: 0.5,
	}, stats)

	//Test error getting stats
	gs.CheckInStatsFn = checkInStatsFnGenerator(errors.New("An error"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	gs.CheckInStatsFn = checkInStatsFnGenerator(nil)

	//access restriction tests
	//Test access by another user
	nonHostAccessTest(t, r, h, &auth, &es, "unauthorized_person")

	//Test access by admin
	adminAccessTest(t, r, h, &auth, func(r *http.Response) {
		json.NewDecoder(r.Body).Decode(&stats)
		test.Equals(t, checkin.GuestStats{
			TotalGuests:      10,
			CheckedIn:        5,
			PercentCheckedIn: 0.5,
		}, stats)
	})

	//Test invalid token
	noValidTokenTest(t, r, h, &auth)

	//Test invalid eventID
	r = httptest.NewRequest("GET", "/api/v0/events/1001/guests/stats", nil)
	eventDoesNotExistTest(t, r, h, &es)
}

func TestHandleReport(t *testing.T) {
	// Inject our mock into our handler.
	var gs mock.GuestService
	var es mock.EventService
	var auth mock.Authenticator
	var gm mock.GuestMessenger
	h := myhttp.NewGuestHandler(&gs, &es, &gm, &auth)

	es.CheckIfExistsFn = checkIfExistsGenerator("100", nil)
	es.CheckHostFn = checkHostGenerator("testing_username", "100", nil)
	auth.AuthenticateFn = authenticateGenerator(true, nil)
	auth.GetAuthInfoFn = getAuthInfoGenerator("testing_username", false, nil)
	guestsCheckedInGenerator := func(names []string, err error) func(string, []string) ([]string, error) {
		return func(eventID string, tags []string) ([]string, error) {
			if eventID != "100" {
				t.Fatalf("unexpected id: %s", eventID)
			}
			if tags != nil && len(tags) != 0 {
				t.Fatal("Expected nil or empty tags but got ", tags)
			}
			return names, err
		}
	}
	gs.GuestsCheckedInFn = guestsCheckedInGenerator([]string{"Alice", "Jim", "Bob"}, nil)
	guestsNotCheckedInFnGenerator := func(names []string, err error) func(string, []string) ([]string, error) {
		return func(eventID string, tags []string) ([]string, error) {
			if eventID != "100" {
				t.Fatalf("unexpected id: %s", eventID)
			}
			if tags != nil && len(tags) != 0 {
				t.Fatal("Expected nil or empty tags but got ", tags)
			}
			return names, err
		}
	}
	gs.GuestsNotCheckedInFn = guestsNotCheckedInFnGenerator([]string{"Herman", "Ritchie"}, nil)

	r := httptest.NewRequest("GET", "/api/v0/events/100/guests/report", nil)

	//Test normal behavior
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	reader := csv.NewReader(w.Result().Body)
	data, err := reader.ReadAll()
	test.Ok(t, err)
	for _, row := range data {
		if row[0] == "Alice" || row[0] == "Jim" || row[0] == "Bob" {
			test.Equals(t, "1", row[1])
		} else if row[0] == "Herman" || row[0] == "Ritchie" {
			test.Equals(t, "0", row[1])
		} else {
			test.Equals(t, row[0], "Name")
		}
	}

	//check empty lists
	gs.GuestsCheckedInFn = guestsCheckedInGenerator([]string{}, nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	reader = csv.NewReader(w.Result().Body)
	data, err = reader.ReadAll()
	test.Ok(t, err)
	for _, row := range data {
		if row[0] != "Name" {
			test.Equals(t, "0", row[1])
		}
	}
	gs.GuestsCheckedInFn = guestsCheckedInGenerator([]string{"Alice", "Jim", "Bob"}, nil)

	gs.GuestsNotCheckedInFn = guestsNotCheckedInFnGenerator([]string{}, nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	reader = csv.NewReader(w.Result().Body)
	data, err = reader.ReadAll()
	test.Ok(t, err)
	for _, row := range data {
		if row[0] != "Name" {
			test.Equals(t, "1", row[1])
		}
	}

	gs.GuestsNotCheckedInFn = guestsNotCheckedInFnGenerator([]string{}, nil)
	gs.GuestsCheckedInFn = guestsCheckedInGenerator([]string{}, nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	reader = csv.NewReader(w.Result().Body)
	data, err = reader.ReadAll()
	test.Ok(t, err)
	test.Equals(t, 1, len(data))
	test.Equals(t, "Name", data[0][0])
	test.Equals(t, "Present", data[0][1])
	gs.GuestsNotCheckedInFn = guestsNotCheckedInFnGenerator([]string{"Herman", "Ritchie"}, nil)
	gs.GuestsCheckedInFn = guestsCheckedInGenerator([]string{"Alice", "Jim", "Bob"}, nil)

	//check internal server error handling
	gs.GuestsCheckedInFn = guestsCheckedInGenerator([]string{}, errors.New("An error"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	gs.GuestsCheckedInFn = guestsCheckedInGenerator([]string{"Alice", "Jim", "Bob"}, nil)

	gs.GuestsNotCheckedInFn = guestsNotCheckedInFnGenerator([]string{}, errors.New("An error"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	gs.GuestsNotCheckedInFn = guestsNotCheckedInFnGenerator([]string{"Herman", "Ritchie"}, nil)

	//access restriction tests
	//Test access by another user
	nonHostAccessTest(t, r, h, &auth, &es, "unauthorized_person")

	//Test access by admin
	adminAccessTest(t, r, h, &auth, func(r *http.Response) {
		reader := csv.NewReader(r.Body)
		data, err := reader.ReadAll()
		test.Ok(t, err)
		for _, row := range data {
			if row[0] == "Alice" || row[0] == "Jim" || row[0] == "Bob" {
				test.Equals(t, "1", row[1])
			} else if row[0] == "Herman" || row[0] == "Ritchie" {
				test.Equals(t, "0", row[1])
			} else {
				test.Equals(t, row[0], "Name")
			}
		}
	})

	//Test invalid token
	noValidTokenTest(t, r, h, &auth)

	//Test invalid eventID
	r = httptest.NewRequest("GET", "/api/v0/events/1001/guests/report", nil)
	eventDoesNotExistTest(t, r, h, &es)
}
