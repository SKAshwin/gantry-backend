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

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
)

type response struct {
	Message string `json:"message"`
}

const (
	dbhost = "DBHOST"
	dbport = "DBPORT"
	dbuser = "DBUSER"
	dbpass = "DBPASS"
	dbname = "DBNAME"
)

var db *sql.DB
var allowedCorsOrigins = []string{"http://localhost:8080"}
var loginEndPoint = "/api/auth/login"
var usersEndPoint = "/api/app/users"

func main() {
	//redirectLogger()
	loadEnvironmentalVariables()
	initDB()

	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		//check if jwt is sent, extracts information
		//also checks if token is expired; returns 401 if not
		ValidationKeyGetter: keyGetter,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err string) {
			writeError(http.StatusUnauthorized, err, w)
		},
		SigningMethod: jwt.SigningMethodHS256,
	})

	r := mux.NewRouter()
	r.Handle(loginEndPoint, adminLoginHandler).Methods("POST")
	r.Handle(usersEndPoint, isAdmin(jwtMiddleware.Handler(listUsersHandler))).Methods("GET")
	handler := cors.New(cors.Options{
		AllowedOrigins: allowedCorsOrigins,
	}).Handler(r) //only allow GETs POSTs from that address (LOGIN_URL, the client-side address); the bare minimum needed
	http.ListenAndServe(":3000", recoverWrap(handler))
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

func isAdmin(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := getJWTClaims(r)
		if err != nil {
			log.Println("isAdmin faced an error: " + err.Error())
			writeError(http.StatusUnauthorized, "Expired Token", w)
		} else {
			if claims[jwtAdminStatus] == true {
				h.ServeHTTP(w, r)
			} else {
				writeError(http.StatusUnauthorized, "Accessing this page requires admin privileges",
					w)
			}
		}
	})
}

func recoverWrap(h http.Handler) http.Handler {
	//Middleware which handles any panic in the handlers by logging it and reporting an internal
	//server error to the front end
	//intended for unexpected errors (expected, handled errors should not cause a panic)
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
				log.Println("Entered Panic with ", err.Error())
				writeError(http.StatusInternalServerError, err.Error(), w)
			}
		}()
		h.ServeHTTP(w, r)
	})
}

//panicIf If error passed to it is not nil, panics with the error, and logs the desired log message
//The log message should ideally identify the source of the problem
func panicIf(err error, logMessage string) {
	if err != nil {
		log.Println(logMessage)
		panic(err)
	}
}

func writeError(statusCode int, message string, w http.ResponseWriter) {
	if statusCode != http.StatusOK {
		log.Println("writeError:", message)
	}
	w.WriteHeader(statusCode)
	reply, _ := json.Marshal(response{Message: message})
	w.Write(reply)
}

func writeMessage(message string, w http.ResponseWriter) {
	writeError(http.StatusOK, message, w)
}
