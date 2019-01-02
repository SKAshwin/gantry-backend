package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"registration-app/config"

	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

const hashCost = 5 //cost must be above 4, the larger you make it the slower the hash function will run
var signingKey = []byte("theSecretPassword")

const jwtUsername, jwtExpiryTime, jwtAdminStatus, jwtID = "username", "exp", "admin", "userID"

type AdminStatus int

const (
	Admin = iota
	User
)

type LoginDetails struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

//Authenticate Given a user's login details and a table name (indicating whether they are admin or users)
//Returns true if the password matches that username, false otherwise
//Returns error if there is a database querying issue
func (user LoginDetails) Authenticate(as AdminStatus) (bool, error) {
	var stmt *sql.Stmt
	var err error
	if as == Admin {
		stmt, err = config.DB.Prepare("SELECT passwordHash FROM app_admin where username = $1")
	} else if as == User {
		stmt, err = config.DB.Prepare("SELECT passwordHash FROM app_user where username = $1")
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

//HashAndSalt given a password (as []byte), uses the bcrypt method to hash a password
//Returns a string containing both the salt and the hash (use bcrypt library to work with it)
func HashAndSalt(pwd []byte) (string, error) {
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

//GetID For a given username (used in the log in process)
//Returns the userID/adminID, the primary key used internally
func (user LoginDetails) GetID(as AdminStatus) (string, error) {
	var uuid string
	var err error
	if as == Admin {
		err = config.DB.QueryRow("SELECT ID from app_admin where username = $1", user.Username).Scan(&uuid)
	} else {
		err = config.DB.QueryRow("SELECT ID from app_user where username = $1", user.Username).Scan(&uuid)
	}

	if err != nil {
		return "", errors.New("Could not fetch user ID: " + err.Error())
	}

	return uuid, nil
}

//CreateToken Given a user (or admin's) login details, returns an encrypted web token
//Uses signing method HS256
func (user LoginDetails) CreateToken(as AdminStatus, ID string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims[jwtID] = ID
	claims[jwtUsername] = user.Username
	claims[jwtExpiryTime] = time.Now().Add(time.Hour).Unix()
	claims[jwtAdminStatus] = (as == Admin) //token claim to be given out if user is logging in as admin (through internal console)

	tokenSigned, err := token.SignedString(signingKey)
	if err != nil {
		return "", errors.New("Token signing failed during creation: " + err.Error())
	}

	return tokenSigned, nil
}

//KeyGetter checks if the provided token follows the appropriate signing method
//Returns an error if not
//Returns the signing key if it does follow the appropriate method
func KeyGetter(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		log.Printf("Unexpected signing method: %v \n", token.Header["alg"])
		return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
	}
	return signingKey, nil
}

//GetJWTClaims Given a http request, extracts the encrypted JWT string from the authorization header
//Returns a map between the JWT claims and their values
//Returns an error if either token parsing failed (possibly incorrect signing method etc) or if the token is expired
func GetJWTClaims(r *http.Request) (map[string]interface{}, error) {
	claims, err := ExtractClaimsFromTokenString(GetJWTString(r))
	return claims, err

}

func GetJWTString(r *http.Request) string {
	reqToken := r.Header.Get("Authorization")
	if reqToken == "" {
		return ""
	}
	splitToken := strings.Split(reqToken, "Bearer ")
	reqToken = splitToken[1]
	return reqToken
}

//ExtractClaimsFromTokenString given the encrypted JWT string (usually taken from the authorization header)
//Returns a jwt.MapClaims object representing the claims in the token
//Returns a non-nil error if either token parsing failed, or the token was expired
func ExtractClaimsFromTokenString(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, KeyGetter)

	if err != nil {
		return nil, errors.New("Token parsing failed during extraction of claims: " + err.Error())
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("Expired token")
}
