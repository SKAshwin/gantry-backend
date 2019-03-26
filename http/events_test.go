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

func checkIfExistsGenerator(expectedID string, err error) func(string) (bool, error) {
	return func(id string) (bool, error) {
		if err != nil {
			return false, err
		}

		return id == expectedID, nil
	}
}

func checkHostGenerator(expectedUsername string, expectedID string, err error) func(string, string) (bool, error) {
	return func(username string, eventID string) (bool, error) {
		if err != nil {
			return false, err
		}

		if username != expectedUsername {
			return false, nil
		} else if eventID != expectedID {
			return false, nil
		}
		return true, nil
	}
}

func authenticateGenerator(authenticate bool, err error) func(r *http.Request) (bool, error) {
	return func(r *http.Request) (bool, error) {
		return authenticate, err
	}
}

func getAuthInfoGenerator(username string, admin bool, err error) func(r *http.Request) (checkin.AuthorizationInfo, error) {
	return func(r *http.Request) (checkin.AuthorizationInfo, error) {
		if err != nil {
			return checkin.AuthorizationInfo{}, err
		}
		return checkin.AuthorizationInfo{
			Username: username,
			IsAdmin:  admin,
		}, nil
	}
}

func noValidTokenTest(t *testing.T, r *http.Request, h http.Handler, auth *mock.Authenticator) {
	original := auth.AuthenticateFn
	auth.AuthenticateFn = authenticateGenerator(false, nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusUnauthorized, w.Result().StatusCode)
	auth.AuthenticateFn = authenticateGenerator(false, errors.New("An error"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusBadRequest, w.Result().StatusCode)
	auth.AuthenticateFn = original
}

func nonHostAccessTest(t *testing.T, r *http.Request, h http.Handler, auth *mock.Authenticator, es *mock.EventService,
	nonHostUsername string) {

	original := auth.GetAuthInfoFn
	auth.GetAuthInfoFn = getAuthInfoGenerator(nonHostUsername, false, nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusForbidden, w.Result().StatusCode)
	auth.GetAuthInfoFn = getAuthInfoGenerator("", false, errors.New("An error"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusBadRequest, w.Result().StatusCode)
	auth.GetAuthInfoFn = original

	//also check what happens if check host fails
	originalCheckHost := es.CheckHostFn
	es.CheckHostFn = checkHostGenerator("", "", errors.New("An error"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	es.CheckHostFn = originalCheckHost
}

func adminAccessTest(t *testing.T, r *http.Request, h http.Handler, auth *mock.Authenticator,
	outputTester func(*http.Response)) {

	original := auth.GetAuthInfoFn
	auth.GetAuthInfoFn = getAuthInfoGenerator("random_admin_name", true, nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	outputTester(w.Result())
	auth.GetAuthInfoFn = getAuthInfoGenerator("", false, errors.New("An error"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusBadRequest, w.Result().StatusCode)
	auth.GetAuthInfoFn = original
}

func eventDoesNotExistTest(t *testing.T, badRequest *http.Request, h http.Handler, es *mock.EventService) {
	w := httptest.NewRecorder()
	h.ServeHTTP(w, badRequest)
	test.Equals(t, http.StatusNotFound, w.Result().StatusCode)

	original := es.CheckIfExistsFn
	es.CheckIfExistsFn = checkIfExistsGenerator("", errors.New("An error"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, badRequest)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	es.CheckIfExistsFn = original
}

func TestHandleReleased(t *testing.T) {
	var es mock.EventService
	var auth mock.Authenticator
	gh := myhttp.GuestHandler{}
	h := myhttp.NewEventHandler(&es, &auth, &gh)

	es.CheckIfExistsFn = checkIfExistsGenerator("300", nil)
	eventFnGenerator := func(offset time.Duration, trueID string, valid bool, err error) func(string) (checkin.Event, error) {
		return func(ID string) (checkin.Event, error) {
			if ID != trueID {
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
	es.EventFn = eventFnGenerator(-1*time.Hour, "300", true, nil)

	//test normal functionality

	//test event released 1 hour ago
	r := httptest.NewRequest("GET", "/api/v0/events/300/released", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	var release bool
	json.NewDecoder(w.Result().Body).Decode(&release)
	test.Equals(t, true, release)

	//test event released 1 hour in the future
	es.EventFn = eventFnGenerator(1*time.Hour, "300", true, nil)
	r = httptest.NewRequest("GET", "/api/v0/events/300/released", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	json.NewDecoder(w.Result().Body).Decode(&release)
	test.Equals(t, false, release)

	//test no release time set
	es.EventFn = eventFnGenerator(0, "300", false, nil)
	r = httptest.NewRequest("GET", "/api/v0/events/300/released", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	json.NewDecoder(w.Result().Body).Decode(&release)
	test.Equals(t, true, release)

	//test error in getting event
	es.EventFn = eventFnGenerator(0, "300", true, errors.New("An error"))
	r = httptest.NewRequest("GET", "/api/v0/events/300/released", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	es.EventFn = eventFnGenerator(-1*time.Hour, "300", true, nil)

	//test event does not exist
	r = httptest.NewRequest("GET", "/api/v0/events/200/released", nil)
	eventDoesNotExistTest(t, r, h, &es)
}

func TestHandleEventsBy(t *testing.T) {
	var es mock.EventService
	var auth mock.Authenticator
	gh := myhttp.GuestHandler{}
	h := myhttp.NewEventHandler(&es, &auth, &gh)

	auth.AuthenticateFn = authenticateGenerator(true, nil)
	auth.GetAuthInfoFn = getAuthInfoGenerator("testing_username", false, nil)
	eventsByGenerator := func(err error) func(username string) ([]checkin.Event, error) {
		return func(username string) ([]checkin.Event, error) {
			if username != "testing_username" {
				t.Fatal("Unexpected username: " + username + ", expected testing_username")
			}

			if err != nil {
				return nil, err
			}
			return []checkin.Event{checkin.Event{ID: "100"}, checkin.Event{ID: "200"}, checkin.Event{ID: "300"}}, nil
		}
	}
	es.EventsByFn = eventsByGenerator(nil)

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
	auth.GetAuthInfoFn = getAuthInfoGenerator("", false, errors.New("An error"))
	r = httptest.NewRequest("GET", "/api/v0/events", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusBadRequest, w.Result().StatusCode)
	auth.GetAuthInfoFn = getAuthInfoGenerator("testing_username", true, nil)

	//test error fetching events
	es.EventsByFn = eventsByGenerator(errors.New("An error"))
	r = httptest.NewRequest("GET", "/api/v0/events", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	es.EventsByFn = eventsByGenerator(nil)

	//test no valid token
	r = httptest.NewRequest("GET", "/api/v0/events", nil)
	noValidTokenTest(t, r, h, &auth)
}

func TestHandleEvent(t *testing.T) {
	var es mock.EventService
	var auth mock.Authenticator
	gh := myhttp.GuestHandler{}
	h := myhttp.NewEventHandler(&es, &auth, &gh)

	es.CheckIfExistsFn = checkIfExistsGenerator("300", nil)
	es.CheckHostFn = checkHostGenerator("testing_username", "300", nil)
	auth.AuthenticateFn = authenticateGenerator(true, nil)
	auth.GetAuthInfoFn = getAuthInfoGenerator("testing_username", false, nil)
	eventGenerator := func(expectedID string, err error) func(string) (checkin.Event, error) {
		return func(ID string) (checkin.Event, error) {
			if ID != "300" {
				t.Fatal("Unexpected username: " + ID + ", expected 300")
			}
			if err != nil {
				return checkin.Event{}, err
			}
			return checkin.Event{ID: "300"}, nil
		}
	}
	es.EventFn = eventGenerator("300", nil)

	//Test normal behavior
	r := httptest.NewRequest("GET", "/api/v0/events/300", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	var event checkin.Event
	json.NewDecoder(w.Result().Body).Decode(&event)
	test.Equals(t, checkin.Event{ID: "300"}, event)

	//Test error fetching event
	es.EventFn = eventGenerator("", errors.New("An error"))
	r = httptest.NewRequest("GET", "/api/v0/events/300", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	es.EventFn = eventGenerator("300", nil)

	//access restriction tests
	r = httptest.NewRequest("GET", "/api/v0/events/300", nil)

	//Test access by another user
	nonHostAccessTest(t, r, h, &auth, &es, "unauthorized_person")

	//Test access by admin
	adminAccessTest(t, r, h, &auth, func(r *http.Response) {
		json.NewDecoder(r.Body).Decode(&event)
		test.Equals(t, checkin.Event{ID: "300"}, event)
	})

	//Test invalid token
	noValidTokenTest(t, r, h, &auth)

	//Test invalid eventID
	r = httptest.NewRequest("GET", "/api/v0/events/100", nil)
	eventDoesNotExistTest(t, r, h, &es)

}

func TestHandleDeleteEvent(t *testing.T) {
	var es mock.EventService
	var auth mock.Authenticator
	gh := myhttp.GuestHandler{}
	h := myhttp.NewEventHandler(&es, &auth, &gh)

	es.CheckIfExistsFn = checkIfExistsGenerator("200", nil)
	es.CheckHostFn = checkHostGenerator("some_guy", "200", nil)
	auth.AuthenticateFn = authenticateGenerator(true, nil)
	auth.GetAuthInfoFn = getAuthInfoGenerator("some_guy", false, nil)
	deleteEventGenerator := func(expectedID string, err error) func(string) error {
		return func(ID string) error {
			if ID != expectedID {
				t.Fatal("Unexpected ID: " + ID + ", expected 300")
			}
			return err
		}
	}
	es.DeleteEventFn = deleteEventGenerator("200", nil)

	//Test normal behavior
	r := httptest.NewRequest("DELETE", "/api/v0/events/200", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusOK, w.Result().StatusCode)

	//Test error in event deletion
	es.DeleteEventFn = deleteEventGenerator("200", errors.New("An error"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	es.DeleteEventFn = deleteEventGenerator("200", nil)

	//Test access restrictions

	//Test access by another user
	nonHostAccessTest(t, r, h, &auth, &es, "unauthorized_person")

	//Test access by admin
	adminAccessTest(t, r, h, &auth, func(r *http.Response) {
		test.Equals(t, http.StatusOK, r.StatusCode)
	})

	//Test invalid token
	noValidTokenTest(t, r, h, &auth)

	//Test invalid eventID
	r = httptest.NewRequest("GET", "/api/v0/events/100", nil)
	eventDoesNotExistTest(t, r, h, &es)

}

func TestHandleURLTaken(t *testing.T) {
	var es mock.EventService
	var auth mock.Authenticator
	gh := myhttp.GuestHandler{}
	h := myhttp.NewEventHandler(&es, &auth, &gh)

	auth.AuthenticateFn = authenticateGenerator(true, nil)
	auth.GetAuthInfoFn = getAuthInfoGenerator("testing_username", false, nil)
	urlExistsGenerator := func(err error) func(url string) (bool, error) {
		return func(url string) (bool, error) {
			if url != "testurl" {
				t.Fatal("Expected test_url, instead: " + url)
			}

			return false, err
		}
	}
	es.URLExistsFn = urlExistsGenerator(nil)

	//Test normal behavior
	r := httptest.NewRequest("GET", "/api/v0/events/takenurls/testurl", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	var taken bool
	json.NewDecoder(w.Result().Body).Decode(&taken)
	test.Equals(t, false, taken)

	//Test if url exists returns an error
	es.URLExistsFn = urlExistsGenerator(errors.New("An error"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)

	//test no valid token
	r = httptest.NewRequest("GET", "/api/v0/events", nil)
	noValidTokenTest(t, r, h, &auth)
}