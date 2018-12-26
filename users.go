package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/guregu/null"
	"github.com/jmoiron/sqlx"
)

type userPublicDetail struct {
	Username     string    `json:"username"`
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	LastLoggedIn null.Time `json:"lastLoggedIn"`
}

type userCreateData struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type userPublicDetails []userPublicDetail

var errUserDoesNotExist = errors.New("User does not exist")

var listUsersHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//writeMessage("Hey you made it here", w)
	userDetails, err := getAllUsers()
	if err != nil {
		log.Println(err.Error())
		writeMessage(http.StatusInternalServerError, "Could not get user data", w)
		return
	}
	reply, _ := json.Marshal(map[string]userPublicDetails{"message": userDetails})
	w.Write(reply)
})

var createUserHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var userData userCreateData
	err := json.NewDecoder(r.Body).Decode(&userData)
	if err != nil {
		log.Println("Error when decoding name: " + err.Error())
		writeMessage(http.StatusBadRequest, "Incorrect fields for creating user", w)
	}

	//check if the user already exists first before attempting to create one
	if userExists, err := checkIfUserExists(userData.Username); err == nil && userExists {
		writeMessage(http.StatusConflict, "Username already taken", w)
		return
	} else if err != nil {
		log.Println("Error checking if user exists: " + err.Error())
		writeMessage(http.StatusInternalServerError, "Error checking if user exists", w)
		return
	}

	err = createUser(userData)
	if err != nil {
		log.Println("Error creating user: " + err.Error())
		writeMessage(http.StatusInternalServerError, "User creation failed", w)
	} else {
		writeMessage(http.StatusCreated, "Registration successful", w)
	}
})

var getUserDetailsHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]
	userData, err := getUserData(username)
	if err != nil {
		if err == errUserDoesNotExist {
			log.Println("User not found")
			writeMessage(http.StatusNotFound, "User does not exist with that username", w)
		} else {
			log.Println("Error fetching user data: " + err.Error())
			writeMessage(http.StatusInternalServerError, "Could not get user data", w)
		}
	} else {
		reply, _ := json.Marshal(userData)
		w.Write(reply)
	}
})

var updateUserDetailsHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

})

var deleteUserHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]

	//confirm the user exists in the first place before deleting it
	if userExists, err := checkIfUserExists(username); err == nil && !userExists {
		writeMessage(http.StatusNotFound, "User does not exist", w)
		return
	} else if err != nil {
		log.Println("Error checking if user exists: " + err.Error())
		writeMessage(http.StatusInternalServerError, "Error checking if user exists", w)
		return
	}

	err := deleteUser(username)

	if err != nil {
		log.Println("Error deleting user: " + err.Error())
		writeMessage(http.StatusInternalServerError, "Error deleting user", w)
	} else {
		writeOKMessage("Successfully deleted user", w)
	}
})

func createUser(userData userCreateData) error {
	//TODO check if username already exists
	passwordHash, err := hashAndSalt([]byte(userData.Password))
	if err != nil {
		return errors.New("createUser: " + err.Error())
	}
	_, err = db.Exec("INSERT into app_user (username,passwordHash,name,createdAt,updatedAt,lastLoggedIn) VALUES ($1, $2, $3, NOW(), NOW(), NULL)",
		userData.Username, passwordHash, userData.Name)
	return err
}

func deleteUser(username string) error {
	_, err := db.Exec("DELETE from app_user where username = $1", username)
	return err
}

func getUserData(username string) (userPublicDetail, error) {
	var userDetail userPublicDetail
	err := db.QueryRowx("SELECT username, name, createdAt, updatedAt, lastLoggedIn from app_user where username = $1", username).StructScan(&userDetail)

	if err == sql.ErrNoRows {
		return userPublicDetail{}, errUserDoesNotExist
	} else if err != nil {
		return userPublicDetail{}, errors.New("Could not fetch user details: " + err.Error())
	}

	return userDetail, nil
}

func checkIfUserExists(username string) (bool, error) {
	var numUsers int
	err := db.QueryRow("SELECT COUNT(*) from app_user where username = $1", username).Scan(&numUsers)
	if err != nil {
		return false, err
	}
	return numUsers != 0, nil
}

func getAllUsers() ([]userPublicDetail, error) {
	rows, err := db.Queryx("SELECT username, name, createdAt, updatedAt, lastLoggedIn from app_user")
	if err != nil {
		return nil, errors.New("Cannot fetch user details: " + err.Error())
	}
	defer rows.Close() //make sure this is after checking for an error, or this will be a nil pointer dereference
	numUsers, err := getNumberOfUsers()
	if err != nil {
		return nil, errors.New("fetchUserDetails: " + err.Error())
	}

	return scanRowsIntoUserDetails(rows, numUsers)
}

func getNumberOfUsers() (int, error) {
	var numUsers int
	err := db.QueryRow("SELECT count(*) from app_user").Scan(&numUsers)

	if err != nil {
		return 0, errors.New("Cannot fetch user count: " + err.Error())
	}

	return numUsers, nil
}

func scanRowsIntoUserDetails(rows *sqlx.Rows, rowCount int) ([]userPublicDetail, error) {
	users := make([]userPublicDetail, rowCount)

	index := 0
	for thereAreMore := rows.Next(); thereAreMore; thereAreMore = rows.Next() {
		var userDetail userPublicDetail
		err := rows.StructScan(&userDetail)
		if err != nil {
			return nil, errors.New("Could not extract user details: " + err.Error())
		}
		users[index] = userDetail
		index++
	}

	return users, nil
}
