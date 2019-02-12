package http

import (
	"checkin"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

//GuestHandler is a sub-handler of the EventHandler, which handles all requests pertaining to
//Guests, using a GuestService and EventService for data access,
// and an Authenticator to grant access
type GuestHandler struct {
	*mux.Router
	GuestService  checkin.GuestService
	EventService  checkin.EventService
	Logger        *log.Logger
	Authenticator Authenticator
}

//NewGuestHandler creates a new GuestHandler, using the default logger, with the
//pre-defined routing
func NewGuestHandler(gs checkin.GuestService, es checkin.EventService, auth Authenticator) *GuestHandler {
	h := &GuestHandler{
		Router:        mux.NewRouter(),
		Logger:        log.New(os.Stderr, "", log.LstdFlags),
		GuestService:  gs,
		EventService:  es,
		Authenticator: auth,
	}

	//Adapters to check if handler should serve the request
	tokenCheck := checkAuth(auth)
	credentialsCheck := isAdminOrHost(auth, es, "eventID")
	existCheck := eventExists(es, "eventID")

	h.Handle("/api/v0/events/{eventID}/guests", Adapt(http.HandlerFunc(h.handleGuests),
		tokenCheck, credentialsCheck, existCheck))

	//GET /api/events/{eventID}/guests should return all Guests, requires a host token or admin token
	//POST /api/events/{eventID}/guests with a JSON argument {name:"Hello",nric:"5678F"} should register
	//a new guest, requires host token or admin token
	//DELETE /api/events/{eventID}/guests with a JSON argument {nric:"5678F"} should remove a guest
	//from the registered list, requires host or admin
	//GET /api/events/{eventID}/guests/checkedin should return Guests who have checked in, requires host
	//or admin
	//POST /api/events/{eventID}/guests/checkedin with JSON argument {nric:"5678F"} should check in a
	//guest that is already registered, no permissions (except CORS)
	//GET /api/events/{eventID}/guests/notcheckedin should return Guests who haven't checked in,
	//requires host or admin
	//GET /api/events/{eventID}/guests/stats should return the summary statistics

	return h
}

func (h *GuestHandler) handleGuests(w http.ResponseWriter, r *http.Request) {
	guests, err := h.GuestService.Guests(mux.Vars(r)["eventID"])
	if err != nil {
		h.Logger.Println("Error in handleGuests: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error fetching all guests for event", w)
		return
	}
	reply, _ := json.Marshal(guests)
	w.Write(reply)
}
