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
	loc, err := time.LoadLocation("Asia/Singapore")
	test.Ok(t, err)

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
			Release: null.Time{Time: time.Now().In(loc).Add(-1 * time.Hour),
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
			Release: null.Time{Time: time.Now().In(loc).Add(1 * time.Hour),
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
			Release: null.Time{Time: time.Now().In(loc).Add(-1 * time.Hour),
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
