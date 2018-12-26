package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
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
		log.Println("adminLoginHandler faced an error: " + err.Error())
		writeError(http.StatusBadRequest, "Authentication JSON malformed", w)
		return
	}
	fmt.Println(ld.Username)

	isAuthenticated, err := authenticate(ld, adminTable)

	if err != nil {
		log.Println("adminLoginHandler faced an error: " + err.Error())
		writeError(http.StatusInternalServerError, "Authentication failed due to server error", w)
		return
	}

	if isAuthenticated {
		jwtToken, err := createToken(ld, true)
		if err != nil {
			log.Println("adminLoginHandler faced an error: " + err.Error())
			writeError(http.StatusInternalServerError, "Token creation failed", w)
		} else {
			reply, _ := json.Marshal(map[string]string{"accessToken": jwtToken})
			w.Write(reply)
		}
	} else {
		writeError(http.StatusUnauthorized, "Incorrect Username or Password", w)
	}

})

//authenticate Given a user's login details and a table name (indicating whether they are admin or users)
//Returns true if the password matches that username, false otherwise
//Returns error if there is a database querying issue
func authenticate(user loginDetails, tableName string) (bool, error) {
	var stmt *sql.Stmt
	var err error
	if tableName == adminTable {
		stmt, err = db.Prepare("SELECT passwordHash FROM app_admin where username = $1")
	} else if tableName == userTable {
		stmt, err = db.Prepare("SELECT passwordHash FROM app_user where username = $1")
	}
	if err != nil {
		return false, errors.New("Statement preparation in authentication failed: " + err.Error())
	}
	var passwordHash string
	err = stmt.QueryRow(user.Username).Scan(&passwordHash)
	if err == sql.ErrNoRows {
		return false, nil //no such username exists
	} else if err != nil {
		//any other error represents a failure
		return false, errors.New("Database Querying in Authentication Failed: " + err.Error())
	}
	return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(user.Password)) == nil, nil //method returns nil if there is a match between password and hash
}

//hashAndSalt given a password (as []byte), uses the bcrypt method to hash a password
//Returns a string containing both the salt and the hash (use bcrypt library to work with it)
func hashAndSalt(pwd []byte) (string, error) {
	//Use GenerateFromPassword to hash & salt pwd.
	//cost must be above 4
	hash, err := bcrypt.GenerateFromPassword(pwd, hashCost)
	if err != nil {
		return "", errors.New("Failed to hash password: " + err.Error())
	}
	// GenerateFromPassword returns a byte slice so we need to
	// convert the bytes to a string and return it
	return string(hash), nil
}

//createToken Given a user (or admin's) login details, returns an encrypted web token
//Uses signing method HS256
func createToken(user loginDetails, isAdmin bool) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims[jwtUsername] = user.Username
	claims[jwtExpiryTime] = time.Now().Add(time.Hour).Unix()
	claims[jwtAdminStatus] = isAdmin //token claim to be given out if user is logging in as admin (through internal console)

	tokenSigned, err := token.SignedString(signingKey)
	if err != nil {
		return "", errors.New("Token signing failed during creation: " + err.Error())
	}

	return tokenSigned, nil
}

//keyGetter checks if the provided token follows the appropriate signing method
//Returns an error if not
//Returns the signing key if it does follow the appropriate method
func keyGetter(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		log.Printf("Unexpected signing method: %v \n", token.Header["alg"])
		return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
	}
	return signingKey, nil
}

//getJWTClaims Given a http request, extracts the encrypted JWT string from the authorization header
//Returns a map between the JWT claims and their values
//Returns an error if either token parsing failed (possibly incorrect signing method etc) or if the token is expired
func getJWTClaims(r *http.Request) (map[string]interface{}, error) {
	reqToken := r.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	reqToken = splitToken[1]
	claims, err := extractClaimsFromTokenString(reqToken)
	return claims, err

}

//extractClaimsFromTokenString given the encrypted JWT string (usually taken from the authorization header)
//Returns a jwt.MapClaims object representing the claims in the token
//Returns a non-nil error if either token parsing failed, or the token was expired
func extractClaimsFromTokenString(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, keyGetter)

	if err != nil {
		return nil, errors.New("Token parsing failed during extraction of claims: " + err.Error())
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("Expired token")
}
