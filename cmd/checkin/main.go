package main

import (
	"checkin/bcrypt"
	"checkin/http"
	"checkin/http/jwt"
	"checkin/postgres"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/joho/godotenv"
)

func main() {
	config := dbConfig()
	db, err := postgres.Open(config["DBHOST"], config["DBPORT"], config["DBUSER"],
		config["DBPASS"], config["DBNAME"])
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	jwtAuthenticator := jwt.Authenticator{SigningKey: []byte("MyPassword")}
	bcryptHashMethod := bcrypt.HashMethod{HashCost: 5}

	us := &postgres.UserService{DB: db, HM: bcryptHashMethod}
	as := &postgres.AuthenticationService{DB: db, HM: bcryptHashMethod}

	authHandler := http.NewAuthHandler(as, jwtAuthenticator, us)

	handler := http.Handler{
		AuthHandler: authHandler,
	}
	server := http.Server{Handler: &handler, Addr: ":5000"}
	server.Open() //note that server.Open starts a new goroutine, so process will end
	//unless blocked

	// Block until an interrupt signal is received, to keep the server alive
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	s := <-c
	fmt.Println("Got signal:", s)
}

func loadEnvironmentalVariables() {
	err := godotenv.Load()
	if err != nil {
		log.Print("Error loading environmental variables: ")
		log.Fatal(err.Error())
	}
}

func dbConfig() map[string]string {
	//reads from environmental variables to work out details of the database
	conf := make(map[string]string)
	host, ok := os.LookupEnv("DBHOST")
	if !ok {
		log.Fatal("DBHOST environment variable required but not set")
	}
	port, ok := os.LookupEnv("DBPORT")
	if !ok {
		log.Fatal("DBPORT environment variable required but not set")
	}
	user, ok := os.LookupEnv("DBUSER")
	if !ok {
		log.Fatal("DBUSER environment variable required but not set")
	}
	password, ok := os.LookupEnv("DBPASS")
	if !ok {
		log.Fatal("DBPASS environment variable required but not set")
	}
	name, ok := os.LookupEnv("DBNAME")
	if !ok {
		log.Fatal("DBNAME environment variable required but not set")
	}
	conf["DBHOST"] = host
	conf["DBPORT"] = port
	conf["DBUSER"] = user
	conf["DBPASS"] = password
	conf["DBNAME"] = name
	return conf
}
