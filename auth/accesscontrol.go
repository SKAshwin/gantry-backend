package auth

import (
	"log"
	"net/http"

	"registration-app/response"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

type AccessRestriction func(jwt.MapClaims, *http.Request) bool

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
		if canAccess(claims, r) {
			h.ServeHTTP(w, r)
		} else {
			response.WriteMessage(http.StatusUnauthorized, "Access Unauthorized", w)
		}
	})
}

func NoRestriction(claims jwt.MapClaims, r *http.Request) bool {
	return true
}

func IsAdmin(claims jwt.MapClaims, r *http.Request) bool {
	return claims[jwtAdminStatus] == true
}

func SpecificUserOrAdmin(claims jwt.MapClaims, r *http.Request) bool {
	return claims[jwtAdminStatus] == true || claims[jwtUsername] == mux.Vars(r)["username"]
}
