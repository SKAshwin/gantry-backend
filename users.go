package main

import (
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

const (
	dbUsername = "username"
	dbPassword = "passwordHash"
	dbName     = "name"
)

var (
	errUserDoesNotExist    = errors.New("User does not exist")
	updateSchemaTranslator = map[string]string{"username": dbUsername, "password": dbPassword, "name": dbName}
)

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
		return
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
		log.Println("Error fetching user data: " + err.Error())
		writeMessage(http.StatusInternalServerError, "Could not get user data", w)
	} else {
		reply, _ := json.Marshal(userData)
		w.Write(reply)
	}
})

var updateUserDetailsHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var updatedFields map[string]string
	err := json.NewDecoder(r.Body).Decode(&updatedFields)
	if err != nil {
		log.Println("Error when decoding update fields: " + err.Error())
		writeMessage(http.StatusBadRequest, "JSON could not be decoded", w)
		return
	}

	username := mux.Vars(r)["username"] //middleware already confirms user exists
	validRequest, err := updateUser(username, updatedFields)

	if err != nil {
		log.Println("Error updating user: " + err.Error())
		writeMessage(http.StatusInternalServerError, "Error updating user", w)
	} else if !validRequest {
		writeMessage(http.StatusBadRequest, "Incorrect fields for user update", w)
	} else {
		writeOKMessage("User updated", w)
	}

})

var deleteUserHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"] //user already confirmed to exist through middleware
	err := deleteUser(username)

	if err != nil {
		log.Println("Error deleting user: " + err.Error())
		writeMessage(http.StatusInternalServerError, "Error deleting user", w)
	} else {
		writeOKMessage("Successfully deleted user", w)
	}
})

func userExists(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := mux.Vars(r)["username"]
		if exists, err := checkIfUserExists(username); err != nil {
			log.Println("Error checking if user exists" + err.Error())
			writeMessage(http.StatusInternalServerError, "Error checking if user exists", w)
		} else if !exists {
			writeMessage(http.StatusNotFound, "User does not exist", w)
		} else {
			h.ServeHTTP(w, r)
		}
	})
}

func createUser(userData userCreateData) error {
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

func updateUser(username string, updateFields map[string]string) (bool, error) {
	//check if the update fields are valid
	//this sanitizes the input for later
	if !isUpdateRequestValid(updateFields) {
		return false, nil
	}

	tx, err := db.Begin()
	if err != nil {
		return false, errors.New("Error opening transaction:" + err.Error())
	}

	defer func() {
		if r := recover(); r != nil {
			log.Println("updateUser entered panic, recovered to rollback, with error: ", r)
			if rollBackErr := tx.Rollback(); rollBackErr != nil {
				log.Println("Could not rollback: " + err.Error())
			}
		}
	}()

	for attribute, newValue := range updateFields {
		//primary key literally can't be updated in this syntax anyway
		_, err := tx.Exec("UPDATE app_user SET "+updateSchemaTranslator[attribute]+" = $1 where username = $2", newValue, username)
		if err != nil {
			tx.Rollback()
			return false, errors.New("Error while updating database: " + err.Error())
		}
	}

	_, err = tx.Exec("UPDATE app_user SET updatedAt = NOW() where username = $1", username)
	if err != nil {
		tx.Rollback()
		return false, errors.New("Error when updating updated field in app_user: " + err.Error())
	}

	err = tx.Commit()
	if err != nil {
		return false, errors.New("Error committing changes to database: " + err.Error())
	}

	return true, nil
}

func isUpdateRequestValid(updateFields map[string]string) bool {
	for attribute := range updateFields {
		if _, exist := updateSchemaTranslator[attribute]; !exist {
			return false
		}
	}
	return true

}

func getUserData(username string) (userPublicDetail, error) {
	var userDetail userPublicDetail
	err := db.QueryRowx("SELECT username, name, createdAt, updatedAt, lastLoggedIn from app_user where username = $1", username).StructScan(&userDetail)

	if err != nil {
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
