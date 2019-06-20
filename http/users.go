package http

import (
	"checkin"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

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
	tokenCheck := checkAuth(auth, h.Logger)
	adminCheck := isAdmin(auth, h.Logger)
	existCheck := userExists(us, "username", h.Logger)
	adminOrUserCheck := isAdminOrUser(auth, us, "username", h.Logger)

	h.Handle("/api/v0/users", Adapt(http.HandlerFunc(h.handleUsers),
		tokenCheck, adminCheck, correctTimezonesOutput, jsonSelector)).Methods("GET")
	h.Handle("/api/v0/users", Adapt(http.HandlerFunc(h.handleCreateUser),
		tokenCheck, adminCheck)).Methods("POST")
	h.Handle("/api/v0/users/{username}", Adapt(http.HandlerFunc(h.handleUser),
		tokenCheck, existCheck, adminOrUserCheck, correctTimezonesOutput, jsonSelector)).Methods("GET")
	h.Handle("/api/v0/users/{username}", Adapt(http.HandlerFunc(h.handleUpdateUser),
		tokenCheck, existCheck, adminOrUserCheck)).Methods("PATCH")
	h.Handle("/api/v0/users/{username}", Adapt(http.HandlerFunc(h.handleDeleteUser),
		tokenCheck, existCheck, adminOrUserCheck)).Methods("DELETE")

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
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(&user)
	if err != nil {
		h.Logger.Println("Error when decoding user creation data: " + err.Error())
		WriteMessage(http.StatusBadRequest, "Could not decode JSON; possibly invalid fields", w)
		return
	}

	if !validateCreateInputs(user) {
		WriteMessage(http.StatusBadRequest, "User creation data invalid. Cannot have blank username, name "+
			"or password; cannot set updatedAt or createdAt fields", w)
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

func validateCreateInputs(u checkin.User) bool {
	return u.Username != "" && u.PasswordPlaintext != nil && *u.PasswordPlaintext != "" && u.Name != "" &&
		u.CreatedAt == time.Time{} && u.UpdatedAt == time.Time{} && !u.LastLoggedIn.Valid
}

//handleUpdateUser Reads the JSON as a map, only attributes to be updated need
//be supplied
func (h *UserHandler) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	user, err := h.UserService.User(mux.Vars(r)["username"])
	if err != nil {
		h.Logger.Println("Error reading existing user data: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Could not fetch original user data", w)
		return
	}
	original := user
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err = dec.Decode(&user)
	if err != nil {
		h.Logger.Println("Error when decoding update fields: " + err.Error())
		WriteMessage(http.StatusBadRequest, "JSON could not be decoded, or invalid fields supplied", w)
		return
	}
	//validation
	if (user.PasswordPlaintext != nil && *user.PasswordPlaintext == "") || user.Name == "" ||
		user.Username == "" || user.CreatedAt != original.CreatedAt || user.UpdatedAt != original.UpdatedAt ||
		user.LastLoggedIn != original.LastLoggedIn {
		WriteMessage(http.StatusBadRequest, "Invalid fields supplied: cannot have empty password, name"+
			" or username, or change createdAt, updatedAt or lastLoggedIn fields", w)
		return
	}

	if original.Username != user.Username { //if the caller is attempting to update the username
		if ok, err := h.UserService.CheckIfExists(user.Username); err != nil {
			h.Logger.Println("Error checking if username taken: " + err.Error())
			WriteMessage(http.StatusInternalServerError, "Error checking if username already taken", w)
			return
		} else if ok {
			WriteMessage(http.StatusConflict, "Username already exists", w)
			return
		}
	}

	err = h.UserService.UpdateUser(original.Username, user)
	if err != nil {
		h.Logger.Println("Error updating user: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error updating user", w)
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
func userExists(us checkin.UserService, usernameKey string, logger *log.Logger) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username := mux.Vars(r)[usernameKey]
			if exists, err := us.CheckIfExists(username); err != nil {
				logger.Println("Error checking if user exists" + err.Error())
				WriteMessage(http.StatusInternalServerError, "Error checking if user exists", w)
			} else if !exists {
				WriteMessage(http.StatusNotFound, "User does not exist", w)
			} else {
				h.ServeHTTP(w, r)
			}
		})
	}
}
