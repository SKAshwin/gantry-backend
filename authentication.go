package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	//jwtRequest "github.com/dgrijalva/jwt-go/request"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

const hashCost = 5 //cost must be above 4, the larger you make it the slower the hash function will run
var signingKey = []byte("theSecretPassword")

const jwtUsername, jwtExpiryTime, jwtAdminStatus = "username", "exp", "admin"
const adminTable = "app_admin"
const userTable = "app_user"

type loginDetails struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

//adminLoginHandler Handles authentication and generation of web tokens in response to the user attempting to login, via /api/auth/login
var adminLoginHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var ld loginDetails
	err := decoder.Decode(&ld)
	panicIf(err)
	fmt.Println(ld.Username)

	if authenticate(ld, adminTable) {
		jwtToken := createToken(ld, true)
		reply, _ := json.Marshal(map[string]string{"accessToken": jwtToken})
		w.Write(reply)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		reply, _ := json.Marshal(response{Message: "Incorrect Username or Password"})
		w.Write(reply)
	}

})

func createToken(user loginDetails, isAdmin bool) string {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims[jwtUsername] = user.Username
	claims[jwtExpiryTime] = time.Now().Add(time.Hour).Unix()
	claims[jwtAdminStatus] = isAdmin

	tokenSigned, err := token.SignedString(signingKey)
	panicIf(err)

	return tokenSigned
}

func authenticate(user loginDetails, tableName string) bool {
	var stmt *sql.Stmt
	var err error
	if tableName == adminTable {
		stmt, err = db.Prepare("SELECT passwordHash FROM app_admin where username = $1")
	} else if tableName == userTable {
		stmt, err = db.Prepare("SELECT passwordHash FROM app_user where username = $1")
	}
	panicIf(err)
	var passwordHash string
	err = stmt.QueryRow(user.Username).Scan(&passwordHash)
	if err == sql.ErrNoRows {
		return false //no such username exists
	}
	panicIf(err)                                                                             //any other error should be panicked on
	return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(user.Password)) == nil //method returns nil if there is a match between password and hash
}

func hashAndSalt(pwd []byte) string {
	// Use GenerateFromPassword to hash & salt pwd.
	//cost must be above 4
	hash, err := bcrypt.GenerateFromPassword(pwd, hashCost)
	if err != nil {
		log.Println(err)
		panic(err)
	}
	// GenerateFromPassword returns a byte slice so we need to
	// convert the bytes to a string and return it
	return string(hash)
}
