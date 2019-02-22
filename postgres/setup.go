package postgres

import (
	"errors"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" //Loads postgres as our database of choice
)

//Open Returns a connection to the database, given the relevant credentials
//Will return an error if the database culd not be opened, or if the database
//could not be pinged after opening
func Open(dbURL string) (*sqlx.DB, error) {
	DB, err := sqlx.Open("postgres", dbURL)
	if err != nil {
		return nil, errors.New("Could not open database: " + err.Error())
	}
	err = DB.Ping()
	if err != nil {
		return nil, errors.New("Could not ping database " + err.Error())
	}
	return DB, nil
}
