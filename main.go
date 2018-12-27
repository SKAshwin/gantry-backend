// registrationApp project main.go
package main

import (
	"net/http"

	_ "github.com/lib/pq"
	"github.com/rs/cors"
)

var allowedCorsOrigins = []string{"http://localhost:8080"}

func main() {
	//redirectLogger()
	LoadEnvironmentalVariables()
	InitDB()

	r := SetUpRouting()

	handler := cors.New(cors.Options{
		AllowedOrigins: allowedCorsOrigins,
	}).Handler(r) //only allow GETs POSTs from that address (LOGIN_URL, the client-side address); the bare minimum needed

	http.ListenAndServe(":3000", RecoverWrap(handler))
}
