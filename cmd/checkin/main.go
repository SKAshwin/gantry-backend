package main

import (
	"checkin/bcrypt"
	"checkin/http"
	"checkin/http/cors"
	"checkin/http/jwt"
	"checkin/postgres"
	"checkin/qrcode"
	"checkin/sha512"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"

	"github.com/joho/godotenv"
)

func main() {
	loadEnvironmentalVariables() //comment this out for heroku production
	config := getConfig()
	db, err := postgres.Open(config["DATABASE_URL"])
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	log.Println("Successfully Connected!")

	hashCost, err := strconv.Atoi(config["HASH_COST"])
	if err != nil {
		log.Fatal("Error parsing HASH_COST env variable" + err.Error())
	}

	jwtAuthenticator := jwt.Authenticator{SigningKey: []byte(config["AUTH_SECRET"])}
	bcryptHashMethod := bcrypt.HashMethod{HashCost: hashCost}
	sha512HashMethod := sha512.HashMethod{}
	qrGenerator := qrcode.Generator{Level: qrcode.High}

	us := &postgres.UserService{DB: db, HM: bcryptHashMethod}
	as := &postgres.AuthenticationService{DB: db, HM: bcryptHashMethod}
	es := &postgres.EventService{DB: db}
	gs := &postgres.GuestService{DB: db, HM: sha512HashMethod}

	authHandler := http.NewAuthHandler(as, jwtAuthenticator, us)
	userHandler := http.NewUserHandler(us, jwtAuthenticator)
	guestHandler := http.NewGuestHandler(gs, es, jwtAuthenticator)
	eventHandler := http.NewEventHandler(es, jwtAuthenticator, guestHandler)
	utilityHandler := http.NewUtilityHandler(qrGenerator)

	handler := http.Handler{
		AuthHandler:    authHandler,
		EventHandler:   eventHandler,
		UserHandler:    userHandler,
		UtilityHandler: utilityHandler,
	}
	server := http.Server{Handler: &handler, Addr: ":" + config["PORT"], PreFlightHandler: configurePFH()}
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

func getConfig() map[string]string {
	//reads from environmental variables
	conf := make(map[string]string)
	dbURL, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		log.Fatal("DATABASE_URL environment variable required but not set")
	}
	authSecret, ok := os.LookupEnv("AUTH_SECRET")
	if !ok {
		log.Fatal("AUTH_SECRET environment variable required but not set")
	}
	hashCost, ok := os.LookupEnv("AUTH_HASH_COST")
	if !ok {
		log.Fatal("AUTH_HASH_COST environment variable required but not set")
	}
	port, ok := os.LookupEnv("PORT")
	if !ok {
		log.Fatal("PORT environment variable required but not set")
	}
	conf["DATABASE_URL"] = dbURL
	conf["AUTH_SECRET"] = authSecret
	conf["HASH_COST"] = hashCost
	conf["PORT"] = port
	return conf
}

func configurePFH() cors.PreFlightHandler {
	corsFile, err := os.Open("../../config/cors.json")
	if err != nil {
		log.Fatal("Error opening cors.json: " + err.Error())
	}
	defer corsFile.Close()
	byteValue, err := ioutil.ReadAll(corsFile)
	if err != nil {
		log.Fatal("Error reading cors.json: " + err.Error())
	}
	var pfh cors.PreFlightHandler
	err = json.Unmarshal([]byte(byteValue), &pfh)
	if err != nil {
		log.Fatal("cors.json formatted wrongly, error when parsing: " + err.Error())
	}
	return pfh
}
