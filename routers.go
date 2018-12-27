package main

import (
	"errors"
	"log"
	"net/http"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

func SetUpRouting() *mux.Router {
	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		//check if jwt is sent, extracts information
		//also checks if token is expired; returns 401 if not
		ValidationKeyGetter: KeyGetter,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err string) {
			WriteMessage(http.StatusUnauthorized, err, w)
		},
		SigningMethod: jwt.SigningMethodHS256,
	})

	r := mux.NewRouter()
	r.Handle("/api/auth/login", AdminLoginHandler).Methods("POST")
	r.Handle("/api/users", IsAdmin(jwtMiddleware.Handler(ListUsersHandler))).Methods("GET")
	r.Handle("/api/users", IsAdmin(jwtMiddleware.Handler(CreateUserHandler))).Methods("POST")
	r.Handle("/api/users/{username}", UserExists(IsAdmin(jwtMiddleware.Handler(GetUserHandler)))).Methods("GET")
	r.Handle("/api/users/{username}", UserExists(IsAdmin(jwtMiddleware.Handler(UpdateUserDetailsHandler)))).Methods("PUT")
	r.Handle("/api/users/{username}", UserExists(IsAdmin(jwtMiddleware.Handler(DeleteUserHandler)))).Methods("DELETE")
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
