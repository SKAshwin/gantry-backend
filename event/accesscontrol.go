package event

import (
	"net/http"
	"registration-app/auth"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

//HostOrAdmin An AccessRestriction that checks that the eventID specified
//in the endpoint called is of an event that is hosted by the user who's token
//is attached with the call
//Alterantively, the token can have admin status to bypass this restriction
func HostOrAdmin(claims jwt.MapClaims, r *http.Request) (bool, error) {
	username := r.Header.Get(auth.JWTUsername)
	eventID := mux.Vars(r)["eventID"]
	ok, err := CheckHost(username, eventID)
	if err != nil {
		return ok, err
	} else {
		return ok || claims[auth.JWTAdminStatus] == true, nil
	}
}
