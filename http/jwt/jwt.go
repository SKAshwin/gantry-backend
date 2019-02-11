package jwt

import (
	"checkin"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

//Authenticator is an implementation of http.Authenticator which uses java web tokens
//for authentication
type Authenticator struct {
	SigningKey []byte
}

const (
	jwtUsername    = "username"
	jwtExpiryTime  = "exp"
	jwtAdminStatus = "admin"
)

//Authenticate returns true if the request has a valid (non-expired) token
//Returns false if the token is expired or otherwise invalid
//Returns false and error if the token isn't even formatted correctly
func (jwta Authenticator) Authenticate(r *http.Request) (bool, error) {
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

//GetAuthInfo From a web token, return the authorization info
//Will return an error if: token string not in the request in the valid format
//token string not parseable, or token invalid
func (jwta Authenticator) GetAuthInfo(r *http.Request) (checkin.AuthorizationInfo, error) {
	jwtString, err := getJWTString(r)
	if err != nil {
		return checkin.AuthorizationInfo{}, errors.New("Error extracting token string: " + err.Error())
	}
	token, err := jwt.Parse(jwtString, jwta.keyGetter)
	if err != nil {
		return checkin.AuthorizationInfo{}, errors.New("Error parsing token: " + err.Error())
	}
	claims, _ := token.Claims.(jwt.MapClaims)
	return checkin.AuthorizationInfo{
		Username: claims[jwtUsername].(string),
		IsAdmin:  claims[jwtAdminStatus].(bool)}, nil
}

//IssueAuthorization Writes a response using the given ResponseWriter containing authorization info
//For the recipient
func (jwta Authenticator) IssueAuthorization(au checkin.AuthorizationInfo, w http.ResponseWriter) error {
	jwt, err := jwta.createToken(au)
	if err != nil {
		return errors.New("Error creating token: " + err.Error())
	}
	reply, _ := json.Marshal(map[string]string{"accessToken": jwt})
	w.Write(reply)
	return nil
}

//createToken Given a user (or admin's) authorization info, returns an encrypted web token string
//Uses signing method HS256
func (jwta Authenticator) createToken(au checkin.AuthorizationInfo) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims[jwtUsername] = au.Username
	claims[jwtExpiryTime] = time.Now().Add(time.Hour).Unix()
	claims[jwtAdminStatus] = au.IsAdmin

	tokenSigned, err := token.SignedString(jwta.SigningKey)
	if err != nil {
		return "", errors.New("Token signing failed during creation: " + err.Error())
	}

	return tokenSigned, nil
}

//keyGetter checks if the provided token follows the appropriate signing method
//Returns an error if not
//Returns the signing key if it does follow the appropriate method
func (jwta Authenticator) keyGetter(token *jwt.Token) (interface{}, error) {
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
