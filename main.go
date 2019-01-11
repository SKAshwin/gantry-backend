// registrationApp project main.go
package main

import (
	"log"
	"net/http"

	"registration-app/config"
	"registration-app/routing"

	_ "github.com/lib/pq"
	"github.com/rs/cors"
)

func init() {
	config.LoadEnvironmentalVariables()
	config.InitDB()
}

func main() {
	//config.RedirectLogger()

	r := routing.SetUp()
	corsConfig := config.GetCorsConfig()
	handler := cors.New(cors.Options{
		AllowedOrigins: corsConfig.AllowedOrigins,
		AllowedMethods: corsConfig.AllowedMethods,
		AllowedHeaders: corsConfig.AllowedHeaders,
	}).Handler(r)

	if err := http.ListenAndServe(":3000", routing.RecoverWrap(handler)); err != nil {
		log.Fatal("Could not listen and serve: " + err.Error())
	}
}
