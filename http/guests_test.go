package http_test

import (
	"checkin"
	myhttp "checkin/http"
	"checkin/mock"
	"checkin/test"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"strings"
)

func TestHandleGuests(t *testing.T) {
	// Inject our mock into our handler.
	var gs mock.GuestService
	var es mock.EventService
	var auth mock.Authenticator
	h := myhttp.NewGuestHandler(&gs, &es, &auth)

	//mock the required calls
	es.CheckIfExistsFn = func(eventID string) (bool, error) {
		return eventID == "100", nil
	}
	es.CheckHostFn = func(username string, eventID string) (bool, error) {
		if username != "testing_username" {
			return false, nil
		} else if eventID != "100" {
			return false, nil
		}
		return true, nil
	}
	auth.AuthenticateFn = func(r *http.Request) (bool, error) {
		return true, nil
	}
	auth.GetAuthInfoFn = func(r *http.Request) (checkin.AuthorizationInfo, error) {
		return checkin.AuthorizationInfo{
			Username: "testing_username",
			IsAdmin:  false,
		}, nil
	}
	gs.GuestsFn = func(eventID string) ([]string, error) {
		if eventID != "100" {
			t.Fatalf("unexpected id: %s", eventID)
		}
		return []string{"Bob", "Jim", "Jacob"}, nil
	}

	r := httptest.NewRequest("GET", "/api/v0/events/100/guests", nil)

	//Test normal behavior
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	var guests []string
	json.NewDecoder(w.Result().Body).Decode(&guests)
	test.Equals(t, []string{"Bob", "Jim", "Jacob"}, guests)

	//Test access by another user
	auth.GetAuthInfoFn = func(r *http.Request) (checkin.AuthorizationInfo, error) {
		return checkin.AuthorizationInfo{
			Username: "unauthorized_person",
			IsAdmin:  false,
		}, nil
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusForbidden, w.Result().StatusCode)

	//Test access by admin
	auth.GetAuthInfoFn = func(r *http.Request) (checkin.AuthorizationInfo, error) {
		return checkin.AuthorizationInfo{
			Username: "admin_person",
			IsAdmin:  true,
		}, nil
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	json.NewDecoder(w.Result().Body).Decode(&guests)
	test.Equals(t, []string{"Bob", "Jim", "Jacob"}, guests)

	//Test invalid eventID
	r = httptest.NewRequest("GET", "/api/v0/events/101/guests", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusNotFound, w.Result().StatusCode)
}

func TestHandleRegisterGuest(t *testing.T) {
	// Inject our mock into our handler.
	var gs mock.GuestService
	var es mock.EventService
	var auth mock.Authenticator
	h := myhttp.NewGuestHandler(&gs, &es, &auth)

	//mock the required calls
	es.CheckIfExistsFn = func(eventID string) (bool, error) {
		return eventID == "300", nil
	}
	es.CheckHostFn = func(username string, eventID string) (bool, error) {
		if username != "testing_username" {
			return false, nil
		} else if eventID != "300" {
			return false, nil
		}
		return true, nil
	}
	auth.AuthenticateFn = func(r *http.Request) (bool, error) {
		return true, nil
	}
	auth.GetAuthInfoFn = func(r *http.Request) (checkin.AuthorizationInfo, error) {
		return checkin.AuthorizationInfo{
			Username: "testing_username",
			IsAdmin:  false,
		}, nil
	}
	gs.RegisterGuestFn = func(eventID string, nric string, name string) error {
		test.Equals(t, "300", eventID)
		test.Equals(t, "5678F", nric)
		test.Equals(t, "Jim", name)
		return nil
	}
	gs.GuestExistsFn = func(eventID string, nric string) (bool, error) {
		test.Equals(t, "300", eventID)
		return nric=="1234F", nil
	}

	//Test normal behavior
	r := httptest.NewRequest("POST", "/api/v0/events/300/guests", 
		strings.NewReader("{\"name\":\"Jim\", \"nric\":\"5678F\"}"))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusCreated,w.Result().StatusCode)

	//Test guest already exists with that nric
	r = httptest.NewRequest("POST", "/api/v0/events/300/guests", 
		strings.NewReader("{\"name\":\"Jim\", \"nric\":\"1234F\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusConflict,w.Result().StatusCode)

	//Now the standard credential checks
	r = httptest.NewRequest("POST", "/api/v0/events/300/guests", 
		strings.NewReader("{\"name\":\"Jim\", \"nric\":\"5678F\"}"))
	//Test access by another user
	auth.GetAuthInfoFn = func(r *http.Request) (checkin.AuthorizationInfo, error) {
		return checkin.AuthorizationInfo{
			Username: "unauthorized_person",
			IsAdmin:  false,
		}, nil
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusForbidden, w.Result().StatusCode)

	//Test access by admin
	auth.GetAuthInfoFn = func(r *http.Request) (checkin.AuthorizationInfo, error) {
		return checkin.AuthorizationInfo{
			Username: "admin_person",
			IsAdmin:  true,
		}, nil
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusCreated,w.Result().StatusCode)

	//Test invalid eventID
	r = httptest.NewRequest("GET", "/api/v0/events/101/guests", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusNotFound, w.Result().StatusCode)
}
