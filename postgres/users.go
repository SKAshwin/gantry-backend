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

//User Fetches the details of the user with that username
func (us *UserService) User(username string) (checkin.User, error) {
	var u checkin.User
	err := us.DB.QueryRowx(
		"SELECT username, name, passwordHash, createdAt, updatedAt, lastLoggedIn from app_user where username = $1",
		username).StructScan(&u)

	if err != nil {
		return checkin.User{}, errors.New("Could not fetch user details: " + err.Error())
	}

	return u, nil
}

//Users Returns the details of all users
func (us *UserService) Users() ([]checkin.User, error) {
	rows, err := us.DB.Queryx("SELECT username, name, passwordHash, createdAt, updatedAt, lastLoggedIn from app_user")
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
	passwordHash, err := us.HM.HashAndSalt(*u.PasswordPlaintext)
	if err != nil {
		return errors.New("createUser: " + err.Error())
	}
	_, err = us.DB.Exec("INSERT into app_user (username,passwordHash,name,createdAt,updatedAt,lastLoggedIn) VALUES ($1, $2, $3, (NOW() at time zone 'utc'), (NOW() at time zone 'utc'), NULL)",
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
func (us *UserService) UpdateUser(originalUsername string, user checkin.User) error {
	tx, err := us.DB.Beginx()
	if err != nil {
		return errors.New("Error opening transaction:" + err.Error())
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

	if user.PasswordPlaintext != nil { //if a new password plain text is set
		user.PasswordHash, err = us.HM.HashAndSalt(*user.PasswordPlaintext)
		if err != nil {
			return errors.New("Error hashing new password: " + err.Error())
		}
	}

	_, err = tx.Exec("UPDATE app_user SET username = $1, passwordHash = $2, "+
		"name = $3 WHERE username = $4", user.Username, user.PasswordHash, user.Name, originalUsername)
	if err != nil {
		tx.Rollback()
		return errors.New("Error while updating database: " + err.Error())
	}

	_, err = tx.Exec("UPDATE app_user SET updatedAt = (NOW() at time zone 'utc') where username = $1", user.Username)
	if err != nil {
		tx.Rollback()
		return errors.New("Error when updating updated field in app_user: " + err.Error())
	}

	err = tx.Commit()
	if err != nil {
		return errors.New("Error committing changes to database: " + err.Error())
	}

	return nil
}

//CheckIfExists sees if the username is already used
func (us *UserService) CheckIfExists(username string) (bool, error) {
	var numUsers int
	err := us.DB.QueryRow("SELECT COUNT(*) from app_user where username = $1", username).Scan(&numUsers)
	if err != nil {
		return false, err
	}
	return numUsers != 0, nil
}

//UpdateLastLoggedIn Sets the lastLoggedIn attribute of this user to the current time
func (us *UserService) UpdateLastLoggedIn(username string) error {
	_, err := us.DB.Exec("UPDATE app_user SET lastLoggedIn = (NOW() at time zone 'utc') where username = $1", username)
	return err
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
