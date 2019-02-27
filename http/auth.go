package http

import (
	"checkin"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

//AuthHandler is an extension of mux.Router which handles all authentication-related
//requests. It needs an AuthenticationService and a logger
//Also needs a UserService to update the last logged in status of the user
//Needs an Authenticator to send the client its authentication tokens
type AuthHandler struct {
	*mux.Router
	AuthService   checkin.AuthenticationService
	UserService   checkin.UserService
	Authenticator Authenticator
	Logger        *log.Logger
}

//NewAuthHandler creates a new AuthHandler which uses the given authentication service to check
//authentication, the given authenticator to issue authorization to the client
//and the given user service to update the last logged in of the User
func NewAuthHandler(as checkin.AuthenticationService, auth Authenticator, us checkin.UserService) *AuthHandler {
	h := &AuthHandler{
		Router:        mux.NewRouter(),
		Logger:        log.New(os.Stderr, "", log.LstdFlags),
		AuthService:   as,
		UserService:   us,
		Authenticator: auth,
	}
	h.Handle("/api/v0/auth/admins/login", http.HandlerFunc(h.handleLogin(true))).Methods("POST")
	h.Handle("/api/v0/auth/users/login", http.HandlerFunc(h.handleLogin(false))).Methods("POST")
	return h
}

//handleLogin is a method which returns a http HandlerFunc which
//handles logging in for either users or administrators, depending
//on the bool value supplied
func (h *AuthHandler) handleLogin(isAdmin bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var loginDetails map[string]string
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&loginDetails)
		if err != nil {
			h.Logger.Println("Login faced an error decoding JSON: " + err.Error())
			WriteMessage(http.StatusBadRequest, "Authentication JSON malformed", w)
			return
		}

		isAuthenticated, err := h.AuthService.Authenticate(loginDetails["username"],
			loginDetails["password"], isAdmin)

		if err != nil {
			h.Logger.Println("Login faced an error in authentication: " + err.Error())
			WriteMessage(http.StatusInternalServerError, "Authentication failed due to server error", w)
			return
		}

		if isAuthenticated {
			if !isAdmin {
				if err := h.UserService.UpdateLastLoggedIn(loginDetails["username"]); err != nil {
					h.Logger.Println("Login faced an error in updated last logged in: " + err.Error())
					WriteMessage(http.StatusInternalServerError, "Error updating last logged in", w)
					return
				}
			}
			err := h.Authenticator.IssueAuthorization(checkin.AuthorizationInfo{
				Username: loginDetails["username"],
				IsAdmin:  isAdmin,
			}, w)
			if err != nil {
				h.Logger.Println("Login faced an error in token creation: " + err.Error())
				WriteMessage(http.StatusInternalServerError, "Token creation failed", w)
			}
		} else {
			WriteMessage(http.StatusUnauthorized, "Incorrect Username or Password", w)
		}
	}
}
