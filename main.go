// registrationApp project main.go
package main

import (
	"net/http"

	"registration-app/config"
	"registration-app/routing"

	_ "github.com/lib/pq"
	"github.com/rs/cors"
)

var allowedCorsOrigins = []string{"http://localhost:8080"}

func init() {
	config.LoadEnvironmentalVariables()
	config.InitDB()
}

func main() {
	//config.RedirectLogger()

	r := routing.SetUp()

	handler := cors.New(cors.Options{
		AllowedOrigins: allowedCorsOrigins,
		AllowedMethods: []string{"GET", "POST", "DELETE", "PUT"},
		AllowedHeaders: []string{"*"},
	}).Handler(r)

	http.ListenAndServe(":3000", routing.RecoverWrap(handler))
}
