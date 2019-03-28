package jwt_test

import (
	"checkin"
	myjwt "checkin/http/jwt"
	"checkin/test"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

func TestValidJWT(t *testing.T) {
	jwta := myjwt.Authenticator{
		SigningKey: []byte("SomePassword"),
		ExpiryTime: time.Hour,
	}
	w := httptest.NewRecorder()
	authInfo := checkin.AuthorizationInfo{
		Username: "Jim",
		IsAdmin:  false,
	}
	err := jwta.IssueAuthorization(authInfo, w)
	test.Ok(t, err)
	var reply map[string]string
	json.NewDecoder(w.Result().Body).Decode(&reply)

	r := httptest.NewRequest("POST", "/hello/some/url", nil)
	r.Header.Set("Authorization", "Bearer "+reply["accessToken"])
	auth, err := jwta.Authenticate(r)
	test.Ok(t, err)
	test.Assert(t, auth, "Expected successful authentication, did not get it")

	ai, err := jwta.GetAuthInfo(r)
	test.Ok(t, err)
	test.Equals(t, authInfo, ai)

	authInfo = checkin.AuthorizationInfo{} //empty value should still totally work
	w = httptest.NewRecorder()
	err = jwta.IssueAuthorization(authInfo, w)
	test.Ok(t, err)
	json.NewDecoder(w.Result().Body).Decode(&reply)

	r.Header.Set("Authorization", "Bearer "+reply["accessToken"])
	auth, err = jwta.Authenticate(r)
	test.Ok(t, err)
	test.Assert(t, auth, "Expected successful authentication, did not get it")

	ai, err = jwta.GetAuthInfo(r)
	test.Ok(t, err)
	test.Equals(t, authInfo, ai)

	//an empty string JWT should still work
	jwta = myjwt.Authenticator{
		SigningKey: []byte(""),
		ExpiryTime: time.Hour,
	}
	w = httptest.NewRecorder()
	err = jwta.IssueAuthorization(authInfo, w)
	test.Ok(t, err)
	json.NewDecoder(w.Result().Body).Decode(&reply)
	r = httptest.NewRequest("POST", "/hello/some/url", nil)
	r.Header.Set("Authorization", "Bearer "+reply["accessToken"])
	auth, err = jwta.Authenticate(r)
	test.Ok(t, err)
	test.Assert(t, auth, "Expected successful authentication, did not get it")

	ai, err = jwta.GetAuthInfo(r)
	test.Ok(t, err)
	test.Equals(t, authInfo, ai)

	//how are empty nil signing keys dealt with
	//pls no panic
	jwta = myjwt.Authenticator{
		SigningKey: nil,
		ExpiryTime: time.Hour,
	}
	w = httptest.NewRecorder()
	authInfo = checkin.AuthorizationInfo{
		Username: "Jim",
		IsAdmin:  false,
	}
	err = jwta.IssueAuthorization(authInfo, w)
	test.Ok(t, err)
	json.NewDecoder(w.Result().Body).Decode(&reply)
	r = httptest.NewRequest("POST", "/hello/some/url", nil)
	r.Header.Set("Authorization", "Bearer "+reply["accessToken"])
	auth, err = jwta.Authenticate(r)
	test.Ok(t, err)
	test.Assert(t, auth, "Expected successful authentication, did not get it")

	ai, err = jwta.GetAuthInfo(r)
	test.Ok(t, err)
	test.Equals(t, authInfo, ai)
}

func TestExpiredJWT(t *testing.T) {
	jwta := myjwt.Authenticator{
		SigningKey: []byte("SomePassword"),
		ExpiryTime: -1 * time.Hour,
	}
	w := httptest.NewRecorder()
	authInfo := checkin.AuthorizationInfo{
		Username: "Jim",
		IsAdmin:  false,
	}
	err := jwta.IssueAuthorization(authInfo, w)
	test.Ok(t, err)
	var reply map[string]string
	json.NewDecoder(w.Result().Body).Decode(&reply)

	r := httptest.NewRequest("POST", "/hello/some/url", nil)
	r.Header.Set("Authorization", "Bearer "+reply["accessToken"])
	auth, err := jwta.Authenticate(r)
	test.Ok(t, err)
	test.Assert(t, !auth, "Expected unsuccessful authentication, was successful")

	ai, err := jwta.GetAuthInfo(r)
	test.Assert(t, err != nil, "Expected non-nil error, got nil")
	test.Equals(t, checkin.AuthorizationInfo{}, ai)
}

func TestBadFormatJWT(t *testing.T) {
	jwta := myjwt.Authenticator{
		SigningKey: []byte("SomePassword"),
		ExpiryTime: 1 * time.Hour,
	}
	r := httptest.NewRequest("POST", "/hello/some/url", nil)
	r.Header.Set("Authorization", "Bearer randomStringWoahH")
	auth, err := jwta.Authenticate(r)
	test.Ok(t, err)
	test.Assert(t, !auth, "Expected unsuccessful authentication, was successful")

	ai, err := jwta.GetAuthInfo(r)
	test.Assert(t, err != nil, "Expected non-nil error, got nil")
	test.Equals(t, checkin.AuthorizationInfo{}, ai)

	r.Header.Set("Authorization", "notevenaformat")
	auth, err = jwta.Authenticate(r)
	test.Assert(t, err != nil, "Expected non-nil error, got nil")

	ai, err = jwta.GetAuthInfo(r)
	test.Assert(t, err != nil, "Expected non-nil error, got nil")
	test.Equals(t, checkin.AuthorizationInfo{}, ai)

	//authorization header missing
	auth, err = jwta.Authenticate(r)
	test.Assert(t, err != nil, "Expected non-nil error, got nil")

	ai, err = jwta.GetAuthInfo(r)
	test.Assert(t, err != nil, "Expected non-nil error, got nil")
	test.Equals(t, checkin.AuthorizationInfo{}, ai)

	//now try different signing method of the same class
	//Should not throw an error, but authentication will fail
	token := jwt.New(jwt.SigningMethodHS384)
	claims := token.Claims.(jwt.MapClaims)
	au := checkin.AuthorizationInfo{
		Username: "Jim",
		IsAdmin:  false,
	}
	claims["username"] = au.Username
	claims["exp"] = time.Now().Add(jwta.ExpiryTime).Unix()
	claims["password"] = au.IsAdmin
	tokenSigned, err := token.SignedString(jwta.SigningKey)
	test.Ok(t, err)

	r.Header.Set("Authorization", "Bearer "+tokenSigned)
	auth, err = jwta.Authenticate(r)
	test.Assert(t, !auth, "Expected no authentication")

	ai, err = jwta.GetAuthInfo(r)
	test.Assert(t, err != nil, "Expected non-nil error, got nil")
	test.Equals(t, checkin.AuthorizationInfo{}, ai)

}

func TestBadJWTAuthenticator(t *testing.T) {

}
