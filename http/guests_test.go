package http_test

import (
	"checkin"
	myhttp "checkin/http"
	"checkin/mock"
	"checkin/test"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
		return nric == "1234F", nil
	}

	//Test normal behavior
	r := httptest.NewRequest("POST", "/api/v0/events/300/guests",
		strings.NewReader("{\"name\":\"Jim\", \"nric\":\"5678F\"}"))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusCreated, w.Result().StatusCode)

	//Test guest already exists with that nric
	r = httptest.NewRequest("POST", "/api/v0/events/300/guests",
		strings.NewReader("{\"name\":\"Jim\", \"nric\":\"1234F\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusConflict, w.Result().StatusCode)

	//Now the standard credential checks
	//Test access by another user
	auth.GetAuthInfoFn = func(r *http.Request) (checkin.AuthorizationInfo, error) {
		return checkin.AuthorizationInfo{
			Username: "unauthorized_person",
			IsAdmin:  false,
		}, nil
	}
	w = httptest.NewRecorder()
	r = httptest.NewRequest("POST", "/api/v0/events/300/guests",
		strings.NewReader("{\"name\":\"Jim\", \"nric\":\"5678F\"}"))
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
	r = httptest.NewRequest("POST", "/api/v0/events/300/guests",
		strings.NewReader("{\"name\":\"Jim\", \"nric\":\"5678F\"}"))
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusCreated, w.Result().StatusCode)

	//Test invalid eventID
	r = httptest.NewRequest("GET", "/api/v0/events/101/guests",
		strings.NewReader("{\"name\":\"Jim\", \"nric\":\"5678F\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusNotFound, w.Result().StatusCode)
}

func TestHandleRemoveGuest(t *testing.T) {
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
	gs.RemoveGuestFn = func(eventID string, nric string) error {
		test.Equals(t, "300", eventID)
		test.Equals(t, "5678F", nric)
		return nil
	}

	//Test normal behavior
	r := httptest.NewRequest("DELETE", "/api/v0/events/300/guests",
		strings.NewReader("{\"nric\":\"5678F\"}"))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusOK, w.Result().StatusCode)

	//Now the standard credentials checks
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
	r = httptest.NewRequest("DELETE", "/api/v0/events/300/guests",
		strings.NewReader("{\"nric\":\"5678F\"}"))
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusOK, w.Result().StatusCode)

	//Test invalid eventID
	r = httptest.NewRequest("GET", "/api/v0/events/101/guests",
		strings.NewReader("{\"nric\":\"5678F\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusNotFound, w.Result().StatusCode)
}

func TestHandleGuestsCheckedIn(t *testing.T) {
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
	gs.GuestsCheckedInFn = func(eventID string) ([]string, error) {
		if eventID != "100" {
			t.Fatalf("unexpected id: %s", eventID)
		}
		return []string{"Bob", "Jim", "Jacob"}, nil
	}

	r := httptest.NewRequest("GET", "/api/v0/events/100/guests/checkedin", nil)

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
	r = httptest.NewRequest("GET", "/api/v0/events/101/guests/checkedin", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusNotFound, w.Result().StatusCode)
}

func TestHandleCheckInGuest(t *testing.T) {
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
	gs.CheckInFn = func(eventID string, nric string) (string, error) {
		test.Equals(t, "300", eventID)
		test.Equals(t, "1234F", nric)
		return "Jim", nil
	}
	gs.GuestExistsFn = func(eventID string, nric string) (bool, error) {
		test.Equals(t, "300", eventID)
		return nric == "1234F", nil
	}

	//Test normal behavior
	r := httptest.NewRequest("POST", "/api/v0/events/300/guests/checkedin",
		strings.NewReader("{\"nric\":\"1234F\"}"))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	var name string
	json.NewDecoder(w.Result().Body).Decode(&name)
	test.Equals(t, "Jim", name)

	//Test guest does not exist with that nric
	r = httptest.NewRequest("POST", "/api/v0/events/300/guests/checkedin",
		strings.NewReader("{\"nric\":\"5678F\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusNotFound, w.Result().StatusCode)

	//Test invalid eventID
	r = httptest.NewRequest("POST", "/api/v0/events/200/guests/checkedin",
		strings.NewReader("{\"nric\":\"5678F\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusNotFound, w.Result().StatusCode)
}

func TestHandleGuestsNotCheckedIn(t *testing.T) {
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
	gs.GuestsNotCheckedInFn = func(eventID string) ([]string, error) {
		if eventID != "100" {
			t.Fatalf("unexpected id: %s", eventID)
		}
		return []string{"Bob", "Jim", "Jacob"}, nil
	}

	r := httptest.NewRequest("GET", "/api/v0/events/100/guests/notcheckedin", nil)

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
	r = httptest.NewRequest("GET", "/api/v0/events/101/guests/notcheckedin", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusNotFound, w.Result().StatusCode)
}

func TestHandleStats(t *testing.T) {
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
	gs.CheckInStatsFn = func(eventID string) (checkin.GuestStats, error) {
		if eventID != "100" {
			t.Fatalf("unexpected id: %s", eventID)
		}
		return checkin.GuestStats{
			TotalGuests:      10,
			CheckedIn:        5,
			PercentCheckedIn: 0.5,
		}, nil
	}

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
	json.NewDecoder(w.Result().Body).Decode(&stats)
	test.Equals(t, checkin.GuestStats{
		TotalGuests:      10,
		CheckedIn:        5,
		PercentCheckedIn: 0.5,
	}, stats)

	//Test invalid eventID
	r = httptest.NewRequest("GET", "/api/v0/events/1001/guests/stats", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusNotFound, w.Result().StatusCode)
}
