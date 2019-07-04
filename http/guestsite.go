package http

import (
	"checkin"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"compress/gzip"

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
	eventExistCheck := eventExists(es, "eventID", h.Logger)
	urlConversion := eventURLToID(es, "eventID", h.Logger)
	siteExistCheck := websiteExists(gss, "eventID", h.Logger)
	h.Handle("/api/v1-4/events/{eventID}/website", Adapt(http.HandlerFunc(h.handleWebsite),
		urlConversion, eventExistCheck, siteExistCheck)).Methods("GET") //GET the website JSON
	h.Handle("/api/v1-4/events/{eventID}/website", Adapt(http.HandlerFunc(h.handleCreateWebsite),
		urlConversion, eventExistCheck, tokenCheck, credentialsCheck)).Methods("POST") //create a new website for that event
	h.Handle("/api/v1-4/events/{eventID}/website", Adapt(http.HandlerFunc(h.handleUpdateWebsite),
		urlConversion, eventExistCheck, siteExistCheck, tokenCheck, credentialsCheck)).Methods("PATCH") //update the website's JSON
	h.Handle("/api/v1-4/events/{eventID}/website", Adapt(http.HandlerFunc(h.handleDeleteWebsite),
		urlConversion, eventExistCheck, siteExistCheck, tokenCheck, credentialsCheck)).Methods("DELETE") //remove the website's JSON

	return h
}

//Returns the website associated with this event
func (h *GuestSiteHandler) handleWebsite(w http.ResponseWriter, r *http.Request) {
	website, err := h.GuestSiteService.GuestSite(mux.Vars(r)["eventID"])
	if err != nil {
		h.Logger.Println("Error fetching website: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error fetching website", w)
	} else {
		w.Header().Add("Accept-Charset", "utf-8")
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		json.NewEncoder(gz).Encode(website)
		gz.Close()
	}
}

func (h *GuestSiteHandler) handleCreateWebsite(w http.ResponseWriter, r *http.Request) {
	var website checkin.GuestSite
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(&website)
	if err != nil {
		h.Logger.Println("Error decoding website JSON: " + err.Error())
		WriteMessage(http.StatusBadRequest, "Badly formatted JSON for website (Possibly invalid question/component types or sizes)", w)
		return
	}

	eventID := mux.Vars(r)["eventID"]
	if exists, err := h.GuestSiteService.GuestSiteExists(eventID); err != nil {
		//check if the URL provided is available
		h.Logger.Println("Error checking if website already exists for event: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error checking if website already exists for event", w)
		return
	} else if exists {
		WriteMessage(http.StatusConflict, "Event already has a website - cannot create a new one. Use PATCH to update the existing one instead.", w)
		return
	}

	err = h.GuestSiteService.CreateGuestSite(eventID, website)
	if err != nil {
		h.Logger.Println("Error in creating website: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error in creating website", w)
	} else {
		WriteMessage(http.StatusCreated, "Website created successfully", w)
	}
}

func (h *GuestSiteHandler) handleUpdateWebsite(w http.ResponseWriter, r *http.Request) {
	//Load original website, marshal JSON into it
	//This updates only the fields that were supplied
	eventID := mux.Vars(r)["eventID"]
	website, err := h.GuestSiteService.GuestSite(eventID)
	if err != nil {
		h.Logger.Println("Error fetching original website: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Could not fetch original website", w)
		return
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err = dec.Decode(&website)
	if err != nil {
		h.Logger.Println("Error when decoding update fields for website: " + err.Error())
		WriteMessage(http.StatusBadRequest, "JSON could not be decoded into website", w)
		return
	}

	err = h.GuestSiteService.UpdateGuestSite(eventID, website)
	if err != nil {
		h.Logger.Println("Error updating user: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error updating event", w)
	} else {
		WriteOKMessage("Event updated", w)
	}
}

func (h *GuestSiteHandler) handleDeleteWebsite(w http.ResponseWriter, r *http.Request) {
	eventID := mux.Vars(r)["eventID"]
	err := h.GuestSiteService.DeleteGuestSite(eventID)
	if err != nil {
		h.Logger.Println("Error deleting website: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error deleting website", w)
	} else {
		WriteOKMessage("Successfully deleted website", w)
	}
}

func websiteExists(gss checkin.GuestSiteService, eventIDKey string, logger *log.Logger) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			eventID := mux.Vars(r)[eventIDKey]
			ok, err := gss.GuestSiteExists(eventID)
			if err != nil {
				logger.Println("Error checking that guest site exists: " + err.Error())
				WriteMessage(http.StatusInternalServerError, "Error checking if guest site exists", w)
			} else if ok {
				h.ServeHTTP(w, r)
			} else {
				WriteMessage(http.StatusNotFound, "Guest site does not exist for event with that ID", w)
			}
		})
	}
}
