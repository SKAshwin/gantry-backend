package http

import (
	"checkin"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

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
	QRGenerator   checkin.QRGenerator
}

//NewGuestHandler creates a new GuestHandler, using the default logger, with the
//pre-defined routing
func NewGuestHandler(gs checkin.GuestService, es checkin.EventService, auth Authenticator, qrg checkin.QRGenerator) *GuestHandler {
	h := &GuestHandler{
		Router:        mux.NewRouter(),
		Logger:        log.New(os.Stderr, "", log.LstdFlags),
		GuestService:  gs,
		EventService:  es,
		Authenticator: auth,
		QRGenerator:   qrg,
	}

	//Adapters to check if handler should serve the request
	tokenCheck := checkAuth(auth)
	credentialsCheck := isAdminOrHost(auth, es, "eventID")
	existCheck := eventExists(es, "eventID")

	h.Handle("/api/v0/events/{eventID}/guests", Adapt(http.HandlerFunc(h.handleGuests),
		tokenCheck, credentialsCheck, existCheck)).Methods("GET")
	h.Handle("/api/v0/events/{eventID}/guests", Adapt(http.HandlerFunc(h.handleRegisterGuest),
		tokenCheck, credentialsCheck, existCheck)).Methods("POST")
	h.Handle("/api/v0/events/{eventID}/guests", Adapt(http.HandlerFunc(h.handleRemoveGuest),
		tokenCheck, credentialsCheck, existCheck)).Methods("DELETE")
	h.Handle("/api/v0/events/{eventID}/guests/checkedin", Adapt(http.HandlerFunc(h.handleGuestsCheckedIn),
		tokenCheck, credentialsCheck, existCheck)).Methods("GET")
	h.Handle("/api/v0/events/{eventID}/guests/checkedin", Adapt(http.HandlerFunc(h.handleCheckInGuest),
		existCheck)).Methods("POST")
	h.Handle("/api/v0/events/{eventID}/guests/notcheckedin", Adapt(http.HandlerFunc(h.handleGuestsNotCheckedIn),
		tokenCheck, credentialsCheck, existCheck)).Methods("GET")
	h.Handle("/api/v0/events/{eventID}/guests/stats", Adapt(http.HandlerFunc(h.handleStats),
		tokenCheck, credentialsCheck, existCheck)).Methods("GET")
	h.Handle("/api/v0/events/{eventID}/guests/qrcode", Adapt(http.HandlerFunc(h.handleQRGeneration),
		existCheck)).Methods("POST")

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
	//GET /api/events/{eventID}/guests/stats should return the summary statistics, requires host or
	//admin

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

func (h *GuestHandler) handleRegisterGuest(w http.ResponseWriter, r *http.Request) {
	var guest checkin.Guest
	err := json.NewDecoder(r.Body).Decode(&guest)
	if err != nil {
		h.Logger.Println("Error when decoding guest details: " + err.Error())
		WriteMessage(http.StatusBadRequest, "Incorrect fields for adding new guest", w)
		return
	}

	eventID := mux.Vars(r)["eventID"]

	//check if the guest already exists first before attempting to create one
	if guestExists, err := h.GuestService.GuestExists(eventID, guest.NRIC); err == nil && guestExists {
		WriteMessage(http.StatusConflict, "Guest with that NRIC already in list", w)
		return
	} else if err != nil {
		h.Logger.Println("Error checking if guest exists: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error checking if guest exists", w)
		return
	}

	err = h.GuestService.RegisterGuest(eventID, guest.NRIC, guest.Name)
	if err != nil {
		h.Logger.Println("Error registering guest: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Guest registration failed", w)
	} else {
		WriteMessage(http.StatusCreated, "Registration successful", w)
	}
}

func (h *GuestHandler) handleRemoveGuest(w http.ResponseWriter, r *http.Request) {
	eventID := mux.Vars(r)["eventID"]
	var guest checkin.Guest
	err := json.NewDecoder(r.Body).Decode(&guest)
	if err != nil {
		h.Logger.Println("Error when decoding guest NRIC: " + err.Error())
		WriteMessage(http.StatusBadRequest, "Incorrect fields for removing guest (need NRIC as string)", w)
		return
	}

	err = h.GuestService.RemoveGuest(eventID, guest.NRIC)
	if err != nil {
		h.Logger.Println("Error deleting guest: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error deleting guest", w)
	} else {
		WriteOKMessage("Successfully deleted guest", w)
	}
}

func (h *GuestHandler) handleGuestsCheckedIn(w http.ResponseWriter, r *http.Request) {
	eventID := mux.Vars(r)["eventID"]
	guests, err := h.GuestService.GuestsCheckedIn(eventID)
	if err != nil {
		h.Logger.Println("Error in handleGuestsCheckedIn: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error fetching checked-in guests for event", w)
		return
	}
	reply, _ := json.Marshal(guests)
	w.Write(reply)
}

func (h *GuestHandler) handleCheckInGuest(w http.ResponseWriter, r *http.Request) {
	var guest checkin.Guest
	err := json.NewDecoder(r.Body).Decode(&guest)
	if err != nil {
		h.Logger.Println("Error when decoding guest details: " + err.Error())
		WriteMessage(http.StatusBadRequest, "Incorrect fields for checking in guest", w)
		return
	}

	eventID := mux.Vars(r)["eventID"]
	//check if the guest exists before attempting to check it in
	if guestExists, err := h.GuestService.GuestExists(eventID, guest.NRIC); err == nil && !guestExists {
		WriteMessage(http.StatusNotFound, "No such guest to check in", w)
		return
	} else if err != nil {
		h.Logger.Println("Error checking if guest exists: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error checking if guest exists", w)
		return
	}

	name, err := h.GuestService.CheckIn(eventID, guest.NRIC)
	if err != nil {
		h.Logger.Println("Error check guest in: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Guest check-in failed", w)
	}

	reply, _ := json.Marshal(name)
	w.Write(reply)
}

func (h *GuestHandler) handleGuestsNotCheckedIn(w http.ResponseWriter, r *http.Request) {
	eventID := mux.Vars(r)["eventID"]
	guests, err := h.GuestService.GuestsNotCheckedIn(eventID)
	if err != nil {
		h.Logger.Println("Error in handleNotGuestsCheckedIn: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error fetching not checked-in guests for event", w)
		return
	}
	reply, _ := json.Marshal(guests)
	w.Write(reply)
}

func (h *GuestHandler) handleStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.GuestService.CheckInStats(mux.Vars(r)["eventID"])
	if err != nil {
		h.Logger.Println("Error in handleStats: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error fetching statistics for event", w)
		return
	}
	reply, _ := json.Marshal(stats)
	w.Write(reply)
}

func (h *GuestHandler) handleQRGeneration(w http.ResponseWriter, r *http.Request) {
	var guest checkin.Guest
	err := json.NewDecoder(r.Body).Decode(&guest)
	if err != nil {
		h.Logger.Println("Error when decoding guest NRIC for QRGeneration: " + err.Error())
		WriteMessage(http.StatusBadRequest, "Incorrect fields for generating QRCode (need NRIC as string)", w)
		return
	}

	img, err := h.QRGenerator.Encode(guest.NRIC, 20)
	if err != nil {
		h.Logger.Println("Error when generating QR Code: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error generating QR Code", w)
		return
	}

	w.Header().Set("Content-Type", http.DetectContentType(img))
	w.Header().Set("Content-Length", strconv.Itoa(len(img)))
	w.Write(img)
}
