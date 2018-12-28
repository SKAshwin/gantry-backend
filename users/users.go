package users

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/guregu/null"
	"github.com/jmoiron/sqlx"

	"registration-app/auth"
	"registration-app/config"
	"registration-app/response"
)

type UserPublicDetail struct {
	Username     string    `json:"username"`
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	LastLoggedIn null.Time `json:"lastLoggedIn"`
}

type UserCreateData struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
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

func UserExists(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := mux.Vars(r)["username"]
		if exists, err := CheckIfExists(username); err != nil {
			log.Println("Error checking if user exists" + err.Error())
			response.WriteMessage(http.StatusInternalServerError, "Error checking if user exists", w)
		} else if !exists {
			response.WriteMessage(http.StatusNotFound, "User does not exist", w)
		} else {
			h.ServeHTTP(w, r)
		}
	})
}

func (userData UserCreateData) CreateUser() error {
	passwordHash, err := auth.HashAndSalt([]byte(userData.Password))
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

func IsUpdateRequestValid(updateFields map[string]string) bool {
	for attribute := range updateFields {
		if _, exist := updateSchemaTranslator[attribute]; !exist {
			return false
		}
	}
	return true

}

func GetData(username string) (UserPublicDetail, error) {
	var userDetail UserPublicDetail
	err := config.DB.QueryRowx("SELECT username, name, createdAt, updatedAt, lastLoggedIn from app_user where username = $1", username).StructScan(&userDetail)

	if err != nil {
		return UserPublicDetail{}, errors.New("Could not fetch user details: " + err.Error())
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

func GetAll() ([]UserPublicDetail, error) {
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

func scanRowsIntoUserDetails(rows *sqlx.Rows, rowCount int) ([]UserPublicDetail, error) {
	users := make([]UserPublicDetail, rowCount)

	index := 0
	for thereAreMore := rows.Next(); thereAreMore; thereAreMore = rows.Next() {
		var userDetail UserPublicDetail
		err := rows.StructScan(&userDetail)
		if err != nil {
			return nil, errors.New("Could not extract user details: " + err.Error())
		}
		users[index] = userDetail
		index++
	}

	return users, nil
}
