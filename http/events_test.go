package http_test

//TODO: Complete tests for all event handler routes

import (
	"checkin"
	myhttp "checkin/http"
	"checkin/mock"
	"checkin/test"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/guregu/null"
)

func TestHandleReleased(t *testing.T) {
	var es mock.EventService
	var auth mock.Authenticator
	gh := myhttp.GuestHandler{}
	h := myhttp.NewEventHandler(&es, &auth, &gh)

	es.CheckIfExistsFn = func(id string) (bool, error) {
		if id == "300" {
			return true, nil
		}
		return false, nil
	}
	es.EventFn = func(ID string) (checkin.Event, error) {
		if ID != "300" {
			t.Fatalf("unexpected id: %s", ID)
		}
		return checkin.Event{
			Release: null.Time{Time: time.Now().UTC().Add(-1 * time.Hour),
				Valid: true,
			},
		}, nil
	}

	//test normal functionality

	//test event released 1 hour ago
	r := httptest.NewRequest("GET", "/api/v0/events/300/released", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	var release bool
	json.NewDecoder(w.Result().Body).Decode(&release)
	test.Equals(t, true, release)

	//test event released 1 hour in the future
	es.EventFn = func(ID string) (checkin.Event, error) {
		if ID != "300" {
			t.Fatalf("unexpected id: %s", ID)
		}
		return checkin.Event{
			Release: null.Time{Time: time.Now().UTC().Add(1 * time.Hour),
				Valid: true,
			},
		}, nil
	}
	r = httptest.NewRequest("GET", "/api/v0/events/300/released", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	json.NewDecoder(w.Result().Body).Decode(&release)
	test.Equals(t, false, release)

	//test no release time set
	es.EventFn = func(ID string) (checkin.Event, error) {
		if ID != "300" {
			t.Fatalf("unexpected id: %s", ID)
		}
		return checkin.Event{
			Release: null.Time{Time: time.Time{},
				Valid: false,
			},
		}, nil
	}
	r = httptest.NewRequest("GET", "/api/v0/events/300/released", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	json.NewDecoder(w.Result().Body).Decode(&release)
	test.Equals(t, true, release)

	//test error in getting event
	es.EventFn = func(ID string) (checkin.Event, error) {
		if ID != "300" {
			t.Fatalf("unexpected id: %s", ID)
		}
		return checkin.Event{}, errors.New("An error")
	}
	r = httptest.NewRequest("GET", "/api/v0/events/300/released", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	es.EventFn = func(ID string) (checkin.Event, error) {
		if ID != "300" {
			t.Fatalf("unexpected id: %s", ID)
		}
		return checkin.Event{
			Release: null.Time{Time: time.Now().UTC().Add(-1 * time.Hour),
				Valid: true,
			},
		}, nil
	}

	//test event does not exist
	r = httptest.NewRequest("GET", "/api/v0/events/200/released", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusNotFound, w.Result().StatusCode)
}

func TestHandleEventsBy(t *testing.T) {
	var es mock.EventService
	var auth mock.Authenticator
	gh := myhttp.GuestHandler{}
	h := myhttp.NewEventHandler(&es, &auth, &gh)

	auth.AuthenticateFn = func(r *http.Request) (bool, error) {
		return true, nil
	}
	auth.GetAuthInfoFn = func(r *http.Request) (checkin.AuthorizationInfo, error) {
		return checkin.AuthorizationInfo{
			Username: "testing_username",
			IsAdmin:  false,
		}, nil
	}
	es.EventsByFn = func(username string) ([]checkin.Event, error) {
		if username != "testing_username" {
			t.Fatal("Unexpected username: " + username + ", expected testing_username")
		}
		return []checkin.Event{checkin.Event{ID: "100"}, checkin.Event{ID: "200"}, checkin.Event{ID: "300"}}, nil
	}

	//Test normal behavior
	r := httptest.NewRequest("GET", "/api/v0/events", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	var events []checkin.Event
	json.NewDecoder(w.Result().Body).Decode(&events)
	test.Equals(t, []checkin.Event{checkin.Event{ID: "100"},
		checkin.Event{ID: "200"},
		checkin.Event{ID: "300"}}, events)

	//Test getting auth info fails
	auth.GetAuthInfoFn = func(r *http.Request) (checkin.AuthorizationInfo, error) {
		return checkin.AuthorizationInfo{}, errors.New("An error")
	}
	r = httptest.NewRequest("GET", "/api/v0/events", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	auth.GetAuthInfoFn = func(r *http.Request) (checkin.AuthorizationInfo, error) {
		return checkin.AuthorizationInfo{
			Username: "testing_username",
			IsAdmin:  false,
		}, nil
	}

	//test error fetching events
	es.EventsByFn = func(username string) ([]checkin.Event, error) {
		if username != "testing_username" {
			t.Fatal("Unexpected username: " + username + ", expected testing_username")
		}
		return nil, errors.New("An error")
	}
	r = httptest.NewRequest("GET", "/api/v0/events", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	es.EventsByFn = func(username string) ([]checkin.Event, error) {
		if username != "testing_username" {
			t.Fatal("Unexpected username: " + username + ", expected testing_username")
		}
		return []checkin.Event{checkin.Event{ID: "100"}, checkin.Event{ID: "200"}, checkin.Event{ID: "300"}}, nil
	}

	//test no valid token
	auth.AuthenticateFn = func(r *http.Request) (bool, error) {
		return false, nil
	}
	r = httptest.NewRequest("GET", "/api/v0/events", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusUnauthorized, w.Result().StatusCode)
}

func TestHandleEvent(t *testing.T) {
	var es mock.EventService
	var auth mock.Authenticator
	gh := myhttp.GuestHandler{}
	h := myhttp.NewEventHandler(&es, &auth, &gh)

	es.CheckIfExistsFn = func(id string) (bool, error) {
		if id == "300" {
			return true, nil
		}
		return false, nil
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
	es.EventFn = func(ID string) (checkin.Event, error) {
		if ID != "300" {
			t.Fatal("Unexpected username: " + ID + ", expected 300")
		}
		return checkin.Event{ID: "300"}, nil
	}

	//Test normal behavior
	r := httptest.NewRequest("GET", "/api/v0/events/300", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	var event checkin.Event
	json.NewDecoder(w.Result().Body).Decode(&event)
	test.Equals(t, checkin.Event{ID: "300"}, event)

	//Test error fetching event
	es.EventFn = func(ID string) (checkin.Event, error) {
		if ID != "300" {
			t.Fatal("Unexpected username: " + ID + ", expected 300")
		}
		return checkin.Event{}, errors.New("An error")
	}
	r = httptest.NewRequest("GET", "/api/v0/events/300", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	es.EventFn = func(ID string) (checkin.Event, error) {
		if ID != "300" {
			t.Fatal("Unexpected username: " + ID + ", expected 300")
		}
		return checkin.Event{ID: "300"}, nil
	}

	//Test access by another user
	auth.GetAuthInfoFn = func(r *http.Request) (checkin.AuthorizationInfo, error) {
		return checkin.AuthorizationInfo{
			Username: "unauthorized_person",
			IsAdmin:  false,
		}, nil
	}
	r = httptest.NewRequest("GET", "/api/v0/events/300", nil)
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
	r = httptest.NewRequest("GET", "/api/v0/events/300", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	json.NewDecoder(w.Result().Body).Decode(&event)
	test.Equals(t, checkin.Event{ID: "300"}, event)

	//Test invalid eventID
	r = httptest.NewRequest("GET", "/api/v0/events/100", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusNotFound, w.Result().StatusCode)

	//Test invalid token
	auth.AuthenticateFn = func(r *http.Request) (bool, error) {
		return false, nil
	}
	r = httptest.NewRequest("GET", "/api/v0/events/300", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusUnauthorized, w.Result().StatusCode)

}
