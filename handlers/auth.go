package handlers

import (
	"checkin/auth"
	"checkin/response"
	"checkin/users"
	"encoding/json"
	"log"
	"net/http"
)

//AdminLogin Handles administrator log in
var AdminLogin = login(auth.Admin)

//UserLogin Handles user log in
var UserLogin = login(auth.User)

func login(status auth.AdminStatus) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var ld auth.LoginDetails
		err := decoder.Decode(&ld)
		if err != nil {
			log.Println("Login faced an error decoding JSON: " + err.Error())
			response.WriteMessage(http.StatusBadRequest, "Authentication JSON malformed", w)
			return
		}

		isAuthenticated, err := ld.Authenticate(status)

		if err != nil {
			log.Println("Login faced an error in authentication: " + err.Error())
			response.WriteMessage(http.StatusInternalServerError, "Authentication failed due to server error", w)
			return
		}

		if isAuthenticated {
			jwtToken, err := ld.CreateToken(status)
			if err != nil {
				log.Println("Login faced an error in token creation: " + err.Error())
				response.WriteMessage(http.StatusInternalServerError, "Token creation failed", w)
			} else {
				if status == auth.User {
					if err := users.UpdateLastLoggedIn(ld.Username); err != nil {
						log.Println("Login faced an error in updated last logged in: " + err.Error())
						response.WriteMessage(http.StatusInternalServerError, "Error updating last logged in", w)
						return
					}
				}
				reply, _ := json.Marshal(map[string]string{"accessToken": jwtToken})
				w.Write(reply)
			}
		} else {
			response.WriteMessage(http.StatusUnauthorized, "Incorrect Username or Password", w)
		}
	})
}
