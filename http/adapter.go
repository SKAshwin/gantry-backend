package http

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"
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

//Middleware which intercepts the response being written, and if any field query string params are supplied
//it will only display those fields of the object being written out
//Also works with arrays of objects, and arrays of arrays of objects, etc
//If the respone being written out is neither an object nor an array (or there are any elements of an array that aren't an object)
//it will just pass through
func jsonSelector(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			log.Println("Could not parse form: " + err.Error())
			WriteMessage(http.StatusBadRequest, "Could not parse form query values", w)
			return
		}
		if len(r.Form["field"]) == 0 { //if no field params, give back all
			h.ServeHTTP(w, r)
		} else {
			h.ServeHTTP(&jsonSelectorWriter{w: w, selectedFields: r.Form["field"]}, r)
		}

	})
}

type jsonSelectorWriter struct {
	w              http.ResponseWriter
	selectedFields []string
	statusCode     int
}

func (jsw *jsonSelectorWriter) Header() http.Header {
	return jsw.w.Header()
}

func (jsw *jsonSelectorWriter) WriteHeader(statusCode int) {
	jsw.statusCode = statusCode
	jsw.w.WriteHeader(statusCode)
}

func (jsw *jsonSelectorWriter) Write(b []byte) (int, error) {
	if jsw.statusCode >= 400 { //if it's an error message
		return jsw.w.Write(b)
	}

	reply, err := selectJSONFields(b, jsw.selectedFields)
	if err != nil {
		//an error here is a programming mistake, as the selectJSONFields method should not fail for any properly formatted
		//object or array of objects (or array of arrays etc)
		//and this middleware should only be used on methods that output JSON request bodies
		log.Println("Error attempting to select JSON fields in middleware: " + err.Error())
		WriteMessage(http.StatusInternalServerError, "Internal error trying to select fields", jsw.w)
		return 200, nil
	}

	log.Println(string(reply))
	log.Println("Made it here!")

	return jsw.w.Write(reply)
}

func selectJSONFields(jsonData []byte, selectedFields []string) ([]byte, error) {
	// Get slice of data with optional leading whitespace removed.
	// See RFC 7159, Section 2 for the definition of JSON whitespace.
	x := bytes.TrimLeft(jsonData, " \t\r\n")

	isArray := len(x) > 0 && x[0] == '['
	isObject := len(x) > 0 && x[0] == '{'

	if isObject {
		var fields map[string]interface{}
		err := json.Unmarshal(jsonData, &fields)
		if err != nil {
			return nil, err
		}
		for name := range fields {
			log.Println(name)
			log.Println(selectedFields)
			if !stringInSlice(name, selectedFields) {
				delete(fields, name) //delete all non-selected fields
			}
		}
		log.Println(fields)
		lol, err := json.Marshal(fields)
		log.Println(string(lol))
		return json.Marshal(fields)
	} else if isArray {
		var elems []interface{}
		err := json.Unmarshal(jsonData, &elems)
		if err != nil {
			return nil, err
		}
		for i, elem := range elems {
			jsonVal, err := json.Marshal(elem)
			if err != nil {
				return nil, err
			}
			newVal, err := selectJSONFields(jsonVal, selectedFields)
			if err != nil {
				return nil, err
			}
			err = json.Unmarshal(newVal, &elems[i])
			if err != nil {
				return nil, err
			}
		}
		return json.Marshal(elems)
	}

	return jsonData, nil
}

//Checks if a string is in the list
//case insensitive
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if strings.ToLower(b) == strings.ToLower(a) {
			return true
		}
	}
	return false
}
