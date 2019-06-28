package http

import (
	"checkin"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

//GuestSiteHandler An extension of mux.Router which handles all website-related requests for events
//Uses the given EventService, the given Logger, and a given Authenticator to check if
//requests are valid
//Uses the GuestSiteServices to make changes to the permanent storage
//Call NewGuestSiteHandler to initialize an GuestSiteHandler with the correct routes
type GuestSiteHandler struct {
	*mux.Router
	GuestSiteService checkin.GuestSiteService
	EventService     checkin.EventService
	Logger           *log.Logger
	Authenticator    Authenticator
}

//NewGuestSiteHandler Creates a new guest site handler using gorilla/mux for routing
//and the default Logger.
//GuestHandler, EventService, Authenticator needs to be set by the calling function
//API endpoint changes happen here, as well as changes to the routing library and logger to be used
//and type of authenticator
func NewGuestSiteHandler(gss checkin.GuestSiteService, es checkin.EventService, auth Authenticator) *GuestSiteHandler {
	h := &GuestSiteHandler{
		Router:           mux.NewRouter(),
		Logger:           log.New(os.Stderr, "", log.LstdFlags),
		Authenticator:    auth,
		GuestSiteService: gss,
		EventService:     es,
	}
	//Adapters to check if handler should serve the request
	tokenCheck := checkAuth(auth, h.Logger)
	credentialsCheck := isAdminOrHost(auth, es, "eventID", h.Logger)
	existCheck := eventExists(es, "eventID", h.Logger)
	urlConversion := eventURLToID(es, "eventID", h.Logger)
	h.Handle("/api/v1-4/events/{eventID}/website", Adapt(http.NotFoundHandler(),
		urlConversion, existCheck)).Methods("GET") //GET the website JSON
	h.Handle("/api/v1-4/events/{eventID}/website", Adapt(http.NotFoundHandler(),
		urlConversion, existCheck, tokenCheck, credentialsCheck)).Methods("POST") //create a new website for that event
	h.Handle("/api/v1-4/events/{eventID}/website", Adapt(http.NotFoundHandler(),
		urlConversion, existCheck, tokenCheck, credentialsCheck)).Methods("PATCH") //update the website's JSON
	h.Handle("/api/v1-4/events/{eventID}/website", Adapt(http.NotFoundHandler(),
		urlConversion, existCheck, tokenCheck, credentialsCheck)).Methods("DELETE") //remove the website's JSON

	return h
}

func (h *GuestSiteHandler) handleWebsite(w http.ResponseWriter, r *http.Request) {

}
