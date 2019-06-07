package http

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

//Adapter a function which takes a http handler and adds some functionality
//Before or after it
type Adapter func(http.Handler) http.Handler

//Adapt returns a handler that calls the adapters, in the given adapter,
//before the original http Handler is executed
func Adapt(h http.Handler, adapters ...Adapter) http.Handler {
	for i := len(adapters) - 1; i >= 0; i-- { //need to iterate through in reverse
		adapter := adapters[i]
		h = adapter(h)
	}
	return h
}

//Middleware which intercepts the response being written out, (assuming that it is in a JSON format), parses it to find any strings that are meant
//to be times, and then corrects the time to be of a timezone that is provided in the ?loc argument
//Timezone must follow the IANA time zone database names
//It will run and attempt to alter times if and only if the status code of the message is less than 400
func correctTimezones(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		locationName := r.FormValue("loc")
		location, err := time.LoadLocation(locationName)
		if err != nil {
			WriteMessage(http.StatusBadRequest, "Could not parse location "+locationName+" in form argument loc. Use IANA time zone database names", w)
			return
		}
		h.ServeHTTP(&timeZoneAdjustedWriter{w: w, loc: location}, r)
	})
}

type timeZoneAdjustedWriter struct {
	w          http.ResponseWriter
	loc        *time.Location
	statusCode int
}

func (tzw *timeZoneAdjustedWriter) Header() http.Header {
	return tzw.w.Header()
}

func (tzw *timeZoneAdjustedWriter) WriteHeader(statusCode int) {
	tzw.statusCode = statusCode
	tzw.w.WriteHeader(statusCode)
}

func (tzw *timeZoneAdjustedWriter) Write(b []byte) (int, error) {
	if tzw.statusCode >= 400 { //if it's an error message
		return tzw.w.Write(b)
	}
	//if not, begin picking through the JSON
	jsonString := string(b)
	parts := strings.Split(jsonString, `"`)
	newJSON := ""
	for i, part := range parts {
		var timeVal time.Time
		var amendedPart string
		if err := json.Unmarshal([]byte(`"`+part+`"`), &timeVal); err != nil {
			//this JSON part was not a time value
			amendedPart = part
			//log.Println(amendedPart)
			//log.Println(err)
		} else {
			//this JSON part was a time value
			//log.Println("It was a time!")
			var newTimeVal time.Time
			if !timeVal.IsZero() { //don't touch zero value times (though the output shouldn't have any, by right)
				newTimeVal = timeVal.In(tzw.loc)
			} else {
				newTimeVal = timeVal
			}

			result, _ := json.Marshal(newTimeVal)
			amendedPart = strings.Split(string(result), `"`)[1] //remove quotes from JSON marshalling
			//log.Println(tzw.loc)
			//log.Println(timeVal)
			//log.Println(newTimeVal)
		}

		if i == len(parts)-1 {
			newJSON += amendedPart
		} else {
			newJSON += amendedPart + `"`
		}

	}
	log.Println("Adjusting timezone for " + jsonString + " to " + newJSON)
	return tzw.w.Write([]byte(newJSON))
}
