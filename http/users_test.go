package http_test

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
)

func userCheckIfExistsGenerator(expectedUsername string, err error) func(string) (bool, error) {
	return func(username string) (bool, error) {
		if err != nil {
			return false, err
		}
		return username == expectedUsername, nil
	}
}

func TestHandleUsers(t *testing.T) {
	var us mock.UserService
	var auth mock.Authenticator
	h := myhttp.NewUserHandler(&us, &auth)

	auth.AuthenticateFn = authenticateGenerator(true, nil)
	auth.GetAuthInfoFn = getAuthInfoGenerator("my_admin", true, nil)
	usersFnGenerator := func(err error) func() ([]checkin.User, error) {
		return func() ([]checkin.User, error) {
			if err != nil {
				return nil, err
			}
			return []checkin.User{checkin.User{Username: "Jim"}, checkin.User{Username: "Bob"},
				checkin.User{Username: "Smith"}}, nil
		}
	}
	us.UsersFn = usersFnGenerator(nil)

	//Test normal behavior
	r := httptest.NewRequest("GET", "/api/v0/users", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	var users []checkin.User
	json.NewDecoder(w.Result().Body).Decode(&users)
	test.Equals(t, []checkin.User{checkin.User{Username: "Jim"}, checkin.User{Username: "Bob"},
		checkin.User{Username: "Smith"}}, users)

	us.UsersFn = usersFnGenerator(errors.New("An error"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	us.UsersFn = usersFnGenerator(nil)
}
