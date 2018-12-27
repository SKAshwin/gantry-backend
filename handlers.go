package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var ListUsersHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//writeMessage("Hey you made it here", w)
	userDetails, err := getAllUsers()
	if err != nil {
		log.Println(err.Error())
		WriteMessage(http.StatusInternalServerError, "Could not get user data", w)
		return
	}
	reply, _ := json.Marshal(map[string]userPublicDetails{"message": userDetails})
	w.Write(reply)
})

var CreateUserHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var userData UserCreateData
	err := json.NewDecoder(r.Body).Decode(&userData)
	if err != nil {
		log.Println("Error when decoding name: " + err.Error())
		WriteMessage(http.StatusBadRequest, "Incorrect fields for creating user", w)
		return
	}

	//check if the user already exists first before attempting to create one
	if userExists, err := checkIfUserExists(userData.Username); err == nil && userExists {
		WriteMessage(http.StatusConflict, "Username already taken", w)
		return
	} else if err != nil {
		log.Println("Error checking if user exists: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error checking if user exists", w)
		return
	}

	err = createUser(userData)
	if err != nil {
		log.Println("Error creating user: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "User creation failed", w)
	} else {
		WriteMessage(http.StatusCreated, "Registration successful", w)
	}
})

var GetUserHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]
	userData, err := getUserData(username)
	if err != nil {
		log.Println("Error fetching user data: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Could not get user data", w)
	} else {
		reply, _ := json.Marshal(userData)
		w.Write(reply)
	}
})

var UpdateUserDetailsHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var updatedFields map[string]string
	err := json.NewDecoder(r.Body).Decode(&updatedFields)
	if err != nil {
		log.Println("Error when decoding update fields: " + err.Error())
		WriteMessage(http.StatusBadRequest, "JSON could not be decoded", w)
		return
	}

	username := mux.Vars(r)["username"] //middleware already confirms user exists
	validRequest, err := updateUser(username, updatedFields)

	if err != nil {
		log.Println("Error updating user: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error updating user", w)
	} else if !validRequest {
		WriteMessage(http.StatusBadRequest, "Incorrect fields for user update", w)
	} else {
		WriteOKMessage("User updated", w)
	}

})

var DeleteUserHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"] //user already confirmed to exist through middleware
	err := deleteUser(username)

	if err != nil {
		log.Println("Error deleting user: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Error deleting user", w)
	} else {
		WriteOKMessage("Successfully deleted user", w)
	}
})
