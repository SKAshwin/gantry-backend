package main

import (
	"checkin/bcrypt"
	"checkin/http"
	"checkin/http/cors"
	websocket "checkin/http/gorillawebsocket"
	"checkin/http/jwt"
	"checkin/postgres"
	"checkin/qrcode"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

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

	jwtAuthenticator := jwt.Authenticator{SigningKey: []byte(config["AUTH_SECRET"]), ExpiryTime: time.Duration(toInt(config["AUTH_HOURS"])) * time.Hour}
	bcryptHashMethod := bcrypt.HashMethod{HashCost: toInt(config["HASH_COST"])}
	qrGenerator := qrcode.Generator{Level: qrcode.High}
	guestMessenger := websocket.NewGuestMessenger(2048, 2048)

	us := &postgres.UserService{DB: db, HM: bcryptHashMethod}
	as := &postgres.AuthenticationService{DB: db, HM: bcryptHashMethod}
	es := &postgres.EventService{DB: db}
	gs := &postgres.GuestService{DB: db, HM: bcryptHashMethod}

	authHandler := http.NewAuthHandler(as, jwtAuthenticator, us)
	userHandler := http.NewUserHandler(us, jwtAuthenticator)
	guestHandler := http.NewGuestHandler(gs, es, guestMessenger, jwtAuthenticator, toInt(config["MAX_LENGTH_GUEST_NAME"]),
		toInt(config["MAX_LENGTH_GUEST_TAG"]))
	eventHandler := http.NewEventHandler(es, jwtAuthenticator, guestHandler, toInt(config["MAX_LENGTH_EVENT_NAME"]),
		toInt(config["MAX_LENGTH_EVENT_URL"]), toInt(config["MAX_LENGTH_EVENT_TIMETAG"]))
	utilityHandler := http.NewUtilityHandler(qrGenerator)

	handler := http.Handler{
		AuthHandler:    authHandler,
		EventHandler:   eventHandler,
		UserHandler:    userHandler,
		UtilityHandler: utilityHandler,
	}
	server := http.Server{Handler: &handler, Addr: ":" + config["PORT"], PreFlightHandler: configurePFH(config)}
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

//adds a value with the given name as a key from the environmental variables into
func addConfig(name string, config map[string]string) {
	val, ok := os.LookupEnv(name)
	if !ok {
		log.Fatal(name + " environment variable required but not set")
	}
	config[name] = val
}

func getConfig() map[string]string {
	//reads from environmental variables
	conf := make(map[string]string)
	addConfig("DATABASE_URL", conf)
	addConfig("AUTH_SECRET", conf)
	addConfig("AUTH_HOURS", conf)
	addConfig("HASH_COST", conf)
	addConfig("PORT", conf)
	addConfig("ALLOWED_ORIGINS", conf)
	addConfig("ALLOWED_METHODS", conf)
	addConfig("ALLOWED_HEADERS", conf)
	addConfig("MAX_LENGTH_GUEST_NAME", conf)
	addConfig("MAX_LENGTH_GUEST_TAG", conf)
	addConfig("MAX_LENGTH_EVENT_NAME", conf)
	addConfig("MAX_LENGTH_EVENT_URL", conf)
	addConfig("MAX_LENGTH_EVENT_TIMETAG", conf)

	return conf
}

func toInt(str string) int {
	val, err := strconv.Atoi(str)
	if err != nil {
		log.Fatal("Error parsing " + str + " to int: " + err.Error())
	}
	return val
}

func configurePFH(env map[string]string) cors.PreFlightHandler {
	return cors.PreFlightHandler{
		AllowedOrigins: tokenizeAndTrim(env["ALLOWED_ORIGINS"]),
		AllowedMethods: tokenizeAndTrim(env["ALLOWED_METHODS"]),
		AllowedHeaders: tokenizeAndTrim(env["ALLOWED_HEADERS"]),
	}
}

func tokenizeAndTrim(str string) []string {
	substrs := strings.Split(str, ",")
	for i, s := range substrs {
		substrs[i] = strings.TrimSpace(s)
	}
	return substrs
}
