package event

import (
	"net/http"
	"registration-app/auth"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

//IsHost An AccessRestriction that checks that the eventID specified
//in the endpoint called is of an event that is hosted by the user who's token
//is attached with the call
func IsHost(claims jwt.MapClaims, r *http.Request) (bool, error) {
	username := r.Header.Get(auth.JWTUsername)
	eventID := mux.Vars(r)["eventID"]
	return CheckHost(username, eventID)
}
