package auth

import (
	"log"
	"net/http"

	"registration-app/response"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

type AccessRestriction func(jwt.MapClaims, *http.Request) (bool, error)

func AccessControl(canAccess AccessRestriction, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := jwt.Parse(GetJWTString(r), KeyGetter)

		if err != nil {
			log.Println("Failed to extract token: " + err.Error())
			response.WriteMessage(http.StatusBadRequest, "Could not decipher authorization token", w)
			return
		}
		if !token.Valid {
			response.WriteMessage(http.StatusUnauthorized, "Token has expired", w)
			return
		}

		claims, _ := token.Claims.(jwt.MapClaims)
		r.Header.Set(JWTUsername, claims[JWTUsername].(string))
		if ok, err := canAccess(claims, r); err != nil {
			log.Println("Error in access control: " + err.Error())
			response.WriteMessage(http.StatusInternalServerError, "Error in checking credentials", w)
		} else if ok {
			h.ServeHTTP(w, r)
		} else {
			response.WriteMessage(http.StatusForbidden, "Access Denied", w)
		}
	})
}

func NoRestriction(claims jwt.MapClaims, r *http.Request) (bool, error) {
	return true, nil
}

func IsAdmin(claims jwt.MapClaims, r *http.Request) (bool, error) {
	return (claims[JWTAdminStatus] == true), nil
}

func SpecificUserOrAdmin(claims jwt.MapClaims, r *http.Request) (bool, error) {
	return (claims[JWTAdminStatus] == true || claims[JWTUsername] == mux.Vars(r)["username"]), nil
}
