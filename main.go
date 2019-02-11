// Package checkin project main.go
package checkin

import (
	"log"
	"net/http"

	"checkin/config"
	"checkin/routing"

	_ "github.com/lib/pq" //this is the main package tf
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
