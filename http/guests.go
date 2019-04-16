package http

import (
	"bytes"
	"checkin"
	"encoding/csv"
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
	GuestService   checkin.GuestService
	EventService   checkin.EventService
	GuestMessenger GuestMessenger
	Logger         *log.Logger
	Authenticator  Authenticator
}

//NewGuestHandler creates a new GuestHandler, using the default logger, with the
//pre-defined routing
func NewGuestHandler(gs checkin.GuestService, es checkin.EventService, gm GuestMessenger,
	auth Authenticator) *GuestHandler {
	h := &GuestHandler{
		Router:         mux.NewRouter(),
		Logger:         log.New(os.Stderr, "", log.LstdFlags),
		GuestService:   gs,
		EventService:   es,
		GuestMessenger: gm,
		Authenticator:  auth,
	}

	//Adapters to check if handler should serve the request
	tokenCheck := checkAuth(auth)
	credentialsCheck := isAdminOrHost(auth, es, "eventID")
	existCheck := eventExists(es, "eventID")
	releaseCheck := eventReleased(es, "eventID")

	h.Handle("/api/v0/events/{eventID}/guests", Adapt(http.HandlerFunc(h.handleGuests),
		tokenCheck, existCheck, credentialsCheck)).Methods("GET")
	h.Handle("/api/v0/events/{eventID}/guests", Adapt(http.HandlerFunc(h.handleRegisterGuest),
		tokenCheck, existCheck, credentialsCheck)).Methods("POST")
	h.Handle("/api/v0/events/{eventID}/guests", Adapt(http.HandlerFunc(h.handleRemoveGuest),
		tokenCheck, existCheck, credentialsCheck)).Methods("DELETE")
	h.Handle("/api/v0/events/{eventID}/guests/checkedin", Adapt(http.HandlerFunc(h.handleGuestsCheckedIn),
		tokenCheck, existCheck, credentialsCheck)).Methods("GET")
	h.Handle("/api/v0/events/{eventID}/guests/checkedin", Adapt(http.HandlerFunc(h.handleCheckInGuest),
		existCheck, releaseCheck)).Methods("POST")
	h.Handle("/api/v0/events/{eventID}/guests/checkedin", Adapt(http.HandlerFunc(h.handleMarkGuestAbsent),
		tokenCheck, existCheck, credentialsCheck)).Methods("DELETE")
	h.Handle("/api/v1-2/events/{eventID}/guests/checkedin/listener",
		Adapt(http.HandlerFunc(h.handleCreateCheckInListener), existCheck)).Methods("POST")
	h.Handle("/api/v1-2/events/{eventID}/guests/checkedin/listener",
		Adapt(http.HandlerFunc(h.handleDeleteCheckInListener), existCheck)).Methods("DELETE")
	h.Handle("/api/v0/events/{eventID}/guests/notcheckedin", Adapt(http.HandlerFunc(h.handleGuestsNotCheckedIn),
		tokenCheck, existCheck, credentialsCheck)).Methods("GET")
	h.Handle("/api/v0/events/{eventID}/guests/stats", Adapt(http.HandlerFunc(h.handleStats),
		tokenCheck, existCheck, credentialsCheck)).Methods("GET")
	h.Handle("/api/v0/events/{eventID}/guests/report", Adapt(http.HandlerFunc(h.handleReport),
		tokenCheck, existCheck, credentialsCheck)).Methods("GET")

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
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(&guest)
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
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(&guest)
	if err != nil {
		h.Logger.Println("Error when decoding guest NRIC: " + err.Error())
		WriteMessage(http.StatusBadRequest, "Incorrect fields for removing guest (need only NRIC)", w)
		return
	}
	if guest.Name != "" {
		WriteMessage(http.StatusBadRequest, "Incorrect fields for removing guest (need only NRIC)", w)
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

func (h *GuestHandler) handleMarkGuestAbsent(w http.ResponseWriter, r *http.Request) {
	var guest checkin.Guest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(&guest)
	if err != nil {
		h.Logger.Println("Error when decoding guest details: " + err.Error())
		WriteMessage(http.StatusBadRequest, "Incorrect fields for marking guest absent", w)
		return
	}
	if guest.Name != "" {
		WriteMessage(http.StatusBadRequest, "Incorrect fields for removing guest (need only NRIC)", w)
		return
	}

	eventID := mux.Vars(r)["eventID"]
	//check if the guest exists before attempting to mark it as absent
	if guestExists, err := h.GuestService.GuestExists(eventID, guest.NRIC); err == nil && !guestExists {
		WriteMessage(http.StatusNotFound, "No such guest to mark absent", w)
		return
	} else if err != nil {
		h.Logger.Println("Error checking if guest exists: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error checking if guest exists", w)
		return
	}

	err = h.GuestService.MarkAbsent(eventID, guest.NRIC)
	if err != nil {
		h.Logger.Println("Error check guest in: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Guest check-in failed", w)
	}

	//if anyone subscribed to a check in listener on this guest, update them
	if h.GuestMessenger.HasConnection(generateGuestID(eventID, guest.NRIC)) {
		err = h.GuestMessenger.Send(generateGuestID(eventID, guest.NRIC), GuestMessage{
			Title: "checkedin/0",
		})
		if err != nil {
			h.Logger.Println("Error sending check in update to guest, but guest successfully checked in: " +
				guest.NRIC + ", due to error: " + err.Error())
			//do not stop execution, this is a non-fatal bug
		}
	}

	WriteOKMessage("Successfully marked guest as absent", w)
}

func (h *GuestHandler) handleCheckInGuest(w http.ResponseWriter, r *http.Request) {
	var guest checkin.Guest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(&guest)
	if err != nil {
		h.Logger.Println("Error when decoding guest details: " + err.Error())
		WriteMessage(http.StatusBadRequest, "Incorrect fields for checking in guest", w)
		return
	}
	if guest.Name != "" {
		WriteMessage(http.StatusBadRequest, "Incorrect fields for removing guest (need only NRIC)", w)
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
		return
	}

	//if anyone subscribed to a check in listener on this guest, update them
	if h.GuestMessenger.HasConnection(generateGuestID(eventID, guest.NRIC)) {
		err = h.GuestMessenger.Send(generateGuestID(eventID, guest.NRIC), GuestMessage{
			Title:   "checkedin/1",
			Content: checkin.Guest{Name: name, NRIC: guest.NRIC},
		})
		if err != nil {
			h.Logger.Println("Error sending check in message to guest, but guest successfully checked in: " +
				guest.NRIC + ", due to error: " + err.Error())
		}
	}

	reply, _ := json.Marshal(name)
	w.Write(reply)
}

//Generates the guest ID to be used for the GuestMessenger
//using NRIC alone would be insufficient as one guest could go to multiple events
func generateGuestID(eventID string, guestNRIC string) string {
	return eventID + " " + guestNRIC
}

func (h *GuestHandler) handleCreateCheckInListener(w http.ResponseWriter, r *http.Request) {
	var guest checkin.Guest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(&guest)
	if err != nil {
		h.Logger.Println("Error when decoding guest details: " + err.Error())
		WriteMessage(http.StatusBadRequest,
			"Incorrect fields for listening on check in - must have {\"nric\":\"something\"}", w)
		return
	}

	eventID := mux.Vars(r)["eventID"]
	guestID := generateGuestID(eventID, guest.NRIC)
	err = h.GuestMessenger.OpenConnection(guestID, w, r)
	if err != nil {
		h.Logger.Println("Error when attempting to open connection with " + guest.NRIC + ": " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error starting listener on check in", w)
		return
	}

	WriteOKMessage("You will now be updated when guest "+guest.NRIC+" is checked in", w)
}

func (h *GuestHandler) handleDeleteCheckInListener(w http.ResponseWriter, r *http.Request) {
	var guest checkin.Guest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(&guest)
	if err != nil {
		h.Logger.Println("Error when decoding guest details: " + err.Error())
		WriteMessage(http.StatusBadRequest,
			"Incorrect fields for stopping check in listening - must have {\"nric\":\"something\"}", w)
		return
	}

	eventID := mux.Vars(r)["eventID"]
	guestID := generateGuestID(eventID, guest.NRIC)
	err = h.GuestMessenger.CloseConnection(guestID)
	if err != nil {
		h.Logger.Println("Error when attempting to close connection with " + guest.NRIC + ": " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error closing listener on check in", w)
		return
	}

	WriteOKMessage("You will no longer be updated when guest "+guest.NRIC+" is checked in", w)
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

func (h *GuestHandler) handleReport(w http.ResponseWriter, r *http.Request) {
	eventID := mux.Vars(r)["eventID"]
	absent, err := h.GuestService.GuestsNotCheckedIn(eventID)
	if err != nil {
		h.Logger.Println("Error in handleReport when getting absent guests: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error fetching absentees", w)
		return
	}
	present, err := h.GuestService.GuestsCheckedIn(eventID)
	if err != nil {
		h.Logger.Println("Error in handleReport when getting present guests: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error fetching those present", w)
		return
	}

	b := &bytes.Buffer{}
	wr := csv.NewWriter(b)
	wr.Write([]string{"Name", "Present"})
	for _, guestName := range present {
		wr.Write([]string{guestName, "1"})
	}
	for _, guestName := range absent {
		wr.Write([]string{guestName, "0"})
	}
	wr.Flush()

	w.Header().Set("Content-Type", "text/csv")
	//set the file name here
	w.Header().Set("Content-Disposition", "attachment;filename=AttendanceReport.csv")
	w.Write(b.Bytes())
}
