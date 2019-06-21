package http_test

//some timezone tests will only work if this test suite is run with Singapore being the local time zone, lol

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

func userCheckIfExistsGenerator(expectedUsername string, err error) func(string) (bool, error) {
	return func(username string) (bool, error) {
		if err != nil {
			return false, err
		}
		return username == expectedUsername, nil
	}
}

func userAccessTest(t *testing.T, r *http.Request, h http.Handler, auth *mock.Authenticator, username string) {
	original := auth.GetAuthInfoFn
	auth.GetAuthInfoFn = getAuthInfoGenerator(username, false, nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusForbidden, w.Result().StatusCode)
	auth.GetAuthInfoFn = getAuthInfoGenerator("", false, errors.New("An error"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusBadRequest, w.Result().StatusCode)
	auth.GetAuthInfoFn = original
}

func userDoesNotExistTest(t *testing.T, badRequest *http.Request, h http.Handler, us *mock.UserService) {
	w := httptest.NewRecorder()
	h.ServeHTTP(w, badRequest)
	test.Equals(t, http.StatusNotFound, w.Result().StatusCode)

	original := us.CheckIfExistsFn
	us.CheckIfExistsFn = userCheckIfExistsGenerator("", errors.New("An error"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, badRequest)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	us.CheckIfExistsFn = original
}

func badRequestTest(t *testing.T, badRequest *http.Request, h http.Handler) {
	w := httptest.NewRecorder()
	h.ServeHTTP(w, badRequest)
	test.Equals(t, http.StatusBadRequest, w.Result().StatusCode)
}

func TestHandleUsers(t *testing.T) {
	var us mock.UserService
	var auth mock.Authenticator
	h := myhttp.NewUserHandler(&us, &auth)

	auth.AuthenticateFn = authenticateGenerator(true, nil)
	auth.GetAuthInfoFn = getAuthInfoGenerator("my_admin", true, nil)
	usersFnGenerator := func(users []checkin.User, err error) func() ([]checkin.User, error) {
		return func() ([]checkin.User, error) {
			if err != nil {
				return nil, err
			}
			return users, nil
		}
	}
	us.UsersFn = usersFnGenerator([]checkin.User{checkin.User{Username: "Jim"}, checkin.User{Username: "Bob", CreatedAt: time.Date(2019, 1, 9, 13, 30, 0, 0, time.UTC)},
		checkin.User{Username: "Smith", LastLoggedIn: null.TimeFrom(time.Date(2019, 2, 14, 14, 30, 0, 0, time.UTC))}}, nil)

	//Test normal behavior
	r := httptest.NewRequest("GET", "/api/v0/users", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	var users []checkin.User
	json.NewDecoder(w.Result().Body).Decode(&users)
	test.Equals(t, []checkin.User{checkin.User{Username: "Jim"}, checkin.User{Username: "Bob", CreatedAt: time.Date(2019, 1, 9, 13, 30, 0, 0, time.UTC)},
		checkin.User{Username: "Smith", LastLoggedIn: null.TimeFrom(time.Date(2019, 2, 14, 14, 30, 0, 0, time.UTC))}}, users)

	//test specifying fields desired
	r = httptest.NewRequest("GET", "/api/v0/users?field=uSeRName", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	users = make([]checkin.User, 10)
	json.NewDecoder(w.Result().Body).Decode(&users)
	test.Equals(t, []checkin.User{checkin.User{Username: "Jim"}, checkin.User{Username: "Bob"},
		checkin.User{Username: "Smith"}}, users)

	r = httptest.NewRequest("GET", "/api/v0/users?field=lastloggedin", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	users = make([]checkin.User, 10)
	json.NewDecoder(w.Result().Body).Decode(&users)
	test.Equals(t, []checkin.User{checkin.User{}, checkin.User{},
		checkin.User{LastLoggedIn: null.TimeFrom(time.Date(2019, 2, 14, 14, 30, 0, 0, time.UTC))}}, users)

	r = httptest.NewRequest("GET", "/api/v0/users?field=lastloggedin&field=username", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	users = make([]checkin.User, 10)
	json.NewDecoder(w.Result().Body).Decode(&users)
	test.Equals(t, []checkin.User{checkin.User{Username: "Jim"}, checkin.User{Username: "Bob"},
		checkin.User{Username: "Smith", LastLoggedIn: null.TimeFrom(time.Date(2019, 2, 14, 14, 30, 0, 0, time.UTC))}}, users)

	r = httptest.NewRequest("GET", "/api/v0/users?field=doesntexist", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	users = make([]checkin.User, 10)
	json.NewDecoder(w.Result().Body).Decode(&users)
	test.Equals(t, []checkin.User{checkin.User{}, checkin.User{},
		checkin.User{}}, users)

	//test ?loc=Asia/Singapore (set output times to Singapore timezone)
	r = httptest.NewRequest("GET", "/api/v0/users?loc=Asia/Singapore", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	json.NewDecoder(w.Result().Body).Decode(&users)
	test.Equals(t, []checkin.User{checkin.User{Username: "Jim"}, checkin.User{Username: "Bob", CreatedAt: time.Date(2019, 1, 9, 21, 30, 0, 0, time.Local)},
		checkin.User{Username: "Smith", LastLoggedIn: null.TimeFrom(time.Date(2019, 2, 14, 22, 30, 0, 0, time.Local))}}, users)

	r = httptest.NewRequest("GET", "/api/v0/users?loc=Doesnt/Exist", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusBadRequest, w.Result().StatusCode)

	//Test no users
	r = httptest.NewRequest("GET", "/api/v0/users?loc=Asia/Singapore", nil)
	us.UsersFn = usersFnGenerator([]checkin.User{}, nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	json.NewDecoder(w.Result().Body).Decode(&users)
	test.Equals(t, []checkin.User{}, users)

	//Test error in getting users
	us.UsersFn = usersFnGenerator(nil, errors.New("An error"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	r = httptest.NewRequest("GET", "/api/v0/users", nil)
	us.UsersFn = usersFnGenerator([]checkin.User{checkin.User{Username: "Jim"}, checkin.User{Username: "Bob", CreatedAt: time.Date(2019, 1, 9, 13, 30, 0, 0, time.UTC)},
		checkin.User{Username: "Smith", LastLoggedIn: null.TimeFrom(time.Date(2019, 2, 14, 14, 30, 0, 0, time.UTC))}}, nil)

	//Test access controls: a user should fail to access
	//Admin should succeed
	//No valid token should fail to access
	userAccessTest(t, r, h, &auth, "some_guy")
	adminAccessTest(t, r, h, &auth, func(r *http.Response) {
		json.NewDecoder(r.Body).Decode(&users)
		test.Equals(t, []checkin.User{checkin.User{Username: "Jim"}, checkin.User{Username: "Bob", CreatedAt: time.Date(2019, 1, 9, 13, 30, 0, 0, time.UTC)},
			checkin.User{Username: "Smith", LastLoggedIn: null.TimeFrom(time.Date(2019, 2, 14, 14, 30, 0, 0, time.UTC))}}, users)
	})
	noValidTokenTest(t, r, h, &auth)
}

func TestHandleUser(t *testing.T) {
	var us mock.UserService
	var auth mock.Authenticator
	h := myhttp.NewUserHandler(&us, &auth)

	auth.AuthenticateFn = authenticateGenerator(true, nil)
	auth.GetAuthInfoFn = getAuthInfoGenerator("somebody", false, nil)
	us.CheckIfExistsFn = checkIfExistsGenerator("somebody", nil)
	userFnGenerator := func(err error) func(string) (checkin.User, error) {
		return func(username string) (checkin.User, error) {
			if username != "somebody" {
				t.Fatal("Expected username somebody instead got: " + username)
			}
			if err != nil {
				return checkin.User{}, err
			}
			return checkin.User{Username: "somebody", UpdatedAt: time.Date(2018, 2, 13, 10, 30, 0, 0, time.UTC)}, nil
		}
	}
	us.UserFn = userFnGenerator(nil)

	//Test normal behavior
	r := httptest.NewRequest("GET", "/api/v0/users/somebody", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	var user checkin.User
	json.NewDecoder(w.Result().Body).Decode(&user)
	test.Equals(t, checkin.User{Username: "somebody", UpdatedAt: time.Date(2018, 2, 13, 10, 30, 0, 0, time.UTC)}, user)

	//Test get specific fields only
	r = httptest.NewRequest("GET", "/api/v0/users/somebody?field=uPDatedat", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	user = checkin.User{}
	json.NewDecoder(w.Result().Body).Decode(&user)
	test.Equals(t, checkin.User{UpdatedAt: time.Date(2018, 2, 13, 10, 30, 0, 0, time.UTC)}, user)

	//Test get specific field, field doesn't exist
	r = httptest.NewRequest("GET", "/api/v0/users/somebody?field=lolwut", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	user = checkin.User{}
	json.NewDecoder(w.Result().Body).Decode(&user)
	test.Equals(t, checkin.User{}, user)

	//Test setting timezone
	r = httptest.NewRequest("GET", "/api/v0/users/somebody?loc=Asia/Singapore", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	json.NewDecoder(w.Result().Body).Decode(&user)
	test.Equals(t, checkin.User{Username: "somebody", UpdatedAt: time.Date(2018, 2, 13, 18, 30, 0, 0, time.Local)}, user)

	//Test error getting user
	us.UserFn = userFnGenerator(errors.New("An error"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	us.UserFn = userFnGenerator(nil)

	//Test access controls: a *different* user should fail to access
	//Admin should succeed
	//No valid token should fail to access
	userAccessTest(t, r, h, &auth, "some_guy")
	adminAccessTest(t, r, h, &auth, func(r *http.Response) {
		json.NewDecoder(w.Result().Body).Decode(&user)
		test.Equals(t, checkin.User{Username: "somebody", UpdatedAt: time.Date(2018, 2, 13, 18, 30, 0, 0, time.Local)}, user)
	})
	noValidTokenTest(t, r, h, &auth)

	r = httptest.NewRequest("GET", "/api/v0/users/somebodyoncetoldme", nil)
	userDoesNotExistTest(t, r, h, &us)

}

func TestHandleCreateUser(t *testing.T) {
	var us mock.UserService
	var auth mock.Authenticator
	h := myhttp.NewUserHandler(&us, &auth)

	auth.AuthenticateFn = authenticateGenerator(true, nil)
	auth.GetAuthInfoFn = getAuthInfoGenerator("my_admin", true, nil) //only admins can create users
	us.CheckIfExistsFn = checkIfExistsGenerator("takenName", nil)
	createUserFnGenerator := func(expectedUser checkin.User, err error) func(checkin.User) error {
		return func(u checkin.User) error {
			//all fields are the same and
			//passwords equal
			if u.PasswordPlaintext == expectedUser.PasswordPlaintext || (u.PasswordPlaintext != nil &&
				expectedUser.PasswordPlaintext != nil &&
				*u.PasswordPlaintext == *expectedUser.PasswordPlaintext) {
				u.PasswordPlaintext, expectedUser.PasswordPlaintext = nil, nil
				if u == expectedUser {
					return err
				} else {
					t.Fatal("Unexpected user in create user. Received: ", u, ", expected: ", expectedUser)
				}
			} else {
				t.Fatal("Unexpected user in create user. Received: ", u, ", expected: ", expectedUser)
			}
			return err
		}
	}
	pwd := "1234"
	expUser := checkin.User{
		Username:          "bob",
		Name:              "Bob",
		PasswordPlaintext: &pwd,
	}
	us.CreateUserFn = createUserFnGenerator(expUser, nil)

	//Test normal functionality
	r := httptest.NewRequest("POST", "/api/v0/users",
		strings.NewReader(`{"username":"bob","password":"1234","name":"Bob"}`))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusCreated, w.Result().StatusCode)

	//test username already used
	r = httptest.NewRequest("POST", "/api/v0/users",
		strings.NewReader(`{"username":"takenName","password":"1234","name":"Bob"}`))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusConflict, w.Result().StatusCode)

	//Test extra meaningless fields added
	r = httptest.NewRequest("POST", "/api/v0/users",
		strings.NewReader(`{"username":"bob","password":"1234","name":"Bob", "passwordhash":"hmm"}`))
	badRequestTest(t, r, h)

	//Test no username, password or name.
	r = httptest.NewRequest("POST", "/api/v0/users",
		strings.NewReader(`{"password":"1234","name":"Bob"}`))
	badRequestTest(t, r, h)
	r = httptest.NewRequest("POST", "/api/v0/users",
		strings.NewReader(`{"username":"bob","name":"Bob"}`))
	badRequestTest(t, r, h)
	r = httptest.NewRequest("POST", "/api/v0/users",
		strings.NewReader(`{"username":"bob","password":"1234"}`))
	badRequestTest(t, r, h)

	//Test attempt to set createdAt, updatedAt or lastLoggedIn
	r = httptest.NewRequest("POST", "/api/v0/users",
		strings.NewReader(`{"username":"bob","password":"1234","name":"Bob",
			createdAt":"2019-03-15T08:20:00Z"}`))
	badRequestTest(t, r, h)
	r = httptest.NewRequest("POST", "/api/v0/users",
		strings.NewReader(`{"username":"bob","password":"1234","name":"Bob",
			updatedAt":"2019-03-15T08:20:00Z"}`))
	badRequestTest(t, r, h)
	r = httptest.NewRequest("POST", "/api/v0/users",
		strings.NewReader(`{"username":"bob","password":"1234","name":"Bob",
			lastLoggedIn":"2019-03-15T08:20:00Z"}`))
	badRequestTest(t, r, h)

	//Test checking if exists fails
	us.CheckIfExistsFn = checkIfExistsGenerator("takenName", errors.New("An error"))
	us.CreateUserInvoked = false
	r = httptest.NewRequest("POST", "/api/v0/users",
		strings.NewReader(`{"username":"bob","password":"1234","name":"Bob"}`))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	test.Assert(t, !us.CreateUserInvoked, "CreateUser invoked even though uniqueness of username not checked")
	us.CheckIfExistsFn = checkIfExistsGenerator("takenName", nil)

	//Test create user fails
	us.CreateUserFn = createUserFnGenerator(expUser, errors.New("An error"))
	r = httptest.NewRequest("POST", "/api/v0/users",
		strings.NewReader(`{"username":"bob","password":"1234","name":"Bob"}`))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	us.CreateUserFn = createUserFnGenerator(expUser, nil)

	//Test access controls: a user should fail to access
	//Admin should succeed
	//No valid token should fail to access
	r = httptest.NewRequest("POST", "/api/v0/users",
		strings.NewReader(`{"username":"bob","password":"1234","name":"Bob"}`))
	userAccessTest(t, r, h, &auth, "some_guy")
	r = httptest.NewRequest("POST", "/api/v0/users",
		strings.NewReader(`{"username":"bob","password":"1234","name":"Bob"}`))
	adminAccessTest(t, r, h, &auth, func(r *http.Response) {
		test.Equals(t, http.StatusCreated, r.StatusCode)
	})
	r = httptest.NewRequest("POST", "/api/v0/users",
		strings.NewReader(`{"username":"bob","password":"1234","name":"Bob"}`))
	noValidTokenTest(t, r, h, &auth)
}

func TestHandleUpdateUser(t *testing.T) {
	var us mock.UserService
	var auth mock.Authenticator
	h := myhttp.NewUserHandler(&us, &auth)

	auth.AuthenticateFn = authenticateGenerator(true, nil)
	auth.GetAuthInfoFn = getAuthInfoGenerator("my_admin", true, nil) //only admins can create users
	checkIfExistsGenerator := func(expectedID string, err error, alwaysAllow bool) func(string) (bool, error) {
		return func(id string) (bool, error) {
			if err != nil && id == expectedID { //this version only fails if was going to return true
				return false, err
			}

			return id == expectedID || alwaysAllow, nil
		}
	}
	us.CheckIfExistsFn = checkIfExistsGenerator("bob", nil, false)
	updateUserFnGenerator := func(expectedUsername string, expectedUser checkin.User, err error) func(string, checkin.User) error {
		return func(originalUsername string, u checkin.User) error {
			//all fields are the same and
			//passwords equal
			test.Equals(t, expectedUsername, originalUsername)
			if u.PasswordPlaintext == expectedUser.PasswordPlaintext || (u.PasswordPlaintext != nil &&
				expectedUser.PasswordPlaintext != nil &&
				*u.PasswordPlaintext == *expectedUser.PasswordPlaintext) {
				u.PasswordPlaintext, expectedUser.PasswordPlaintext = nil, nil
				if u == expectedUser {
					return err
				} else {
					t.Fatal("Unexpected user in create user. Received: ", u, ", expected: ", expectedUser)
				}
			} else {
				t.Fatal("Unexpected user in create user. Received: ", u, ", expected: ", expectedUser)
			}
			return err
		}
	}
	pwd := "5678"
	userFnGenerator := func(user checkin.User, err error) func(string) (checkin.User, error) {
		return func(username string) (checkin.User, error) {
			if username != user.Username {
				t.Fatal("Error in UserF, expected " + user.Username + " but received " + username)
			}
			return user, err
		}
	}
	original := checkin.User{
		Username:     "bob",
		Name:         "Bob",
		LastLoggedIn: null.TimeFrom(time.Date(2019, 3, 26, 27, 35, 13, 0, time.UTC)),
		CreatedAt:    time.Date(2019, 3, 26, 15, 35, 10, 0, time.UTC),
		UpdatedAt:    time.Date(2019, 3, 26, 15, 35, 10, 0, time.UTC),
	}
	expUser := checkin.User{ //expected user after update
		Username:          "max",
		Name:              "Max",
		PasswordPlaintext: &pwd,
		LastLoggedIn:      null.TimeFrom(time.Date(2019, 3, 26, 27, 35, 13, 0, time.UTC)),
		CreatedAt:         time.Date(2019, 3, 26, 15, 35, 10, 0, time.UTC),
		UpdatedAt:         time.Date(2019, 3, 26, 15, 35, 10, 0, time.UTC),
	}
	us.UpdateUserFn = updateUserFnGenerator("bob", expUser, nil)
	us.UserFn = userFnGenerator(original, nil)

	//Test normal functionality
	r := httptest.NewRequest("PATCH", "/api/v0/users/bob",
		strings.NewReader(`{"username":"max","password":"5678","name":"Max"}`))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusOK, w.Result().StatusCode)

	//Test no amendment to password
	expUser.PasswordPlaintext = nil
	us.UpdateUserFn = updateUserFnGenerator("bob", expUser, nil)
	r = httptest.NewRequest("PATCH", "/api/v0/users/bob",
		strings.NewReader(`{"username":"max","name":"Max"}`))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusOK, w.Result().StatusCode)
	expUser.PasswordPlaintext = &pwd
	us.UpdateUserFn = updateUserFnGenerator("bob", expUser, nil)

	//Test extra meaningless fields added
	r = httptest.NewRequest("PATCH", "/api/v0/users/bob",
		strings.NewReader(`{"username":"bob","password":"1234","name":"Bob", "passwordhash":"hmm"}`))
	badRequestTest(t, r, h)

	//Test attempt to set createdAt, updatedAt or lastLoggedIn
	r = httptest.NewRequest("PATCH", "/api/v0/users/bob",
		strings.NewReader(`{"username":"bob","password":"1234","name":"Bob",
			createdAt":"2019-03-15T08:20:00Z"}`))
	badRequestTest(t, r, h)
	r = httptest.NewRequest("PATCH", "/api/v0/users/bob",
		strings.NewReader(`{"username":"bob","password":"1234","name":"Bob",
			updatedAt":"2019-03-15T08:20:00Z"}`))
	badRequestTest(t, r, h)
	r = httptest.NewRequest("PATCH", "/api/v0/users/bob",
		strings.NewReader(`{"username":"bob","password":"1234","name":"Bob",
			lastLoggedIn":"2019-03-15T08:20:00Z"}`))
	badRequestTest(t, r, h)

	//Set blank username, password or name
	r = httptest.NewRequest("PATCH", "/api/v0/users/bob",
		strings.NewReader(`{"username":"","password":"5678","name":"Max"}`))
	badRequestTest(t, r, h)
	r = httptest.NewRequest("PATCH", "/api/v0/users/bob",
		strings.NewReader(`{"username":"max","":"5678","name":"Max"}`))
	badRequestTest(t, r, h)
	r = httptest.NewRequest("PATCH", "/api/v0/users/bob",
		strings.NewReader(`{"username":"max","password":"5678","name":""}`))
	badRequestTest(t, r, h)

	//Test change username to already used username
	us.CheckIfExistsFn = checkIfExistsGenerator("bob", nil, true) //set to always says exists
	r = httptest.NewRequest("PATCH", "/api/v0/users/bob",
		strings.NewReader(`{"username":"jim","password":"5678","name":"Max"}`))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusConflict, w.Result().StatusCode)
	us.CheckIfExistsFn = checkIfExistsGenerator("bob", nil, false) //set to always says exists

	//Test check if exists fails on second try
	us.CheckIfExistsFn = checkIfExistsGenerator("jim", errors.New("An error"), true) //set to always says exists
	//will fail only if ID = jim, so first call succeeds (to check if "bob" exists), second call fails
	r = httptest.NewRequest("PATCH", "/api/v0/users/bob",
		strings.NewReader(`{"username":"jim","password":"5678","name":"Max"}`))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	us.CheckIfExistsFn = checkIfExistsGenerator("bob", nil, false)

	//check error on getting user
	us.UserFn = userFnGenerator(original, errors.New("An error"))
	r = httptest.NewRequest("PATCH", "/api/v0/users/bob",
		strings.NewReader(`{"username":"max","password":"5678","name":"Max"}`))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	us.UserFn = userFnGenerator(original, nil)

	//check error updating user
	us.UpdateUserFn = updateUserFnGenerator("bob", expUser, errors.New("An error"))
	r = httptest.NewRequest("PATCH", "/api/v0/users/bob",
		strings.NewReader(`{"username":"max","password":"5678","name":"Max"}`))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	us.UpdateUserFn = updateUserFnGenerator("bob", expUser, nil)

	//Test access controls: another user should fail to access
	//Admin should succeed
	//No valid token should fail to access
	r = httptest.NewRequest("PATCH", "/api/v0/users/bob",
		strings.NewReader(`{"username":"max","password":"5678","name":"Max"}`))
	userAccessTest(t, r, h, &auth, "someguy")
	r = httptest.NewRequest("PATCH", "/api/v0/users/bob",
		strings.NewReader(`{"username":"max","password":"5678","name":"Max"}`))
	adminAccessTest(t, r, h, &auth, func(r *http.Response) {
		test.Equals(t, http.StatusOK, r.StatusCode)
	})
	r = httptest.NewRequest("PATCH", "/api/v0/users/bob",
		strings.NewReader(`{"username":"max","password":"5678","name":"Max"}`))
	noValidTokenTest(t, r, h, &auth)

	r = httptest.NewRequest("PATCH", "/api/v0/users/somefellow",
		strings.NewReader(`{"username":"max","password":"5678","name":"Max"}`))
	userDoesNotExistTest(t, r, h, &us)
}

func TestHandleDeleteUser(t *testing.T) {
	var us mock.UserService
	var auth mock.Authenticator
	h := myhttp.NewUserHandler(&us, &auth)

	auth.AuthenticateFn = authenticateGenerator(true, nil)
	auth.GetAuthInfoFn = getAuthInfoGenerator("somebody", false, nil)
	us.CheckIfExistsFn = checkIfExistsGenerator("somebody", nil)
	deleteUserFnGenerator := func(err error) func(string) error {
		return func(username string) error {
			if username != "somebody" {
				t.Fatal("Expected username somebody instead got: " + username)
			}
			return err
		}
	}
	us.DeleteUserFn = deleteUserFnGenerator(nil)

	//Test normal behavior
	r := httptest.NewRequest("DELETE", "/api/v0/users/somebody", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusOK, w.Result().StatusCode)

	//Test error getting user
	us.DeleteUserFn = deleteUserFnGenerator(errors.New("An error"))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	us.DeleteUserFn = deleteUserFnGenerator(nil)

	//Test access controls: a *different* user should fail to access
	//Admin should succeed
	//No valid token should fail to access
	userAccessTest(t, r, h, &auth, "some_guy")
	adminAccessTest(t, r, h, &auth, func(r *http.Response) {
		test.Equals(t, http.StatusOK, r.StatusCode)
	})
	noValidTokenTest(t, r, h, &auth)

	r = httptest.NewRequest("GET", "/api/v0/users/somebodyoncetoldme", nil)
	userDoesNotExistTest(t, r, h, &us)
}
