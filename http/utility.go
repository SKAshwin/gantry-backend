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

//UtilityHandler An extension of mux.Router which handles some utility requests
//Utility requests are requests which do not need any of the data services
//and do not require authentication
type UtilityHandler struct {
	*mux.Router
	Logger      *log.Logger
	QRGenerator checkin.QRGenerator
}

//NewUtilityHandler creates a new UtilityHandler, using the default logger
//with the pre-defined routes registered
func NewUtilityHandler(qrg checkin.QRGenerator) *UtilityHandler {
	h := &UtilityHandler{
		Router:      mux.NewRouter(),
		Logger:      log.New(os.Stderr, "", log.LstdFlags),
		QRGenerator: qrg,
	}
	h.Handle("/api/v0/utility/qrcode", http.HandlerFunc(h.handleQRGeneration)).Methods("POST")
	return h
}

func (h *UtilityHandler) handleQRGeneration(w http.ResponseWriter, r *http.Request) {
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
