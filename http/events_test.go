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
	"strings"
	"testing"
	"time"

	"github.com/guregu/null"
)

//Generates a CheckIfExists mock function which will return true if the ID passed is equal to
//expectedID
//If err is non-nil, function generated will always return an error (and zero values)
func checkIfExistsGenerator(expectedID string, err error) func(string) (bool, error) {
	return func(id string) (bool, error) {
		if err != nil {
			return false, err
		}

		return id == expectedID, nil
	}
}

//Generates a CheckHost mock function which will return true if the username and eventID passed in
//both match the expectedUsername and expectedID - returns false otherwise
//If err is non-nil, will always return an error (and zero values)
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

//Generates an Authenticate mock function which will return (authenticate, err)
func authenticateGenerator(authenticate bool, err error) func(r *http.Request) (bool, error) {
	return func(r *http.Request) (bool, error) {
		return authenticate, err
	}
}

//Generates a GetAuthInfo mock function which will return a checkin.AuthorizationInfo object with the username
//and admin status supplied.
//If error is non-nil, will return an error and an empty checkin.AuthorizationInfo{}
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

//Generates a URLExists mock function which returns true if the url passed in matches
//the expected string
//If error is non-nil, will return error and false.
func urlExistsGenerator(expected string, err error) func(string) (bool, error) {
	return func(url string) (bool, error) {
		if err != nil {
			return false, err
		}
		return url == expected, nil
	}
}

//Tests if a nonvalid token can access an endpoint (it should not be able to)
//The request r must be made to an endpoint with said access control
//A mock AuthenticateFn is set up to return false
//Tests whether this results in a 401 error
//Also tests what happens if the authenticate function returns an error, which suggests
//the token was badly formed, so checks for a 400 error
func noValidTokenTest(t *testing.T, r *http.Request, h http.Handler, auth *mock.Authenticator) {
	original := auth.AuthenticateFn
	auth.AuthenticateFn = authenticateGenerator(false, errors.New("An error"))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusBadRequest, w.Result().StatusCode)
	auth.AuthenticateFn = authenticateGenerator(false, nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusUnauthorized, w.Result().StatusCode)
	auth.AuthenticateFn = original
}

//Tests if a non-host can access an endpoint, with the expectation that they cant
//The request r must be made to an endpoint with said access control, and a username
//must be provided that is not recognized as the username by the CheckHost function (consider deprecating
//in future, set CheckHost within this test as well)
//A mock GetAuthInfoFn is set to return the fake host without admin controls, and check if this results in
//a 403 error
//Also checks for handling of errors in GetAuthInfo (400 Bad Request) and CheckHost (500 Internal Server Error)
func nonHostAccessTest(t *testing.T, r *http.Request, h http.Handler, auth *mock.Authenticator, es *mock.EventService,
	nonHostUsername string) {

	original := auth.GetAuthInfoFn
	auth.GetAuthInfoFn = getAuthInfoGenerator("", false, errors.New("An error"))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusBadRequest, w.Result().StatusCode)
	auth.GetAuthInfoFn = getAuthInfoGenerator(nonHostUsername, false, nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusForbidden, w.Result().StatusCode)
	auth.GetAuthInfoFn = original

	//also check what happens if check host fails
	originalCheckHost := es.CheckHostFn
	es.CheckHostFn = checkHostGenerator("", "", errors.New("An error"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	es.CheckHostFn = originalCheckHost
}

//Tests if admins can access an endpoint
//The request r must be made to an endpoint with said access control
//Sets up GetAuthInfoFn to return an admin account
//Note that this request is supposed to succeed, so an outputTester function must be supplied
//Which should test if the output (the http Response) is what is expected in a success case
//Also test for handling of error in GetAuthInfo (400 Bad Request)
func adminAccessTest(t *testing.T, r *http.Request, h http.Handler, auth *mock.Authenticator,
	outputTester func(*http.Response)) {

	original := auth.GetAuthInfoFn
	auth.GetAuthInfoFn = getAuthInfoGenerator("", false, errors.New("An error"))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusBadRequest, w.Result().StatusCode)
	auth.GetAuthInfoFn = getAuthInfoGenerator("random_admin_name", true, nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	outputTester(w.Result())
	auth.GetAuthInfoFn = original
}

//Tests a request that is made to an endpoint which the mock event service CheckIfExists should
//return false (IE a non-existent endpoint). Checks that a 404 is returnd
//Also checks for a 500 if checkIfExists returns an error
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

func TestHandleCreateEvent(t *testing.T) {
	var es mock.EventService
	var auth mock.Authenticator
	gh := myhttp.GuestHandler{}
	h := myhttp.NewEventHandler(&es, &auth, &gh)

	es.URLExistsFn = urlExistsGenerator("/hello", nil)
	auth.AuthenticateFn = authenticateGenerator(true, nil)
	auth.GetAuthInfoFn = getAuthInfoGenerator("testing_username", false, nil)
	createEventGenerator := func(expectedEvent checkin.Event, err error) func(checkin.Event, string) error {
		return func(event checkin.Event, hostname string) error {
			event.ID = ""
			if event != expectedEvent {
				t.Fatal("Event was differented than expected. Received: ", event, " expected",
					expectedEvent)
			}
			if hostname != "testing_username" {
				t.Fatal("Host name was different than expected. Received: " + hostname + ", expected " +
					"testing_username")
			}
			return err
		}
	}
	expectedEvent := checkin.Event{
		Name:    "MyEvent",
		URL:     "/hello2",
		Start:   null.Time{Time: time.Date(2019, 3, 15, 8, 20, 0, 0, time.UTC), Valid: true},
		End:     null.Time{Time: time.Date(2019, 3, 15, 10, 0, 0, 0, time.UTC), Valid: true},
		Release: null.Time{Time: time.Date(2019, 3, 15, 8, 0, 0, 0, time.UTC), Valid: true},
		Lat:     null.FloatFrom(1.388),
		Long:    null.FloatFrom(2),
		Radius:  null.FloatFrom(5),
	}
	es.CreateEventFn = createEventGenerator(expectedEvent, nil)

	//test normal functionality
	r := httptest.NewRequest("POST", "/api/v0/events",
		strings.NewReader("{\"name\":\"MyEvent\",\"url\":\"/hello2\",\"startDateTime\":\"2019-03-15T08:20:00Z\","+
			"\"endDateTime\":\"2019-03-15T10:00:00Z\", \"releaseDateTime\":\"2019-03-15T08:00:00Z\","+
			"\"lat\":\"1.388\",\"long\":\"2\",\"radius\":\"5\"}"))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusCreated, w.Result().StatusCode)

	//test additional unknown fields added
	r = httptest.NewRequest("POST", "/api/v0/events",
		strings.NewReader("{\"name\":\"MyEvent\",\"url\":\"/hello2\",\"startDateTime\":\"2019-03-15T08:20:00Z\","+
			"\"endDateTime\":\"2019-03-15T10:00:00Z\", \"releaseDateTime\":\"2019-03-15T08:00:00Z\","+
			"\"lat\":\"1.388\",\"long\":\"2\",\"radius\":\"5\", \"host\":\"Bob\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusBadRequest, w.Result().StatusCode)

	//test supplied update or create
	es.CreateEventInvoked = false
	r = httptest.NewRequest("POST", "/api/v0/events",
		strings.NewReader("{\"name\":\"MyEvent\",\"url\":\"/hello2\",\"startDateTime\":\"2019-03-15T08:20:00Z\","+
			"\"endDateTime\":\"2019-03-15T10:00:00Z\", \"releaseDateTime\":\"2019-03-15T08:00:00Z\","+
			"\"lat\":\"1.388\",\"long\":\"2\",\"radius\":\"5\", \"updatedAt\":\"2019-03-12T09:30:00Z\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusBadRequest, w.Result().StatusCode)
	test.Assert(t, !es.CreateEventInvoked, "Create event invoked even though updatedAt or createdAt invoked")
	es.CreateEventInvoked = false
	r = httptest.NewRequest("POST", "/api/v0/events",
		strings.NewReader("{\"name\":\"MyEvent\",\"url\":\"/hello2\",\"startDateTime\":\"2019-03-15T08:20:00Z\","+
			"\"endDateTime\":\"2019-03-15T10:00:00Z\", \"releaseDateTime\":\"2019-03-15T08:00:00Z\","+
			"\"lat\":\"1.388\",\"long\":\"2\",\"radius\":\"5\", \"createdAt\":\"2019-03-12T16:00:30Z\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusBadRequest, w.Result().StatusCode)
	test.Assert(t, !es.CreateEventInvoked, "Create event invoked even though updatedAt or createdAt invoked")

	//test no URL or name
	es.CreateEventInvoked = false
	r = httptest.NewRequest("POST", "/api/v0/events",
		strings.NewReader("{\"url\":\"/hello2\",\"startDateTime\":\"2019-03-15T08:20:00Z\","+
			"\"endDateTime\":\"2019-03-15T10:00:00Z\", \"releaseDateTime\":\"2019-03-15T08:00:00Z\","+
			"\"lat\":\"1.388\",\"long\":\"2\",\"radius\":\"5\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusBadRequest, w.Result().StatusCode)
	test.Assert(t, !es.CreateEventInvoked, "Create event invoked even though blank URL/name")
	es.CreateEventInvoked = false
	r = httptest.NewRequest("POST", "/api/v0/events",
		strings.NewReader("{\"name\":\"MyEvent\",\"startDateTime\":\"2019-03-15T08:20:00Z\","+
			"\"endDateTime\":\"2019-03-15T10:00:00Z\", \"releaseDateTime\":\"2019-03-15T08:00:00Z\","+
			"\"lat\":\"1.388\",\"long\":\"2\",\"radius\":\"5\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusBadRequest, w.Result().StatusCode)
	test.Assert(t, !es.CreateEventInvoked, "Create event invoked even though blank URL/name")

	//test URL already in use
	r = httptest.NewRequest("POST", "/api/v0/events",
		strings.NewReader("{\"name\":\"MyEvent\",\"url\":\"/hello\",\"startDateTime\":\"2019-03-15T08:20:00Z\","+
			"\"endDateTime\":\"2019-03-15T10:00:00Z\", \"releaseDateTime\":\"2019-03-15T08:00:00Z\","+
			"\"lat\":\"1.388\",\"long\":\"2\",\"radius\":\"5\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusConflict, w.Result().StatusCode)

	//test minimum required fields
	expectedEvent = checkin.Event{
		Name: "MyEvent",
		URL:  "/hello2",
	}
	es.CreateEventFn = createEventGenerator(expectedEvent, nil)
	r = httptest.NewRequest("POST", "/api/v0/events",
		strings.NewReader("{\"name\":\"MyEvent\",\"url\":\"/hello2\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusCreated, w.Result().StatusCode)

	//Test invalid time formats
	r = httptest.NewRequest("POST", "/api/v0/events",
		strings.NewReader("{\"name\":\"MyEvent\",\"url\":\"/hello2\", \"startDateTime\":\"2019-03-15 08:20\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusBadRequest, w.Result().StatusCode)
	r = httptest.NewRequest("POST", "/api/v0/events",
		strings.NewReader("{\"name\":\"MyEvent\",\"url\":\"/hello2\", \"startDateTime\":\"2019-03-15T08:20:00Z08:00\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusBadRequest, w.Result().StatusCode)

	//Test error checking if url exists
	es.URLExistsFn = urlExistsGenerator("/hello", errors.New("An error"))
	r = httptest.NewRequest("POST", "/api/v0/events",
		strings.NewReader("{\"name\":\"MyEvent\",\"url\":\"/hello2\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	es.URLExistsFn = urlExistsGenerator("/hello", nil)

	//test error getting auth info (for username for host name of event)
	auth.GetAuthInfoFn = getAuthInfoGenerator("", false, errors.New("An error"))
	r = httptest.NewRequest("POST", "/api/v0/events",
		strings.NewReader("{\"name\":\"MyEvent\",\"url\":\"/hello2\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	auth.GetAuthInfoFn = getAuthInfoGenerator("testing_username", false, nil)

	//test error creating event
	es.CreateEventFn = createEventGenerator(expectedEvent, errors.New("An error"))
	r = httptest.NewRequest("POST", "/api/v0/events",
		strings.NewReader("{\"name\":\"MyEvent\",\"url\":\"/hello2\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	es.CreateEventFn = createEventGenerator(expectedEvent, nil)

	//test invalid token
	r = httptest.NewRequest("POST", "/api/v0/events",
		strings.NewReader("{\"name\":\"MyEvent\",\"url\":\"/hello2\"}"))
	noValidTokenTest(t, r, h, &auth)
}

func TestHandleUpdateEvent(t *testing.T) {
	var es mock.EventService
	var auth mock.Authenticator
	gh := myhttp.GuestHandler{}
	h := myhttp.NewEventHandler(&es, &auth, &gh)

	es.CheckIfExistsFn = checkIfExistsGenerator("300", nil)
	es.CheckHostFn = checkHostGenerator("testing_username", "300", nil)
	auth.AuthenticateFn = authenticateGenerator(true, nil)
	auth.GetAuthInfoFn = getAuthInfoGenerator("testing_username", false, nil)
	srcEvent := checkin.Event{ID: "300",
		Name:      "Hello",
		URL:       "/someplace",
		CreatedAt: time.Date(2019, 3, 26, 15, 35, 10, 0, time.UTC),
		UpdatedAt: time.Date(2019, 3, 26, 15, 35, 10, 0, time.UTC),
	}
	eventGenerator := func(srcEvent checkin.Event, err error) func(string) (checkin.Event, error) {
		return func(ID string) (checkin.Event, error) {
			if ID != "300" {
				t.Fatal("Unexpected username: " + ID + ", expected 300")
			}
			if err != nil {
				return checkin.Event{}, err
			}
			return srcEvent, nil
		}
	}
	es.EventFn = eventGenerator(srcEvent, nil)
	es.URLExistsFn = urlExistsGenerator("/knownurl", nil)
	updateEventGenerator := func(expectedEvent checkin.Event, err error) func(checkin.Event) error {
		return func(event checkin.Event) error {
			if event != expectedEvent {
				t.Fatal("Unexpected event. Expected: ", expectedEvent, ", received ", event)
			}
			return err
		}
	}
	expEvent := checkin.Event{
		ID:        "300",
		Name:      "MyEvent",
		URL:       "/hello2",
		Start:     null.Time{Time: time.Date(2019, 3, 15, 8, 20, 0, 0, time.UTC), Valid: true},
		End:       null.Time{Time: time.Date(2019, 3, 15, 10, 0, 0, 0, time.UTC), Valid: true},
		Release:   null.Time{Time: time.Date(2019, 3, 15, 8, 0, 0, 0, time.UTC), Valid: true},
		Lat:       null.FloatFrom(1.388),
		Long:      null.FloatFrom(2),
		Radius:    null.FloatFrom(5),
		CreatedAt: time.Date(2019, 3, 26, 15, 35, 10, 0, time.UTC),
		UpdatedAt: time.Date(2019, 3, 26, 15, 35, 10, 0, time.UTC),
	}
	es.UpdateEventFn = updateEventGenerator(expEvent, nil)

	//test normal functionality, replace everything
	r := httptest.NewRequest("PATCH", "/api/v0/events/300",
		strings.NewReader("{\"name\":\"MyEvent\",\"url\":\"/hello2\",\"startDateTime\":\"2019-03-15T08:20:00Z\","+
			"\"endDateTime\":\"2019-03-15T10:00:00Z\", \"releaseDateTime\":\"2019-03-15T08:00:00Z\","+
			"\"lat\":\"1.388\",\"long\":\"2\",\"radius\":\"5\"}"))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusOK, w.Result().StatusCode)

	//test unknown fields
	r = httptest.NewRequest("PATCH", "/api/v0/events/300",
		strings.NewReader("{\"name\":\"MyEvent\",\"url\":\"/hello2\",\"startDateTime\":\"2019-03-15T08:20:00Z\","+
			"\"endDateTime\":\"2019-03-15T10:00:00Z\", \"releaseDateTime\":\"2019-03-15T08:00:00Z\","+
			"\"lat\":\"1.388\",\"long\":\"2\",\"radius\":\"5\", \"lmao\":\"12\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusBadRequest, w.Result().StatusCode)

	//test try to change update/create/ID
	es.UpdateEventInvoked = false
	r = httptest.NewRequest("PATCH", "/api/v0/events/300",
		strings.NewReader("{\"name\":\"MyEvent\", \"updatedAt\":\"2019-03-12T09:30:00Z\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusBadRequest, w.Result().StatusCode)
	test.Assert(t, !es.UpdateEventInvoked, "Update event invoked even though updatedAt changed")
	es.UpdateEventInvoked = false
	r = httptest.NewRequest("PATCH", "/api/v0/events/300",
		strings.NewReader("{\"name\":\"MyEvent\", \"createdAt\":\"2019-03-12T16:00:30Z\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusBadRequest, w.Result().StatusCode)
	test.Assert(t, !es.UpdateEventInvoked, "Update event invoked even though createdAt changed")
	es.UpdateEventInvoked = false
	r = httptest.NewRequest("PATCH", "/api/v0/events/300",
		strings.NewReader("{\"name\":\"MyEvent\", \"eventId\":\"200\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusBadRequest, w.Result().StatusCode)
	test.Assert(t, !es.UpdateEventInvoked, "Update event invoked even though ID changed")

	//test try to set blank URL/name
	es.UpdateEventInvoked = false
	r = httptest.NewRequest("PATCH", "/api/v0/events/300",
		strings.NewReader("{\"name\":\"\", \"startDateTime\":\"2019-03-12T09:30:00Z\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusBadRequest, w.Result().StatusCode)
	test.Assert(t, !es.UpdateEventInvoked, "Update event invoked even though name set to blank")
	es.UpdateEventInvoked = false
	r = httptest.NewRequest("PATCH", "/api/v0/events/300",
		strings.NewReader("{\"url\":\"\", \"startDateTime\":\"2019-03-12T09:30:00Z\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusBadRequest, w.Result().StatusCode)
	test.Assert(t, !es.UpdateEventInvoked, "Update event invoked even though url set to blank")

	//test URL already in use
	r = httptest.NewRequest("PATCH", "/api/v0/events/300",
		strings.NewReader("{\"url\":\"/knownurl\",\"lat\":\"1.388\",\"long\":\"2\",\"radius\":\"5\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusConflict, w.Result().StatusCode)

	expEvent = checkin.Event{
		ID:        "300",
		Name:      "MyEvent",
		URL:       "/hello2",
		Start:     null.Time{Time: time.Date(2019, 3, 15, 8, 20, 0, 0, time.UTC), Valid: true},
		CreatedAt: time.Date(2019, 3, 26, 15, 35, 10, 0, time.UTC),
		UpdatedAt: time.Date(2019, 3, 26, 15, 35, 10, 0, time.UTC),
	}

	//Test fetching original event fails
	es.EventFn = eventGenerator(checkin.Event{}, errors.New("An error"))
	r = httptest.NewRequest("PATCH", "/api/v0/events/300",
		strings.NewReader("{\"name\":\"MyEvent\", \"url\":\"/hello2\", \"startDateTime\":\"2019-03-15T08:20:00Z\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	es.EventFn = eventGenerator(srcEvent, nil)

	//Test invalid time format
	r = httptest.NewRequest("PATCH", "/api/v0/events/300",
		strings.NewReader("{\"name\":\"MyEvent\", \"url\":\"/hello2\", \"startDateTime\":\"2019-03-15T08:20Z\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusBadRequest, w.Result().StatusCode)

	//Test error checking if new URL already taken
	es.URLExistsFn = urlExistsGenerator("/knownurl", errors.New("An error"))
	es.UpdateEventInvoked = false
	r = httptest.NewRequest("PATCH", "/api/v0/events/300",
		strings.NewReader("{\"name\":\"MyEvent\", \"url\":\"/hello2\", \"startDateTime\":\"2019-03-15T08:20:00Z\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	test.Assert(t, !es.UpdateEventInvoked, "Update event invoked even though could not determine if URL unique")
	es.URLExistsFn = urlExistsGenerator("/knownurl", nil)

	//Test error in updating event
	es.UpdateEventFn = updateEventGenerator(expEvent, errors.New("An error"))
	r = httptest.NewRequest("PATCH", "/api/v0/events/300",
		strings.NewReader("{\"name\":\"MyEvent\", \"url\":\"/hello2\", \"startDateTime\":\"2019-03-15T08:20:00Z\"}"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	es.UpdateEventFn = updateEventGenerator(expEvent, nil)

	//Test usual access control
	//Test access by another user
	r = httptest.NewRequest("PATCH", "/api/v0/events/300",
		strings.NewReader("{\"name\":\"MyEvent\", \"url\":\"/hello2\", \"startDateTime\":\"2019-03-15T08:20:00Z\"}"))
	nonHostAccessTest(t, r, h, &auth, &es, "unauthorized_person")

	//Test access by admin
	r = httptest.NewRequest("PATCH", "/api/v0/events/300",
		strings.NewReader("{\"name\":\"MyEvent\", \"url\":\"/hello2\", \"startDateTime\":\"2019-03-15T08:20:00Z\"}"))
	adminAccessTest(t, r, h, &auth, func(r *http.Response) {
		test.Equals(t, http.StatusOK, r.StatusCode)
	})

	//Test invalid token
	r = httptest.NewRequest("PATCH", "/api/v0/events/300",
		strings.NewReader("{\"name\":\"MyEvent\", \"url\":\"/hello2\", \"startDateTime\":\"2019-03-15T08:20:00Z\"}"))

	noValidTokenTest(t, r, h, &auth)

	//Test invalid eventID
	r = httptest.NewRequest("PATCH", "/api/v0/events/200",
		strings.NewReader("{\"name\":\"MyEvent\", \"url\":\"/hello2\", \"startDateTime\":\"2019-03-15T08:20:00Z\"}"))
	eventDoesNotExistTest(t, r, h, &es)

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
