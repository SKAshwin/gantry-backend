package jwt

import (
	"checkin"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
)

//Authenticator is an implementation of Authenticator (see http) which uses java web tokens
//for authentication
type Authenticator struct {
	SigningKey []byte
}

//Authenticate returns true if the request has a valid (non-expired) token
//Returns false if the token is expired or otherwise invalid
//Returns false and error if the token isn't even formatted correctly
func (jwta *Authenticator) Authenticate(r *http.Request) (bool, error) {
	jwtString, err := getJWTString(r)
	if err != nil {
		return false, errors.New("Error in authenticate: " + err.Error())
	}

	_, err = jwt.Parse(jwtString, jwta.keyGetter)
	if err != nil { //err is nil if token is invalid
		return false, nil
	}
	return true, nil
}

//GetAuthInfo From what is ASSUMED TO BE A REQUEST CONTAINING A VALID WEB TOKEN
//Return the authorization information
//Will panic if assumption fails
func (jwta *Authenticator) GetAuthInfo(r *http.Request) checkin.AuthorizationInfo {
	jwtString, _ := getJWTString(r)
	token, _ := jwt.Parse(jwtString, jwta.keyGetter)
	claims, _ := token.Claims.(jwt.MapClaims)
	return checkin.AuthorizationInfo{Username: claims["username"].(string),
		IsAdmin: claims["isAdmin"].(bool)}
}

//keyGetter checks if the provided token follows the appropriate signing method
//Returns an error if not
//Returns the signing key if it does follow the appropriate method
func (jwta *Authenticator) keyGetter(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		log.Printf("Unexpected signing method: %v \n", token.Header["alg"])
		return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
	}
	return jwta.SigningKey, nil
}

//getJWTString extracts the JWT string from the Authorization header, in a Bearer {{token}} format
//Returns an error If there is no JWT string, or the header's value is in an invalid format
func getJWTString(r *http.Request) (string, error) {
	reqToken := r.Header.Get("Authorization")
	if reqToken == "" {
		return "", errors.New("No string in authorization header")
	}
	splitToken := strings.Split(reqToken, "Bearer ")
	if len(splitToken) != 2 {
		return "", errors.New("Authorization header value not in Bearer {{token}} format")
	}
	reqToken = splitToken[1]
	return reqToken, nil
}
