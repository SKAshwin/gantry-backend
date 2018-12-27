package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func SetUpRouting() *mux.Router {
	r := mux.NewRouter()
	r.Handle("/api/auth/login", AdminLoginHandler).Methods("POST")
	r.Handle("/api/users", AccessControl(true, ListUsersHandler)).Methods("GET")
	r.Handle("/api/users", AccessControl(true, CreateUserHandler)).Methods("POST")
	r.Handle("/api/users/{username}", UserExists(AccessControl(true, GetUserHandler))).Methods("GET")
	r.Handle("/api/users/{username}", UserExists(AccessControl(true, UpdateUserDetailsHandler))).Methods("PUT")
	r.Handle("/api/users/{username}", UserExists(AccessControl(true, DeleteUserHandler))).Methods("DELETE")
	// /users/profile

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
				WriteMessage(http.StatusInternalServerError, err.Error(), w)
			}
		}()
		h.ServeHTTP(w, r)
	})
}
