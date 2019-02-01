package users

import (
	"errors"
	"log"
	"time"

	"github.com/guregu/null"
	"github.com/jmoiron/sqlx"

	"checkin/auth"
	"checkin/config"
)

//User represents a user of the website creator - i.e., hosts of events, who want to run an event
//check-in page.
//This is distinguished from guests, who attend events, and
type User struct {
	Username          string    `json:"username,omitempty" db:"username"`
	PasswordPlaintext string    `json:"password,omitempty"`
	PasswordHash      string    `json:"-" db:"passwordHash"` //always omitted upon JSON marshalling
	Name              string    `json:"name,omitempty" db:"name"`
	CreatedAt         time.Time `json:"createdAt,omitempty"`
	UpdatedAt         time.Time `json:"updatedAt,omitempty"`
	LastLoggedIn      null.Time `json:"lastLoggedIn,omitempty"`
}

const (
	dbUsername = "username"
	dbPassword = "passwordHash"
	dbName     = "name"
)

var (
	errUserDoesNotExist    = errors.New("User does not exist")
	updateSchemaTranslator = map[string]string{"username": dbUsername, "password": dbPassword, "name": dbName}
)

//CreateUser Given a User object with a plaintext password, username and name,
//enters a new user into the database
func (userData User) CreateUser() error {
	passwordHash, err := auth.HashAndSalt([]byte(userData.PasswordPlaintext))
	if err != nil {
		return errors.New("createUser: " + err.Error())
	}
	_, err = config.DB.Exec("INSERT into app_user (username,passwordHash,name,createdAt,updatedAt,lastLoggedIn) VALUES ($1, $2, $3, NOW(), NOW(), NULL)",
		userData.Username, passwordHash, userData.Name)
	return err
}

func Delete(username string) error {
	_, err := config.DB.Exec("DELETE from app_user where username = $1", username)
	return err
}

//Update updates a particular user given their username, and a map of attributes to new values
//Returns a boolean flag indicating if the arguments were valid
//Returns a non-nil error if there was an error updating the user
func Update(username string, updateFields map[string]string) (bool, error) {
	//check if the update fields are valid
	//this sanitizes the input for later
	if !IsUpdateRequestValid(updateFields) {
		return false, nil
	}

	tx, err := config.DB.Begin()
	if err != nil {
		return false, errors.New("Error opening transaction:" + err.Error())
	}

	defer func() {
		if r := recover(); r != nil {
			log.Println("users.Update entered panic, recovered to rollback, with error: ", r)
			if rollBackErr := tx.Rollback(); rollBackErr != nil {
				log.Println("Could not rollback: " + rollBackErr.Error())
			}
			panic("user.Update panicked")
		}
	}()

	for attribute, newValue := range updateFields {
		if attribute == "password" { //password needs to be hashed for update
			newValue, err = auth.HashAndSalt([]byte(newValue))
			if err != nil {
				return false, errors.New("Could not hash new password: " + err.Error())
			}
		}
		_, err := tx.Exec("UPDATE app_user SET "+updateSchemaTranslator[attribute]+" = $1 where username = $2", newValue, username)
		if err != nil {
			tx.Rollback()
			return false, errors.New("Error while updating database: " + err.Error())
		}
		if attribute == "username" { //if primary key, username, was changed
			username = newValue //need to know for all future changes
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

//IsUpdateRequestValid checks if the fields provided in an update request
//are allowed. Only specific columns are allowed to be updated
func IsUpdateRequestValid(updateFields map[string]string) bool {
	for attribute := range updateFields {
		if _, exist := updateSchemaTranslator[attribute]; !exist {
			return false
		}
	}
	return true

}

//GetData returns a User object with the username, name, creation/update/lastloggedin time stamps
//for a user with the given username
func GetData(username string) (User, error) {
	var userDetail User
	err := config.DB.QueryRowx("SELECT username, name, createdAt, updatedAt, lastLoggedIn from app_user where username = $1", username).StructScan(&userDetail)

	if err != nil {
		return User{}, errors.New("Could not fetch user details: " + err.Error())
	}

	return userDetail, nil
}

func CheckIfExists(username string) (bool, error) {
	var numUsers int
	err := config.DB.QueryRow("SELECT COUNT(*) from app_user where username = $1", username).Scan(&numUsers)
	if err != nil {
		return false, err
	}
	return numUsers != 0, nil
}

func GetAll() ([]User, error) {
	rows, err := config.DB.Queryx("SELECT username, name, createdAt, updatedAt, lastLoggedIn from app_user")
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

//UpdateLastLoggedIn Sets the lastLoggedIn attribute of this user to the current time
func UpdateLastLoggedIn(username string) error {
	_, err := config.DB.Exec("UPDATE app_user SET lastLoggedIn = NOW() where username = $1", username)
	return err
}

func getNumberOfUsers() (int, error) {
	var numUsers int
	err := config.DB.QueryRow("SELECT count(*) from app_user").Scan(&numUsers)

	if err != nil {
		return 0, errors.New("Cannot fetch user count: " + err.Error())
	}

	return numUsers, nil
}

func scanRowsIntoUserDetails(rows *sqlx.Rows, rowCount int) ([]User, error) {
	users := make([]User, rowCount)

	index := 0
	for thereAreMore := rows.Next(); thereAreMore; thereAreMore = rows.Next() {
		var userDetail User
		err := rows.StructScan(&userDetail)
		if err != nil {
			return nil, errors.New("Could not extract user details: " + err.Error())
		}
		users[index] = userDetail
		index++
	}

	return users, nil
}
