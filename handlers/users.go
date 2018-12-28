package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"registration-app/response"
	"registration-app/users"

	"github.com/gorilla/mux"
)

var ListUsers = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//writeMessage("Hey you made it here", w)
	userDetails, err := users.GetAll()
	if err != nil {
		log.Println(err.Error())
		response.WriteMessage(http.StatusInternalServerError, "Could not get user data", w)
		return
	}
	type userPublicDetails []users.UserPublicDetail
	reply, _ := json.Marshal(map[string]userPublicDetails{"message": userDetails})
	w.Write(reply)
})

var CreateUser = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var userData users.UserCreateData
	err := json.NewDecoder(r.Body).Decode(&userData)
	if err != nil {
		log.Println("Error when decoding name: " + err.Error())
		response.WriteMessage(http.StatusBadRequest, "Incorrect fields for creating user", w)
		return
	}

	//check if the user already exists first before attempting to create one
	if userExists, err := users.CheckIfExists(userData.Username); err == nil && userExists {
		response.WriteMessage(http.StatusConflict, "Username already taken", w)
		return
	} else if err != nil {
		log.Println("Error checking if user exists: " + err.Error())
		response.WriteMessage(http.StatusInternalServerError, "Error checking if user exists", w)
		return
	}

	err = userData.CreateUser()
	if err != nil {
		log.Println("Error creating user: " + err.Error())
		response.WriteMessage(http.StatusInternalServerError, "User creation failed", w)
	} else {
		response.WriteMessage(http.StatusCreated, "Registration successful", w)
	}
})

var GetUser = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]
	userData, err := users.GetData(username)
	if err != nil {
		log.Println("Error fetching user data: " + err.Error())
		response.WriteMessage(http.StatusInternalServerError, "Could not get user data", w)
	} else {
		reply, _ := json.Marshal(userData)
		w.Write(reply)
	}
})

var UpdateUserDetails = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var updatedFields map[string]string
	err := json.NewDecoder(r.Body).Decode(&updatedFields)
	if err != nil {
		log.Println("Error when decoding update fields: " + err.Error())
		response.WriteMessage(http.StatusBadRequest, "JSON could not be decoded", w)
		return
	}

	username := mux.Vars(r)["username"] //middleware already confirms user exists
	validRequest, err := users.Update(username, updatedFields)

	if err != nil {
		log.Println("Error updating user: " + err.Error())
		response.WriteMessage(http.StatusInternalServerError, "Error updating user", w)
	} else if !validRequest {
		response.WriteMessage(http.StatusBadRequest, "Incorrect fields for user update", w)
	} else {
		response.WriteOKMessage("User updated", w)
	}

})

var DeleteUser = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"] //user already confirmed to exist through middleware
	err := users.Delete(username)

	if err != nil {
		log.Println("Error deleting user: " + err.Error())
		response.WriteMessage(http.StatusInternalServerError, "Error deleting user", w)
	} else {
		response.WriteOKMessage("Successfully deleted user", w)
	}
})
