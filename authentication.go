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
	if err != nil {
		log.Println(err.Error())
		writeError(http.StatusBadRequest, "Authentication JSON malformed", w)
		return
	}
	fmt.Println(ld.Username)

	isAuthenticated, err := authenticate(ld, adminTable)

	if err != nil {
		writeError(http.StatusInternalServerError, "Authentication failed due to server error", w)
		return
	}

	if isAuthenticated {
		jwtToken, err := createToken(ld, true)
		if err != nil {
			writeError(http.StatusInternalServerError, "Token creation failed", w)
		} else {
			reply, _ := json.Marshal(map[string]string{"accessToken": jwtToken})
			w.Write(reply)
		}
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		reply, _ := json.Marshal(response{Message: "Incorrect Username or Password"})
		w.Write(reply)
	}

})

func createToken(user loginDetails, isAdmin bool) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims[jwtUsername] = user.Username
	claims[jwtExpiryTime] = time.Now().Add(time.Hour).Unix()
	claims[jwtAdminStatus] = isAdmin

	tokenSigned, err := token.SignedString(signingKey)
	if err != nil {
		log.Println(err.Error())
		return "", err
	}

	return tokenSigned, nil
}

func authenticate(user loginDetails, tableName string) (bool, error) {
	var stmt *sql.Stmt
	var err error
	if tableName == adminTable {
		stmt, err = db.Prepare("SELECT passwordHash FROM app_admin where username = $1")
	} else if tableName == userTable {
		stmt, err = db.Prepare("SELECT passwordHash FROM app_user where username = $1")
	}
	if err != nil {
		log.Println("Statement preparation in authentication failed: ", err.Error())
		return false, err
	}
	var passwordHash string
	err = stmt.QueryRow(user.Username).Scan(&passwordHash)
	if err == sql.ErrNoRows {
		return false, nil //no such username exists
	} else if err != nil {
		//any other error represents a failure
		log.Println("Database Querying in Authentication Failed: ", err.Error())
		return false, err
	}
	return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(user.Password)) == nil, nil //method returns nil if there is a match between password and hash
}

func hashAndSalt(pwd []byte) (string, error) {
	// Use GenerateFromPassword to hash & salt pwd.
	//cost must be above 4
	hash, err := bcrypt.GenerateFromPassword(pwd, hashCost)
	if err != nil {
		return "", err
	}
	// GenerateFromPassword returns a byte slice so we need to
	// convert the bytes to a string and return it
	return string(hash), nil
}
