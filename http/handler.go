package http

import (
	"net/http"
	"strings"
)

//Handler the overall handler for the HTTP server, redirects requests towards its
//sub-handlers
//Implements the http.Handler interface
type Handler struct {
	EventHandler *EventHandler
	UserHandler  *UserHandler
	AuthHandler  *AuthHandler
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/v0/events") {
		h.EventHandler.ServeHTTP(w, r)
	} else if strings.HasPrefix(r.URL.Path, "/api/v0/users") {
		h.UserHandler.ServeHTTP(w, r)
	} else if strings.HasPrefix(r.URL.Path, "/api/v0/auth") {
		h.AuthHandler.ServeHTTP(w, r)
	} else {
		http.NotFound(w, r)
	}
}
