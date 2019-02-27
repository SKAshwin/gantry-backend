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
	"strings"
	"testing"
)

func TestHandleLogin(t *testing.T) {
	var as mock.AuthenticationService
	var us mock.UserService
	var auth mock.Authenticator

	h := myhttp.NewAuthHandler(&as, &auth, &us)
	as.AuthenticateFn = func(username string, pwdPlaintext string, isAdmin bool) (bool, error) {
		if username == "user123" && pwdPlaintext == "abcd" && !isAdmin {
			return true, nil
		} else if username == "admin123" && pwdPlaintext == "wsxd" && isAdmin {
			return true, nil
		} else {
			return false, nil
		}
	}
	auth.IssueAuthorizationFn = func(au checkin.AuthorizationInfo, w http.ResponseWriter) error {
		reply, err := json.Marshal(au)
		w.Write(reply)
		return err
	}
	us.UpdateLastLoggedInFn = func(username string) error {
		test.Assert(t, username == "user123" || username == "admin123", "Unexpected username obtained")
		return nil
	}

	rUser := httptest.NewRequest("POST", "/api/v0/auth/users/login", strings.NewReader(`{"username":"user123","password":"abcd"}`))
	rAdmin := httptest.NewRequest("POST", "/api/v0/auth/admins/login", strings.NewReader(`{"username":"admin123","password":"wsxd"}`))

	//Test normal behavior for both endpoints
	w := httptest.NewRecorder()
	h.ServeHTTP(w, rUser)
	var ai checkin.AuthorizationInfo
	json.NewDecoder(w.Result().Body).Decode(&ai)
	test.Equals(t, checkin.AuthorizationInfo{Username: "user123", IsAdmin: false}, ai)

	w = httptest.NewRecorder()
	h.ServeHTTP(w, rAdmin)
	var aiAdmin checkin.AuthorizationInfo
	json.NewDecoder(w.Result().Body).Decode(&aiAdmin)
	test.Equals(t, checkin.AuthorizationInfo{Username: "admin123", IsAdmin: true}, aiAdmin)

	//test BadRequest
	rUser = httptest.NewRequest("POST", "/api/v0/auth/users/login", strings.NewReader(`{"username":"user123","passw`))
	rAdmin = httptest.NewRequest("POST", "/api/v0/auth/admins/login", strings.NewReader(`{"usern:"admin123","passw`))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, rUser)
	test.Equals(t, http.StatusBadRequest, w.Result().StatusCode)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, rAdmin)
	test.Equals(t, http.StatusBadRequest, w.Result().StatusCode)

	//test invalid password
	rUser = httptest.NewRequest("POST", "/api/v0/auth/users/login", strings.NewReader(`{"username":"user123","password":"acd"}`))
	rAdmin = httptest.NewRequest("POST", "/api/v0/auth/admins/login", strings.NewReader(`{"username":"admin123","password":"sxd"}`))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, rUser)
	test.Equals(t, http.StatusUnauthorized, w.Result().StatusCode)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, rAdmin)
	test.Equals(t, http.StatusUnauthorized, w.Result().StatusCode)

	//test user login to admin and vice versa
	rUser = httptest.NewRequest("POST", "/api/v0/auth/users/login", strings.NewReader(`{"username":"admin123","password":"sxd"}`))
	rAdmin = httptest.NewRequest("POST", "/api/v0/auth/admins/login", strings.NewReader(`{"username":"user123","password":"acd"}`))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, rUser)
	test.Equals(t, http.StatusUnauthorized, w.Result().StatusCode)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, rAdmin)
	test.Equals(t, http.StatusUnauthorized, w.Result().StatusCode)

	//test error in authentication
	as.AuthenticateFn = func(username string, pwdPlaintext string, isAdmin bool) (bool, error) {
		return false, errors.New("An error")
	}
	rUser = httptest.NewRequest("POST", "/api/v0/auth/users/login", strings.NewReader(`{"username":"user123","password":"abcd"}`))
	rAdmin = httptest.NewRequest("POST", "/api/v0/auth/admins/login", strings.NewReader(`{"username":"admin123","password":"wsxd"}`))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, rUser)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, rAdmin)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)

	as.AuthenticateFn = func(username string, pwdPlaintext string, isAdmin bool) (bool, error) {
		return true, nil
	}

	//test issuing of authorization fails
	auth.IssueAuthorizationFn = func(au checkin.AuthorizationInfo, w http.ResponseWriter) error {
		return errors.New("An error")
	}
	rUser = httptest.NewRequest("POST", "/api/v0/auth/users/login", strings.NewReader(`{"username":"user123","password":"abcd"}`))
	rAdmin = httptest.NewRequest("POST", "/api/v0/auth/admins/login", strings.NewReader(`{"username":"admin123","password":"wsxd"}`))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, rUser)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, rAdmin)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)

	auth.IssueAuthorizationFn = func(au checkin.AuthorizationInfo, w http.ResponseWriter) error {
		reply, err := json.Marshal(au)
		w.Write(reply)
		return err
	}

	//test updating last logged in fails
	us.UpdateLastLoggedInFn = func(username string) error {
		return errors.New("An error")
	}
	rUser = httptest.NewRequest("POST", "/api/v0/auth/users/login", strings.NewReader(`{"username":"user123","password":"abcd"}`))
	rAdmin = httptest.NewRequest("POST", "/api/v0/auth/admins/login", strings.NewReader(`{"username":"admin123","password":"wsxd"}`))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, rUser)
	test.Equals(t, http.StatusInternalServerError, w.Result().StatusCode)
	ai = checkin.AuthorizationInfo{}
	json.NewDecoder(w.Result().Body).Decode(&ai) //auth credentials should not be written
	test.Equals(t, checkin.AuthorizationInfo{Username: "", IsAdmin: false}, ai)

	//admin endpoint should still work since it doesn't use the userservice
	w = httptest.NewRecorder()
	ai = checkin.AuthorizationInfo{}
	h.ServeHTTP(w, rAdmin)
	test.Equals(t, http.StatusOK, w.Result().StatusCode)
	json.NewDecoder(w.Result().Body).Decode(&ai)
	test.Equals(t, checkin.AuthorizationInfo{Username: "admin123", IsAdmin: true}, ai)

}
