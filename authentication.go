package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var hashCost = 5 //cost must be above 4, the larger you make it the slower the hash function will run
var allowedCorsOrigins = []string{"http://localhost:8080"}
var loginURL = "/api/auth/login"
var signingKey = []byte("theSecretPassword")
var loginTokenUsername, loginTokenExpiryTime = "username", "exp"

type loginDetails struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type response struct {
	Message string `json:"message"`
}

//loginHandler Handles authentication and generation of web tokens in response to the user attempting to login, via /api/auth/login
var loginHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var ld loginDetails
	err := decoder.Decode(&ld)
	if err != nil {
		panic(err)
	}
	fmt.Println(ld.Username)

	if authenticate(ld) {
		jwtToken := createToken(ld)
		reply, _ := json.Marshal(map[string]string{"accessToken": jwtToken})
		w.Write(reply)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		reply, _ := json.Marshal(response{Message: "Incorrect Username or Password"})
		w.Write(reply)
	}

})

func createToken(user loginDetails) string {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims[loginTokenUsername] = user.Username
	claims[loginTokenExpiryTime] = time.Now().Add(time.Hour).Unix()

	tokenSigned, err := token.SignedString(signingKey)
	if err != nil {
		panic(err)
	}

	return tokenSigned
}

func authenticate(user loginDetails) bool {
	//still in testing
	if user.Username == "admin567" {
		return bcrypt.CompareHashAndPassword([]byte("$2a$05$Is.BydwHRaXnXTB5rVFDQerDElDYS6Qbl4KH.T5fVyTvdQHXWNZTS"), []byte(user.Password)) == nil //method returns nil if there is a match between password and hash
		//current password is password
	}
	return false
}

func hashAndSalt(pwd []byte) string {
	// Use GenerateFromPassword to hash & salt pwd.
	//cost must be above 4
	hash, err := bcrypt.GenerateFromPassword(pwd, hashCost)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	// GenerateFromPassword returns a byte slice so we need to
	// convert the bytes to a string and return it
	return string(hash)
}
