package http

import (
	"bytes"
	"checkin"
	"encoding/csv"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

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
	h.Handle("/api/v1-3/events/{eventID}/guests", Adapt(http.HandlerFunc(h.handleRegisterGuests),
		tokenCheck, existCheck, credentialsCheck)).Methods("POST")
	h.Handle("/api/v0/events/{eventID}/guests", Adapt(http.HandlerFunc(h.handleRemoveGuest),
		tokenCheck, existCheck, credentialsCheck)).Methods("DELETE")
	h.Handle("/api/v1-3/events/{eventID}/guests/tags", Adapt(http.HandlerFunc(h.handleTags),
		tokenCheck, existCheck, credentialsCheck)).Methods("GET")
	h.Handle("/api/v0/events/{eventID}/guests/checkedin", Adapt(http.HandlerFunc(h.handleGuestsCheckedIn),
		tokenCheck, existCheck, credentialsCheck)).Methods("GET")
	h.Handle("/api/v0/events/{eventID}/guests/checkedin", Adapt(http.HandlerFunc(h.handleCheckInGuest),
		existCheck, releaseCheck)).Methods("POST")
	h.Handle("/api/v0/events/{eventID}/guests/checkedin", Adapt(http.HandlerFunc(h.handleMarkGuestAbsent),
		tokenCheck, existCheck, credentialsCheck)).Methods("DELETE")
	h.Handle("/api/v1-2/events/{eventID}/guests/checkedin/listener/{nric}",
		Adapt(http.HandlerFunc(h.handleCreateCheckInListener), existCheck))
	h.Handle("/api/v0/events/{eventID}/guests/notcheckedin", Adapt(http.HandlerFunc(h.handleGuestsNotCheckedIn),
		tokenCheck, existCheck, credentialsCheck)).Methods("GET")
	h.Handle("/api/v0/events/{eventID}/guests/stats", Adapt(http.HandlerFunc(h.handleStats),
		tokenCheck, existCheck, credentialsCheck)).Methods("GET")
	h.Handle("/api/v0/events/{eventID}/guests/report", Adapt(http.HandlerFunc(h.handleReport),
		tokenCheck, existCheck, credentialsCheck)).Methods("GET")

	return h
}

func (h *GuestHandler) handleTags(w http.ResponseWriter, r *http.Request) {
	tags, err := h.GuestService.AllTags(mux.Vars(r)["eventID"])
	if err != nil {
		h.Logger.Println("Error in handleTags: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error fetching all tags for event", w)
		return
	}
	reply, _ := json.Marshal(tags)
	w.Write(reply)
}

func (h *GuestHandler) handleGuests(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		h.Logger.Println("Error parsing form queries: " + err.Error())
		WriteMessage(http.StatusBadRequest, "Could not parse query string", w)
		return
	}
	var guestsFunction func(string, []string) ([]string, error)
	if val, ok := r.Form["checkedin"]; !ok {
		//no checkedin=true or checkedin=false is set, so get all guests
		guestsFunction = h.GuestService.Guests
	} else if strings.ToLower(val[0]) == "true" {
		guestsFunction = h.GuestService.GuestsCheckedIn
	} else if strings.ToLower(val[0]) == "false" {
		guestsFunction = h.GuestService.GuestsNotCheckedIn
	} else {
		WriteMessage(http.StatusBadRequest, "Form value 'checkedin' must be either true or false (non-case sensitive)", w)
		return
	}

	guests, err := guestsFunction(mux.Vars(r)["eventID"], r.Form["tag"])
	if err != nil {
		h.Logger.Println("Error in handleGuests: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error fetching all guests for event", w)
		return
	}
	reply, _ := json.Marshal(guests)
	w.Write(reply)
}

func (h *GuestHandler) handleRegisterGuests(w http.ResponseWriter, r *http.Request) {
	var guests []checkin.Guest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(&guests)
	if err != nil {
		h.Logger.Println("Error when decoding guests's details: " + err.Error())
		WriteMessage(http.StatusBadRequest,
			`Incorrect fields for adding new guests. Must be an array of guest objects, i.e. in the form [{"nric":"1234A","name":"Hello","tags":["VIP","CONFIRMED"]}]. Even registering one guest should be one guest object in an array`,
			w)
		return
	}

	if guests == nil || len(guests) == 0 {
		WriteMessage(http.StatusBadRequest, `Cannot register an empty or null array of guests`, w)
		return
	}

	eventID := mux.Vars(r)["eventID"]
	for _, guest := range guests {
		//check if the guest already exists first before attempting to create one, for each guest
		if guestExists, err := h.GuestService.GuestExists(eventID, guest.NRIC); err == nil && guestExists {
			WriteMessage(http.StatusConflict, "Guest with that NRIC already in list for guest: "+guest.NRIC, w)
			return
		} else if err != nil {
			h.Logger.Println("Error checking if guest exists: " + err.Error())
			WriteMessage(http.StatusInternalServerError, "Error checking if guest exists, for guest: "+guest.NRIC, w)
			return
		}
	}

	err = h.GuestService.RegisterGuests(eventID, guests)
	if err != nil {
		h.Logger.Println("Error registering guests: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Guests registration failed; thus none of the guests supplied were registered", w)
	} else {
		WriteMessage(http.StatusCreated, "Registration successful for all guests", w)
	}
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

	err = h.GuestService.RegisterGuest(eventID, guest)
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
	guests, err := h.GuestService.GuestsCheckedIn(eventID, nil)
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
	nric := mux.Vars(r)["nric"]
	eventID := mux.Vars(r)["eventID"]
	guestID := generateGuestID(eventID, nric)

	err := h.GuestMessenger.OpenConnection(guestID, w, r)
	if err != nil {
		h.Logger.Println("Error when attempting to open guest messenger connection: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error starting listener on check in", w)
		return
	}
	//replying a 101 Protocol Changed is handled by the Open Connection method
}

func (h *GuestHandler) handleGuestsNotCheckedIn(w http.ResponseWriter, r *http.Request) {
	eventID := mux.Vars(r)["eventID"]
	guests, err := h.GuestService.GuestsNotCheckedIn(eventID, nil)
	if err != nil {
		h.Logger.Println("Error in handleNotGuestsCheckedIn: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error fetching not checked-in guests for event", w)
		return
	}
	reply, _ := json.Marshal(guests)
	w.Write(reply)
}

func (h *GuestHandler) handleStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.GuestService.CheckInStats(mux.Vars(r)["eventID"], nil)
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
	absent, err := h.GuestService.GuestsNotCheckedIn(eventID, nil)
	if err != nil {
		h.Logger.Println("Error in handleReport when getting absent guests: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error fetching absentees", w)
		return
	}
	present, err := h.GuestService.GuestsCheckedIn(eventID, nil)
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
