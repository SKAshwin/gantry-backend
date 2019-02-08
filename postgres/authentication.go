package postgres

import (
	"checkin"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

//AuthenticationService is a postgres implementation of the AuthenticationService interface
//Needs to be supplied with the database connection and a HashMethod
type AuthenticationService struct {
	DB *sqlx.DB
	HM checkin.HashMethod
}

//Authenticate compares the given plaintext password against a database hash
//for that user's password (using the HashMethod)
//If the user does not exist or the password does not match with the hash, returns false
//If there is a database querying error, returns an error
func (as *AuthenticationService) Authenticate(username string, pwdPlaintext string, isAdmin bool) (bool, error) {
	var stmt *sqlx.Stmt
	var err error
	if isAdmin {
		stmt, err = as.DB.Preparex("SELECT passwordHash FROM app_admin where username = $1")
	} else {
		stmt, err = as.DB.Preparex("SELECT passwordHash FROM app_user where username = $1")
	}
	if err != nil {
		return false, errors.New("Statement preparation in authentication failed: " + err.Error())
	}
	var passwordHash string
	err = stmt.QueryRow(u.Username).Scan(&passwordHash)
	if err == sql.ErrNoRows {
		return false, nil //no such username exists
	} else if err != nil {
		//any other error represents a failure
		return false, errors.New("Database Querying in Authentication Failed: " + err.Error())
	}
	return as.HM.CompareHashAndPassword(passwordHash, u.PasswordPlaintext), nil
}
