package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

const (
	dbhost = "DBHOST"
	dbport = "DBPORT"
	dbuser = "DBUSER"
	dbpass = "DBPASS"
	dbname = "DBNAME"
)

//CORSConfig Contains the configuration information for CORS
type CORSConfig struct {
	AllowedOrigins []string `json:"allowedOrigins"`
	AllowedMethods []string `json:"allowedMethods"`
	AllowedHeaders []string `json:"allowedHeaders"`
}

//DB the connection to the database used by the entire program
var DB *sqlx.DB

//GetCorsConfig returns the CORSConfig information in cors.json
func GetCorsConfig() CORSConfig {
	corsFile, err := os.Open("config/cors.json")
	if err != nil {
		log.Fatal("Error opening cors.json: " + err.Error())
	}
	defer corsFile.Close()
	byteValue, err := ioutil.ReadAll(corsFile)
	if err != nil {
		log.Fatal("Error reading cors.json: " + err.Error())
	}
	var config CORSConfig
	err = json.Unmarshal([]byte(byteValue), &config)
	if err != nil {
		log.Fatal("cors.json formatted wrongly, error when parsing: " + err.Error())
	}
	return config
}

//LoadEnvironmentalVariables Loads variables from the .env
func LoadEnvironmentalVariables() {
	err := godotenv.Load()
	if err != nil {
		log.Print("Error loading environmental variables: ")
		log.Fatal(err.Error())
	}
}

//RedirectLogger redirects logging output to a file instead of the console
//for use in production
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

//InitDB initializes the Db based on the environmental variables
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
