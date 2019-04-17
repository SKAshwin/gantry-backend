package http

import (
	"net/http"
	"strings"
)

//Handler the overall handler for the HTTP server, redirects requests towards its
//sub-handlers
//Implements the http.Handler interface
type Handler struct {
	EventHandler   *EventHandler
	UserHandler    *UserHandler
	AuthHandler    *AuthHandler
	UtilityHandler *UtilityHandler
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	urlSections := strings.Split(r.URL.Path, "/")
	if len(urlSections) < 4 {
		http.NotFound(w, r)
		return
	}
	// URL in the form of /api/{versionNumber}/{handlerType}/....
	// e.g. /api/v1-2/auth/failedattempts/{username}
	//Use {handlerType} to distinguish what handler to use
	if urlSections[3] == "events" {
		h.EventHandler.ServeHTTP(w, r)
	} else if urlSections[3] == "users" {
		h.UserHandler.ServeHTTP(w, r)
	} else if urlSections[3] == "auth" {
		h.AuthHandler.ServeHTTP(w, r)
	} else if urlSections[3] == "utility" {
		h.UtilityHandler.ServeHTTP(w, r)
	} else {
		http.NotFound(w, r)
	}
}
