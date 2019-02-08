package http

import (
	"checkin"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

//UserHandler An extension of mux.Router which handles all user-related requests
//Uses the given UserService and the given Logger
//Call NewUserHandler to initiate a UserHandler with the correct routes
type UserHandler struct {
	*mux.Router
	UserService   checkin.UserService
	Logger        *log.Logger
	Authenticator Authenticator
}

//NewUserHandler creates a new UserHandler that uses the given UserService and Authenticator
//And logs to the standard error output
func NewUserHandler(us checkin.UserService, auth Authenticator) *UserHandler {
	h := &UserHandler{
		Router:        mux.NewRouter(),
		Logger:        log.New(os.Stderr, "", log.LstdFlags),
		UserService:   us,
		Authenticator: auth,
	}

	//Adapters to check if handler should serve the request
	tokenCheck := checkAuth(auth)
	adminCheck := isAdmin(auth)
	existCheck := userExists(us, "username")
	adminOrUserCheck := isAdminOrUser(auth, us, "username")

	h.Handle("/api/users/v0", Adapt(http.HandlerFunc(h.handleUsers),
		tokenCheck, adminCheck)).Methods("GET")
	h.Handle("/api/users/v0", Adapt(http.HandlerFunc(h.handleCreateUser),
		tokenCheck, adminCheck)).Methods("POST")
	h.Handle("/api/users/v0/{username}", Adapt(http.HandlerFunc(h.handleUser),
		tokenCheck, adminOrUserCheck, existCheck)).Methods("GET")
	h.Handle("/api/users/v0/{username}", Adapt(http.HandlerFunc(h.handleUpdateUser),
		tokenCheck, adminOrUserCheck, existCheck)).Methods("PUT")
	h.Handle("/api/users/v0/{username}", Adapt(http.HandlerFunc(h.handleDeleteUser),
		tokenCheck, adminOrUserCheck, existCheck)).Methods("DELETE")

	return h
}

//handleUsers Sends JSON array of all users
func (h *UserHandler) handleUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.UserService.Users()
	if err != nil {
		h.Logger.Println("Error fetching all user data: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Could not get user data", w)
		return
	}
	reply, _ := json.Marshal(users)
	w.Write(reply)
}

//handleUser Sends one user's details back, based on the username in the URL
func (h *UserHandler) handleUser(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]
	user, err := h.UserService.User(username)
	if err != nil {
		h.Logger.Println("Error fetching user data: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Could not get user data", w)
		return
	}
	reply, _ := json.Marshal(user)
	w.Write(reply)
}

//handleCreateUser Creates a user given the necessary fields in JSON format in the
//body of the request
func (h *UserHandler) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var user checkin.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		h.Logger.Println("Error when decoding name: " + err.Error())
		WriteMessage(http.StatusBadRequest, "Incorrect fields for creating user", w)
		return
	}

	//check if the user already exists first before attempting to create one
	if userExists, err := h.UserService.CheckIfExists(user.Username); err == nil && userExists {
		WriteMessage(http.StatusConflict, "Username already taken", w)
		return
	} else if err != nil {
		h.Logger.Println("Error checking if user exists: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error checking if user exists", w)
		return
	}

	err = h.UserService.CreateUser(user)
	if err != nil {
		h.Logger.Println("Error creating user: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "User creation failed", w)
	} else {
		WriteMessage(http.StatusCreated, "Registration successful", w)
	}
}

//handleUpdateUser Reads the JSON as a map, only attributes to be updated need
//be supplied
func (h *UserHandler) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	var updatedFields map[string]string
	err := json.NewDecoder(r.Body).Decode(&updatedFields)
	if err != nil {
		h.Logger.Println("Error when decoding update fields: " + err.Error())
		WriteMessage(http.StatusBadRequest, "JSON could not be decoded", w)
		return
	}

	if val, ok := updatedFields["username"]; ok { //if the caller is attempting to update the username
		if ok, err := h.UserService.CheckIfExists(val); err != nil {
			h.Logger.Println("Error checking if username taken: " + err.Error())
			WriteMessage(http.StatusInternalServerError, "Error checking if username already taken", w)
			return
		} else if ok {
			WriteMessage(http.StatusConflict, "Username already exists", w)
			return
		}
	}

	username := mux.Vars(r)["username"] //middleware already confirms user exists
	validRequest, err := h.UserService.UpdateUser(username, updatedFields)

	if err != nil {
		h.Logger.Println("Error updating user: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error updating user", w)
	} else if !validRequest {
		WriteMessage(http.StatusBadRequest, "Incorrect fields for user update", w)
	} else {
		WriteOKMessage("User updated", w)
	}
}

func (h *UserHandler) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"] //user already confirmed to exist through middleware
	err := h.UserService.DeleteUser(username)

	if err != nil {
		h.Logger.Println("Error deleting user: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error deleting user", w)
	} else {
		WriteOKMessage("Successfully deleted user", w)
	}
}

//Adapter generator that checks if a user exists before allowing
//handler to serve request
func userExists(us checkin.UserService, usernameKey string) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username := mux.Vars(r)[usernameKey]
			if exists, err := us.CheckIfExists(username); err != nil {
				log.Println("Error checking if user exists" + err.Error())
				WriteMessage(http.StatusInternalServerError, "Error checking if user exists", w)
			} else if !exists {
				WriteMessage(http.StatusNotFound, "User does not exist", w)
			} else {
				h.ServeHTTP(w, r)
			}
		})
	}
}
