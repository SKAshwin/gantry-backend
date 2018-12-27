// registrationApp project main.go
package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"

	"registration-app/config"

	_ "github.com/lib/pq"
	"github.com/rs/cors"
)

var allowedCorsOrigins = []string{"http://localhost:8080"}

func main() {
	//redirectLogger()
	LoadEnvironmentalVariables()
	config.InitDB()

	r := SetUpRouting()

	handler := cors.New(cors.Options{
		AllowedOrigins: allowedCorsOrigins,
	}).Handler(r) //only allow GETs POSTs from that address (LOGIN_URL, the client-side address); the bare minimum needed

	http.ListenAndServe(":3000", RecoverWrap(handler))
}

func LoadEnvironmentalVariables() {
	err := godotenv.Load()
	if err != nil {
		log.Print("Error loading environmental variables: ")
		log.Fatal(err.Error())
	}
}
