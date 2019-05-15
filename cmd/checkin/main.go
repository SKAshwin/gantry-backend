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
	//loadEnvironmentalVariables() //comment this out for heroku production
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

	authHours, err := strconv.Atoi(config["AUTH_HOURS"])
	if err != nil {
		log.Fatal("Error parsing AUTH_HOURS env variable" + err.Error())
	}

	jwtAuthenticator := jwt.Authenticator{SigningKey: []byte(config["AUTH_SECRET"]), ExpiryTime: time.Duration(authHours) * time.Hour}
	bcryptHashMethod := bcrypt.HashMethod{HashCost: hashCost}
	qrGenerator := qrcode.Generator{Level: qrcode.Highest}
	guestMessenger := websocket.NewGuestMessenger(2048, 2048)

	us := &postgres.UserService{DB: db, HM: bcryptHashMethod}
	as := &postgres.AuthenticationService{DB: db, HM: bcryptHashMethod}
	es := &postgres.EventService{DB: db}
	gs := &postgres.GuestService{DB: db, HM: bcryptHashMethod}

	authHandler := http.NewAuthHandler(as, jwtAuthenticator, us)
	userHandler := http.NewUserHandler(us, jwtAuthenticator)
	guestHandler := http.NewGuestHandler(gs, es, guestMessenger, jwtAuthenticator)
	eventHandler := http.NewEventHandler(es, jwtAuthenticator, guestHandler)
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
	authHours, ok := os.LookupEnv("AUTH_HOURS")
	if !ok {
		log.Fatal("AUTH_HOURS environment variable required but not set")
	}
	hashCost, ok := os.LookupEnv("HASH_COST")
	if !ok {
		log.Fatal("HASH_COST environment variable required but not set")
	}
	port, ok := os.LookupEnv("PORT")
	if !ok {
		log.Fatal("PORT environment variable required but not set")
	}
	origins, ok := os.LookupEnv("ALLOWED_ORIGINS")
	if !ok {
		log.Fatal("ALLOWED_ORIGINS environment variable required but not set")
	}
	methods, ok := os.LookupEnv("ALLOWED_METHODS")
	if !ok {
		log.Fatal("ALLOWED_METHODS environment variable required but not set")
	}
	headers, ok := os.LookupEnv("ALLOWED_HEADERS")
	if !ok {
		log.Fatal("ALLOWED_HEADERS environment variable required but not set")
	}

	conf["DATABASE_URL"] = dbURL
	conf["AUTH_SECRET"] = authSecret
	conf["HASH_COST"] = hashCost
	conf["AUTH_HOURS"] = authHours
	conf["PORT"] = port
	conf["ALLOWED_ORIGINS"] = origins
	conf["ALLOWED_METHODS"] = methods
	conf["ALLOWED_HEADERS"] = headers

	return conf
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
