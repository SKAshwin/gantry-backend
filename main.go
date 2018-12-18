// registrationApp project main.go
package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	"golang.org/x/crypto/bcrypt"
)

var hashCost = 5 //cost must be above 4, the larger you make it the slower the hash function will run
var allowedCorsOrigins = []string{"http://localhost:8080"}
var loginURL = "/api/auth/login"
var signingKey = []byte("theSecretPassword")
var loginTokenUsername, loginTokenExpiryTime = "username", "exp"

const (
	dbhost = "DBHOST"
	dbport = "DBPORT"
	dbuser = "DBUSER"
	dbpass = "DBPASS"
	dbname = "DBNAME"
)

var db *sql.DB

func main() {
	//redirectLogger()
	loadEnvironmentalVariables()
	initDB()
	r := mux.NewRouter()
	r.Handle(loginURL, recoverWrap(loginHandler)).Methods("POST")
	handler := cors.New(cors.Options{
		AllowedOrigins: allowedCorsOrigins,
	}).Handler(r) //only allow GETs POSTs from that address (LOGIN_URL, the client-side address); the bare minimum needed
	http.ListenAndServe(":3000", handler) //PostGres listens on 5432
}

func loadEnvironmentalVariables() {
	err := godotenv.Load()
	if err != nil {
		log.Print("Error loading environmental variables: ")
		log.Fatal(err.Error())
	}
}

func redirectLogger() {
	//redirects logger output to a logger file
	//for use in production
	file, err := os.OpenFile("info.log", os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()
	log.SetOutput(file)
}

func initDB() {
	//initializes the db variable
	//forms a connection to the database
	config := dbConfig()
	var err error
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		config[dbhost], config[dbport],
		config[dbuser], config[dbpass], config[dbname])

	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("Could not open databse")
	}
	err = db.Ping()
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

func recoverWrap(h http.Handler) http.Handler {
	//Middleware which handles any error in the api by logging it and reporting an internal server error to
	//the front end
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		defer func() {
			r := recover()
			if r != nil {
				switch t := r.(type) {
				case string:
					err = errors.New(t)
				case error:
					err = t
				default:
					err = errors.New("Unknown error")
				}
				log.Println(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}()
		h.ServeHTTP(w, r)
	})
}

type loginDetails struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type response struct {
	Message string `json:"message"`
}

var loginHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var ld loginDetails
	err := decoder.Decode(&ld)
	if err != nil {
		panic(err)
	}
	fmt.Println(ld.Username)

	if authenticate(ld) {
		jwtToken := createToken(ld)
		reply, _ := json.Marshal(map[string]string{"accessToken": jwtToken})
		w.Write(reply)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		reply, _ := json.Marshal(response{Message: "Incorrect Username or Password"})
		w.Write(reply)
	}

})

func createToken(user loginDetails) string {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims[loginTokenUsername] = user.Username
	claims[loginTokenExpiryTime] = time.Now().Add(time.Hour).Unix()

	tokenSigned, err := token.SignedString(signingKey)
	if err != nil {
		panic(err)
	}

	return tokenSigned
}

func authenticate(user loginDetails) bool {
	//still in testing
	if user.Username == "admin567" {
		return bcrypt.CompareHashAndPassword([]byte("$2a$05$Is.BydwHRaXnXTB5rVFDQerDElDYS6Qbl4KH.T5fVyTvdQHXWNZTS"), []byte(user.Password)) == nil //method returns nil if there is a match between password and hash
		//current password is password
	}
	return false
}

func hashAndSalt(pwd []byte) string {
	// Use GenerateFromPassword to hash & salt pwd.
	//cost must be above 4
	hash, err := bcrypt.GenerateFromPassword(pwd, hashCost)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	// GenerateFromPassword returns a byte slice so we need to
	// convert the bytes to a string and return it
	return string(hash)
}

var notImplemented = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Not Implemented")
})
