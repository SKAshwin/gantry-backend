package cors

import (
	"net/http"

	"github.com/rs/cors"
)

//PreFlightHandler Contains the configuration information for CORS
type PreFlightHandler struct {
	AllowedOrigins []string `json:"allowedOrigins"`
	AllowedMethods []string `json:"allowedMethods"`
	AllowedHeaders []string `json:"allowedHeaders"`
}

//Handle Takes a http handler, and adapts it to include handling for CORS pre-flight
//requests
//Handles the pre-flight requests using the Config information supplies, such as what
//methods are allowed via CORS
func (pfh PreFlightHandler) Handle(h http.Handler) http.Handler {
	return cors.New(cors.Options{
		AllowedOrigins: pfh.AllowedOrigins,
		AllowedMethods: pfh.AllowedMethods,
		AllowedHeaders: pfh.AllowedHeaders,
	}).Handler(h)
}
