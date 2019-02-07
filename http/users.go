package http

import (
	"checkin"
	"log"

	"github.com/gorilla/mux"
)

//UserHandler An extension of mux.Router which handles all user-related requests
//Uses the given UserService and the given Logger
//Call NewUserHandler to initiate a UserHandler with the correct routes
type UserHandler struct {
	*mux.Router
	UserService checkin.UserService
	Logger      *log.Logger
}
