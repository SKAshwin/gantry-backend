package postgres

import (
	"checkin"
	"errors"
	"log"

	"github.com/jmoiron/sqlx"
)

//UserService Implementation of a user service
//Needs to be supplied with a database connection as well as a hashing method
type UserService struct {
	DB *sqlx.DB
	HM checkin.HashMethod
}

//Maps json fields to DB fields
var userUpdateSchema = map[string]string{
	"username": "username",
	"password": "passwordHash",
	"name":     "name"}

//User Fetches the details of the user with that username
func (us *UserService) User(username string) (checkin.User, error) {
	var u checkin.User
	err := us.DB.QueryRowx(
		"SELECT username, name, createdAt, updatedAt, lastLoggedIn from app_user where username = $1",
		username).StructScan(&u)

	if err != nil {
		return checkin.User{}, errors.New("Could not fetch user details: " + err.Error())
	}

	return u, nil
}

//Users Returns the details of all users
func (us *UserService) Users() ([]checkin.User, error) {
	rows, err := us.DB.Queryx("SELECT username, name, createdAt, updatedAt, lastLoggedIn from app_user")
	if err != nil {
		return nil, errors.New("Cannot fetch user details: " + err.Error())
	}
	defer rows.Close() //make sure this is after checking for an error, or this will be a nil pointer dereference
	numUsers, err := us.getNumberOfUsers()
	if err != nil {
		return nil, errors.New("Cannot fetch number of users: " + err.Error())
	}

	return us.scanRowsIntoUserDetails(rows, numUsers)
}

//CreateUser Adds a user with the given username, password (will hash it) and name to the records
func (us *UserService) CreateUser(u checkin.User) error {
	passwordHash, err := us.HM.HashAndSalt(u.PasswordPlaintext)
	if err != nil {
		return errors.New("createUser: " + err.Error())
	}
	_, err = us.DB.Exec("INSERT into app_user (username,passwordHash,name,createdAt,updatedAt,lastLoggedIn) VALUES ($1, $2, $3, NOW(), NOW(), NULL)",
		u.Username, passwordHash, u.Name)
	return err
}

//DeleteUser Deletes the records of the user with the given username
func (us *UserService) DeleteUser(username string) error {
	_, err := us.DB.Exec("DELETE from app_user where username = $1", username)
	return err
}

//UpdateUser updates a particular user given their username, and a map of attributes to new values
//Returns a boolean flag indicating if the arguments were valid
//Returns a non-nil error if there was an error updating the user
func (us *UserService) UpdateUser(username string, updateFields map[string]string) (bool, error) {
	//check if the update fields are valid
	//this sanitizes the input for later
	if !isUserUpdateRequestValid(updateFields) {
		return false, nil
	}

	tx, err := us.DB.Begin()
	if err != nil {
		return false, errors.New("Error opening transaction:" + err.Error())
	}

	defer func() {
		if r := recover(); r != nil {
			log.Println("UpdateUser entered panic, recovered to rollback, with error: ", r)
			if rollBackErr := tx.Rollback(); rollBackErr != nil {
				log.Println("Could not rollback: " + rollBackErr.Error())
			}
			panic("UpdateUser panicked")
		}
	}()

	for attribute, newValue := range updateFields {
		if attribute == "password" { //password needs to be hashed for update
			newValue, err = us.HM.HashAndSalt(newValue)
			if err != nil {
				return false, errors.New("Could not hash new password: " + err.Error())
			}
		}
		_, err := tx.Exec("UPDATE app_user SET "+userUpdateSchema[attribute]+" = $1 where username = $2", newValue, username)
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

//CheckIfExists sees if the username is already used
func (us UserService) CheckIfExists(username string) (bool, error) {
	var numUsers int
	err := us.DB.QueryRow("SELECT COUNT(*) from app_user where username = $1", username).Scan(&numUsers)
	if err != nil {
		return false, err
	}
	return numUsers != 0, nil
}

//UpdateLastLoggedIn Sets the lastLoggedIn attribute of this user to the current time
func (us UserService) UpdateLastLoggedIn(username string) error {
	_, err := us.DB.Exec("UPDATE app_user SET lastLoggedIn = NOW() where username = $1", username)
	return err
}

func isUserUpdateRequestValid(updateFields map[string]string) bool {
	for attribute := range updateFields {
		if _, exist := userUpdateSchema[attribute]; !exist {
			return false
		}
	}
	return true
}

func (us *UserService) getNumberOfUsers() (int, error) {
	var i int
	err := us.DB.QueryRow("SELECT count(*) from app_user").Scan(&i)

	if err != nil {
		return 0, errors.New("Cannot fetch user count: " + err.Error())
	}

	return i, nil
}

func (us *UserService) scanRowsIntoUserDetails(rows *sqlx.Rows, rowCount int) ([]checkin.User, error) {
	users := make([]checkin.User, rowCount)

	index := 0
	for ok := rows.Next(); ok; ok = rows.Next() {
		var u checkin.User
		err := rows.StructScan(&u)
		if err != nil {
			return nil, errors.New("Could not extract user details: " + err.Error())
		}
		users[index] = u
		index++
	}

	return users, nil
}
