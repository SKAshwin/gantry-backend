package config

import (
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
)

const (
	dbhost = "DBHOST"
	dbport = "DBPORT"
	dbuser = "DBUSER"
	dbpass = "DBPASS"
	dbname = "DBNAME"
)

var DB *sqlx.DB

func RedirectLogger() {
	//redirects logger output to a logger file
	//for use in production
	file, err := os.OpenFile("info.log", os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()
	log.SetOutput(file)
}

func InitDB() {
	//initializes the db variable
	//forms a connection to the database
	config := dbConfig()
	var err error
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		config[dbhost], config[dbport],
		config[dbuser], config[dbpass], config[dbname])

	DB, err = sqlx.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("Could not open databse")
	}
	err = DB.Ping()
	if err != nil {
		log.Fatal("Could not ping database")
	}
	log.Println("Successfully connected!")
}

func dbConfig() map[string]string {
	//reads from environmental variables to work out details of the database
	conf := make(map[string]string)
	host, ok := os.LookupEnv(dbhost)
	if !ok {
		log.Fatal("DBHOST environment variable required but not set")
	}
	port, ok := os.LookupEnv(dbport)
	if !ok {
		log.Fatal("DBPORT environment variable required but not set")
	}
	user, ok := os.LookupEnv(dbuser)
	if !ok {
		log.Fatal("DBUSER environment variable required but not set")
	}
	password, ok := os.LookupEnv(dbpass)
	if !ok {
		log.Fatal("DBPASS environment variable required but not set")
	}
	name, ok := os.LookupEnv(dbname)
	if !ok {
		panic("DBNAME environment variable required but not set")
	}
	conf[dbhost] = host
	conf[dbport] = port
	conf[dbuser] = user
	conf[dbpass] = password
	conf[dbname] = name
	return conf
}
