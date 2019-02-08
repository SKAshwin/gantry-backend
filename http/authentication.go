package http

import (
	"checkin"
	"log"
	"net/http"
)

//Authenticator An interface to check if a given request is authorized, and to issue
//authorization to a given destination
type Authenticator interface {
	IssueAuthorization(au checkin.AuthorizationInfo, w http.ResponseWriter)
	GetAuthInfo(r *http.Request) (checkin.AuthorizationInfo, error)
	Authenticate(r *http.Request) (bool, error)
}

//CheckAuth an adapter generator which checks if the request has valid authentication
func CheckAuth(au Authenticator) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ok, err := au.Authenticate(r)
			if err != nil {
				log.Println("Error in parsing JWT: " + err.Error())
				WriteMessage(http.StatusBadRequest, "Authorization token in invalid format", w)
			} else if !ok {
				WriteMessage(http.StatusUnauthorized, "Invalid Token", w)
			} else {
				h.ServeHTTP(w, r)
			}
		})
	}
}
