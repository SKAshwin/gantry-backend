package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"registration-app/auth"
	"registration-app/response"
)

//AdminLoginHandler Handles authentication and generation of web tokens in response to the user attempting to login, via /api/auth/login
var AdminLogin = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var ld auth.LoginDetails
	err := decoder.Decode(&ld)
	if err != nil {
		log.Println("adminLoginHandler faced an error: " + err.Error())
		response.WriteMessage(http.StatusBadRequest, "Authentication JSON malformed", w)
		return
	}
	fmt.Println(ld.Username)

	isAuthenticated, err := ld.Authenticate(auth.Admin)

	if err != nil {
		log.Println("adminLoginHandler faced an error: " + err.Error())
		response.WriteMessage(http.StatusInternalServerError, "Authentication failed due to server error", w)
		return
	}

	if isAuthenticated {
		jwtToken, err := ld.CreateToken(auth.Admin)
		if err != nil {
			log.Println("AdminLoginHandler faced an error: " + err.Error())
			response.WriteMessage(http.StatusInternalServerError, "Token creation failed", w)
		} else {
			reply, _ := json.Marshal(map[string]string{"accessToken": jwtToken})
			w.Write(reply)
		}
	} else {
		response.WriteMessage(http.StatusUnauthorized, "Incorrect Username or Password", w)
	}

})
