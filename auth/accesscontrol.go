package auth

import (
	"log"
	"net/http"

	"registration-app/response"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

//AccessOption is a function which, given a http request and the java web token claims attached
//Returns true or false depending on whether this request is allowed
//Returns a non-nil error if there was an error checking credentials (an error will block access)
//AccessControl checks for 1. the existence of a web token and 2. it's validity by default
//AccessOptions are for requirements beyond this
type AccessOption func(jwt.MapClaims, *http.Request) (bool, error)

//AccessControl is a middleware which checks authorization
//It firstly checks that there is a valid java web token attached to the request.
//If not, it does not execute the handler
//AccessOptions can also be specified. If specified, at least one of them must return true for access to be allowed
//and for the handler to be executed
//If not specified, all requests with valid web tokens will be allowed, and the handler will be executed
//AccessOptions are evaluated in the sequence provided/with shortcircuiting
//If an error occurs in any of the AccessOptions after one of them returns true, the error will not stop execution
//If an error occurs before any AccessOption returns true, error will be logged and handler will not execute
func AccessControl(h http.Handler, requirements ...AccessOption) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString, err := GetJWTString(r)
		if err != nil {
			log.Println("Failed to extract token string: " + err.Error())
			response.WriteMessage(http.StatusBadRequest, "Authorization token in invalid format", w)
			return
		}
		token, err := jwt.Parse(tokenString, KeyGetter)

		if err != nil {
			log.Println("Failed to extract token: " + err.Error())
			response.WriteMessage(http.StatusBadRequest, "Could not decipher authorization token", w)
			return
		}
		if !token.Valid { //always check for token validity
			response.WriteMessage(http.StatusUnauthorized, "Token has expired", w)
			return
		}

		claims, _ := token.Claims.(jwt.MapClaims)
		r.Header.Set(JWTUsername, claims[JWTUsername].(string)) //add the username to the request, for use by the handlers

		if len(requirements) == 0 { //if there are no further requirements, serve the request to the handler
			h.ServeHTTP(w, r)
			return
		}

		for _, req := range requirements {
			if ok, err := req(claims, r); err != nil {
				log.Println("Error in access control: " + err.Error())
				response.WriteMessage(http.StatusInternalServerError, "Error in checking credentials", w)
				return
			} else if ok {
				//if at least one requirement is met, allow access, serve up the request to the handler
				h.ServeHTTP(w, r)
				return
			}
		}
		//if none of the requirements are met, deny access
		response.WriteMessage(http.StatusForbidden, "Access Denied", w)

	})
}

//IsAdmin returns true if the web token's admin status is true
//Returns false otherwise
//Error is always nil
func IsAdmin(claims jwt.MapClaims, r *http.Request) (bool, error) {
	return (claims[JWTAdminStatus] == true), nil
}

//IsSpecifiedUser returns true if the username in the endpoint corresponds
//to the username in the web token
//Returns false otherwise
//Error is always nil
func IsSpecifiedUser(claims jwt.MapClaims, r *http.Request) (bool, error) {
	return claims[JWTUsername] == mux.Vars(r)["username"], nil
}
