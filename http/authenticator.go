package http

import (
	"checkin"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

//Authenticator An interface to check if a given request is authorized, and to issue
//authorization to a given destination
type Authenticator interface {
	IssueAuthorization(au checkin.AuthorizationInfo, w http.ResponseWriter) error
	GetAuthInfo(r *http.Request) (checkin.AuthorizationInfo, error)
	Authenticate(r *http.Request) (bool, error)
}

//CheckAuth an adapter generator which checks if the request has valid authentication
func checkAuth(au Authenticator) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ok, err := au.Authenticate(r)
			if err != nil {
				log.Println("Error in authentication: " + err.Error())
				WriteMessage(http.StatusBadRequest, "Authorization token in invalid format", w)
			} else if !ok {
				WriteMessage(http.StatusUnauthorized, "Invalid Token", w)
			} else {
				h.ServeHTTP(w, r)
			}
		})
	}
}

//isAdmin Allows the handler to serve the request only if it is admin authorized
func isAdmin(au Authenticator) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authDetails, err := au.GetAuthInfo(r)
			if err != nil {
				log.Println("Error in fetching authorization info: " + err.Error())
				WriteMessage(http.StatusBadRequest, "Authorization could not be deciphered", w)
			} else if authDetails.IsAdmin {
				h.ServeHTTP(w, r)
			} else {
				WriteMessage(http.StatusForbidden, "Access Denied", w)
			}
		})
	}
}

//isAdminOrHost Allows the handler to serve the request only if it is admin authorized, or
//if the username attached to the request is the host of the event
//Needs an authenticator, an event service, and a string indicating what mux placeholder
//is used to store the eventID
func isAdminOrHost(au Authenticator, es checkin.EventService, eventIDKey string) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authDetails, err := au.GetAuthInfo(r)
			if err != nil {
				log.Println("Error in fetching authorization info: " + err.Error())
				WriteMessage(http.StatusBadRequest, "Authorization could not be deciphered", w)
			} else if authDetails.IsAdmin {
				h.ServeHTTP(w, r)
			} else {
				eventID := mux.Vars(r)[eventIDKey]
				ok, err := es.CheckHost(authDetails.Username, eventID)
				if err != nil {
					log.Println("Error in checking if user is host: " + err.Error())
					WriteMessage(http.StatusInternalServerError, "Error checking host", w)
				} else if ok {
					h.ServeHTTP(w, r)
				} else {
					WriteMessage(http.StatusForbidden, "Access Denied", w)
				}

			}
		})
	}
}

func isAdminOrUser(au Authenticator, us checkin.UserService, usernameKey string) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authDetails, err := au.GetAuthInfo(r)
			if err != nil {
				log.Println("Error in fetching authorization info: " + err.Error())
				WriteMessage(http.StatusBadRequest, "Authorization could not be deciphered", w)
			} else if authDetails.IsAdmin || authDetails.Username == mux.Vars(r)[usernameKey] {
				h.ServeHTTP(w, r)
			} else {
				WriteMessage(http.StatusForbidden, "Access Denied", w)
			}
		})
	}
}
