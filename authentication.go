package main

import (
	"database/sql"
	"encoding/json"
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
		writeError(http.StatusUnauthorized, "Incorrect Username or Password", w)
	}

})

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

func keyGetter(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		log.Printf("Unexpected signing method: %v \n", token.Header["alg"])
		return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
	}
	return signingKey, nil
}

func getJWTClaims(r *http.Request) (map[string]interface{}, bool) {
	reqToken := r.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	reqToken = splitToken[1]
	claims, ok := extractClaimsFromTokenString(reqToken)
	return claims, ok

}

func extractClaimsFromTokenString(tokenStr string) (jwt.MapClaims, bool) {
	token, err := jwt.Parse(tokenStr, keyGetter)

	if err != nil {
		log.Println("Token parsing failed: ", err.Error())
		return nil, false
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, true
	}
	log.Println("Expired JWT Token")
	return nil, false
}
