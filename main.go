// registrationApp project main.go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"golang.org/x/crypto/bcrypt"
)

var HASH_COST = 5 //cost must be above 4
var ALLOWED_CORS_ORIGINS = []string{"http://localhost:8080"}
var LOGIN_URL = "/api/auth/login"
var signingKey = []byte("theSecretPassword")
var LOGIN_TOKEN_USERNAME, LOGIN_TOKEN_EXPIRY_TIME = "username", "exp"

func main() {
	r := mux.NewRouter()
	r.Handle(LOGIN_URL, loginHandler).Methods("POST")
	handler := cors.New(cors.Options{
		AllowedOrigins: ALLOWED_CORS_ORIGINS,
	}).Handler(r) //only allow GETs POSTs from that address; the bare minimum needed
	http.ListenAndServe(":3000", handler) //PostGres listens on 5432
}

type LoginDetails struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var loginHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var ld LoginDetails
	err := decoder.Decode(&ld)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	fmt.Println(ld.Username)

	if authenticate(ld) {
		jwtToken := createToken(ld)
		w.Write([]byte(jwtToken))
	} else {
		fmt.Println("FAILED AUTHENTICATE")
	}

})

func createToken(user LoginDetails) string {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims[LOGIN_TOKEN_USERNAME] = user.Username
	claims[LOGIN_TOKEN_EXPIRY_TIME] = time.Now().Add(time.Hour).Unix()

	tokenSigned, _ := token.SignedString(signingKey)

	return tokenSigned
}

func authenticate(user LoginDetails) bool {
	//still in testing
	if user.Username == "admin567" {
		return bcrypt.CompareHashAndPassword([]byte("$2a$05$eFNkk5Pdou0I6tzHErvo3ug6VXnercYCMDWdAssVPT0l.BVxLhOXK"), []byte(user.Password)) == nil //method returns nil if there is a match between password and hash
	}
	return false
}

func hashAndSalt(pwd []byte) string {
	// Use GenerateFromPassword to hash & salt pwd.
	//cost must be above 4
	hash, err := bcrypt.GenerateFromPassword(pwd, HASH_COST)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	// GenerateFromPassword returns a byte slice so we need to
	// convert the bytes to a string and return it
	return string(hash)
}

var NotImplemented = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Not Implemented")
})
