package mock

import (
	"checkin"
	"net/http"
)

//Authenticator a mock implementation of http.Authenticator
type Authenticator struct {
	IssueAuthorizationFn      func(au checkin.AuthorizationInfo, w http.ResponseWriter) error
	IssueAuthorizationInvoked bool

	GetAuthInfoFn      func(r *http.Request) (checkin.AuthorizationInfo, error)
	GetAuthInfoInvoked bool

	AuthenticateFn      func(r *http.Request) (bool, error)
	AuthenticateInvoked bool
}

//IssueAuthorization calls the injected function and marks it as invoked
func (a *Authenticator) IssueAuthorization(au checkin.AuthorizationInfo, w http.ResponseWriter) error {
	a.IssueAuthorizationInvoked = true
	return a.IssueAuthorizationFn(au, w)
}

//GetAuthInfo calls the injected function and marks it as invoked
func (a *Authenticator) GetAuthInfo(r *http.Request) (checkin.AuthorizationInfo, error) {
	a.GetAuthInfoInvoked = true
	return a.GetAuthInfoFn(r)
}

//Authenticate calls the injected function and marks it as invoked
func (a *Authenticator) Authenticate(r *http.Request) (bool, error) {
	a.AuthenticateInvoked = true
	return a.AuthenticateFn(r)
}
