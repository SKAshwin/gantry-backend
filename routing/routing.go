package routing

import (
	"errors"
	"log"
	"net/http"
	"registration-app/handlers"
	"registration-app/response"

	"github.com/gorilla/mux"
)

//SetUp Returns a router configured with all the handlers for each endpoint
func SetUp() *mux.Router {
	r := mux.NewRouter()
	r.Handle("/api/admins/auth/login", handlers.AdminLogin).Methods("POST")
	r.Handle("/api/users/auth/login", handlers.UserLogin).Methods("POST")

	r.Handle("/api/users", Validate(handlers.ListUsers, IsAdmin)).Methods("GET")
	r.Handle("/api/users", Validate(handlers.CreateUser, IsAdmin)).Methods("POST")
	r.Handle("/api/users/{username}", handlers.UserExists(Validate(handlers.GetUser, IsSpecifiedUser, IsAdmin))).Methods("GET")
	r.Handle("/api/users/{username}", handlers.UserExists(Validate(handlers.UpdateUser, IsSpecifiedUser, IsAdmin))).Methods("PUT")
	r.Handle("/api/users/{username}", handlers.UserExists(Validate(handlers.DeleteUser, IsSpecifiedUser, IsAdmin))).Methods("DELETE")

	r.Handle("/api/events", Validate(handlers.GetUsersEvents)).Methods("GET")
	r.Handle("/api/events", Validate(handlers.CreateEvent)).Methods("POST")
	r.Handle("/api/events/exists/{eventURL}", Validate(handlers.EventURLAvailable)).Methods("GET")
	r.Handle("/api/events/{eventID}", handlers.EventExists(Validate(handlers.GetEvent, IsHost, IsAdmin))).Methods("GET")
	r.Handle("/api/events/{eventID}", handlers.EventExists(Validate(handlers.UpdateEvent, IsHost, IsAdmin))).Methods("PUT")
	r.Handle("/api/events/{eventID}", handlers.EventExists(Validate(handlers.DeleteEvent, IsHost, IsAdmin))).Methods("DELETE")

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
