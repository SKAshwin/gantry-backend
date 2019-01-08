package main

import (
	"errors"
	"log"
	"net/http"
	"registration-app/handlers"

	"registration-app/auth"
	"registration-app/event"
	"registration-app/response"

	"github.com/gorilla/mux"
)

//SetUpRouting Returns a router configured with all the handlers for each endpoint
func SetUpRouting() *mux.Router {
	r := mux.NewRouter()
	r.Handle("/api/admins/auth/login", handlers.AdminLogin).Methods("POST")
	r.Handle("/api/users/auth/login", handlers.UserLogin).Methods("POST")

	r.Handle("/api/users", auth.AccessControl(handlers.ListUsers, auth.IsAdmin)).Methods("GET")
	r.Handle("/api/users", auth.AccessControl(handlers.CreateUser, auth.IsAdmin)).Methods("POST")
	r.Handle("/api/users/{username}", handlers.UserExists(auth.AccessControl(handlers.GetUser, auth.IsSpecifiedUser, auth.IsAdmin))).Methods("GET")
	r.Handle("/api/users/{username}", handlers.UserExists(auth.AccessControl(handlers.UpdateUser, auth.IsSpecifiedUser, auth.IsAdmin))).Methods("PUT")
	r.Handle("/api/users/{username}", handlers.UserExists(auth.AccessControl(handlers.DeleteUser, auth.IsSpecifiedUser, auth.IsAdmin))).Methods("DELETE")

	r.Handle("/api/events", auth.AccessControl(handlers.GetUsersEvents)).Methods("GET")
	r.Handle("/api/events", auth.AccessControl(handlers.CreateEvent)).Methods("POST")
	r.Handle("/api/events/exists/{eventURL}", auth.AccessControl(handlers.EventURLAvailable)).Methods("GET")
	r.Handle("/api/events/{eventID}", handlers.EventExists(auth.AccessControl(handlers.GetEvent, event.IsHost, auth.IsAdmin))).Methods("GET")
	r.Handle("/api/events/{eventID}", handlers.EventExists(auth.AccessControl(handlers.UpdateEvent, event.IsHost, auth.IsAdmin))).Methods("PUT")
	r.Handle("/api/events/{eventID}", handlers.EventExists(auth.AccessControl(handlers.DeleteEvent, event.IsHost, auth.IsAdmin))).Methods("DELETE")

	return r
}

//RecoverWrap Middleware which handles any panic in the handlers by logging it and reporting an internal
//server error to the front end
//intended for unexpected errors (expected, handled errors should not cause a panic)
func RecoverWrap(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		defer func() {
			r := recover()
			if r != nil {
				switch t := r.(type) {
				case string:
					err = errors.New(t)
				case error:
					err = t
				default:
					err = errors.New("Unknown error")
				}
				log.Println("Entered Panic with ", err.Error())
				response.WriteMessage(http.StatusInternalServerError, err.Error(), w)
			}
		}()
		h.ServeHTTP(w, r)
	})
}
