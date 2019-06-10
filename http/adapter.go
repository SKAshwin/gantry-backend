package http

import (
	"encoding/json"
	"io/ioutil"
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
//If no ?loc aergument is given, ensures timezones are in UTC
func correctTimezonesOutput(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		locationName := r.FormValue("loc") //no loc will return "", which will be parsed as UTC
		location, err := time.LoadLocation(locationName)
		if err != nil {
			WriteMessage(http.StatusBadRequest, "Could not parse location "+locationName+" in form argument loc. Use IANA time zone database names", w)
			return
		}
		h.ServeHTTP(&timeZoneAdjustedWriter{w: w, loc: location}, r)
	})
}

//Middleware which intercepts the request, assuming that its body is in a JSON format, and parses it to find any strings that are meant to be times
//and then ensures that the time is interpreted to be of the timezone that is provided in the ?loc form query parameter
//Timezone must follow the IANA time zone database names
//If no ?loc argument is given, times will be interpreted as UTC
func correctTimezonesInput(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		locationName := r.FormValue("loc") //no loc will return "", which will be parsed as UTC
		location, err := time.LoadLocation(locationName)
		if err != nil {
			WriteMessage(http.StatusBadRequest, "Could not parse location "+locationName+" in form argument loc. Use IANA time zone database names", w)
			return
		}
		res, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println("Could not read request body in correct timezone middleware: " + err.Error())
			WriteMessage(http.StatusBadRequest, "Could not read request body to correct timezones", w)
			return
		}
		json := string(res)

		newJSON := correctJSONTimeZone(json, func(old time.Time) time.Time {
			if old.IsZero() {
				return old
			}
			return time.Date(old.Year(), old.Month(), old.Day(), old.Hour(), old.Minute(), old.Second(), old.Nanosecond(), location)
		})

		r.Body = ioutil.NopCloser(strings.NewReader(newJSON))
		r.ContentLength = int64(len(newJSON))

		h.ServeHTTP(w, r)
	})
}

//Writer which adjusts timezones being written to it before sending it to the client
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
	//if not, alter any times in the JSON
	jsonString := string(b)
	newJSON := correctJSONTimeZone(jsonString, func(timeVal time.Time) time.Time {
		var newTimeVal time.Time
		if !timeVal.IsZero() { //don't touch zero value times (though the output shouldn't have any, by right)
			newTimeVal = timeVal.In(tzw.loc)
		} else {
			newTimeVal = timeVal
		}
		return newTimeVal
	})
	log.Println("Adjusting timezone for " + jsonString + " to " + newJSON)
	return tzw.w.Write([]byte(newJSON))
}

func correctJSONTimeZone(jsonString string, correctionMethod func(time.Time) time.Time) string {
	parts := strings.Split(jsonString, `"`)
	newJSON := ""
	for i, part := range parts {
		var timeVal time.Time
		var amendedPart string
		if err := json.Unmarshal([]byte(`"`+part+`"`), &timeVal); err != nil {
			//this JSON part was not a time value
			amendedPart = part
		} else {
			//this JSON part was a time value
			newTimeVal := correctionMethod(timeVal)
			result, _ := json.Marshal(newTimeVal)
			amendedPart = strings.Split(string(result), `"`)[1] //remove quotes from JSON marshalling
		}

		if i == len(parts)-1 {
			newJSON += amendedPart
		} else {
			newJSON += amendedPart + `"`
		}
	}

	return newJSON
}
