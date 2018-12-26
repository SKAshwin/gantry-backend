package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/guregu/null"
)

type userPublicDetail struct {
	Username     string    `json:"username"`
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	LastLoggedIn null.Time `json:"lastLoggedIn"`
}

type userPublicDetails []userPublicDetail

var listUsersHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//writeMessage("Hey you made it here", w)
	userDetails, err := fetchUserDetails()
	if err != nil {
		log.Println(err.Error())
		writeError(http.StatusInternalServerError, "Could not get user data", w)
		return
	}
	reply, _ := json.Marshal(map[string]userPublicDetails{"message": userDetails})
	w.Write(reply)
})

func fetchUserDetails() ([]userPublicDetail, error) {
	rows, err := db.Query("SELECT username, name, createdAt, updatedAt, lastLoggedIn from app_user")
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

func scanRowsIntoUserDetails(rows *sql.Rows, rowCount int) ([]userPublicDetail, error) {
	users := make([]userPublicDetail, rowCount)

	index := 0
	for thereAreMore := rows.Next(); thereAreMore; thereAreMore = rows.Next() {
		var username, name string
		var createdAt, updatedAt time.Time
		var lastLoggedIn null.Time
		err := rows.Scan(&username, &name, &createdAt, &updatedAt, &lastLoggedIn)
		if err != nil {
			return nil, errors.New("Could not extract user details: " + err.Error())
		}
		users[index] = userPublicDetail{username, name, createdAt, updatedAt, lastLoggedIn}
		index++
	}

	return users, nil
}
