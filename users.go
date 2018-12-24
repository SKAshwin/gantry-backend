package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/guregu/null"
)

type userPublicDetail struct {
	Username     string
	Name         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastLoggedIn null.Time
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

	var numUsers int
	err = db.QueryRow("SELECT count(*) from app_user").Scan(&numUsers)

	if err != nil {
		return nil, errors.New("Cannot fetch user count: " + err.Error())
	}

	users := make([]userPublicDetail, numUsers)

	index := 0
	for thereAreMore := rows.Next(); thereAreMore; thereAreMore = rows.Next() {
		var username, name string
		var createdAt, updatedAt time.Time
		var lastLoggedIn null.Time
		err = rows.Scan(&username, &name, &createdAt, &updatedAt, &lastLoggedIn)
		if err != nil {
			return nil, errors.New("Could not extract user details: " + err.Error())
		}
		users[index] = userPublicDetail{username, name, createdAt, updatedAt, lastLoggedIn}
		index++
	}
	return users, nil
}
