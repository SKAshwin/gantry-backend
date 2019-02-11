package postgres

import (
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" //Loads postgres as our database of choice
)

//Open Returns a connection to the database, given the relevant credentials
//Will return an error if the database culd not be opened, or if the database
//could not be pinged after opening
func Open(host string, port string, user string, password string, name string) (*sqlx.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable", host, port, user, password, name)

	DB, err := sqlx.Open("postgres", psqlInfo)
	if err != nil {
		return nil, errors.New("Could not open database: " + err.Error())
	}
	err = DB.Ping()
	if err != nil {
		return nil, errors.New("Could not ping database " + err.Error())
	}
	return DB, nil
}
