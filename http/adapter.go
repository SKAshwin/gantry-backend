package http

import (
	"bytes"
	"encoding/json"
	"net/http"
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
			if !stringInSlice(name, selectedFields) {
				delete(fields, name) //delete all non-selected fields
			}
		}
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
			newElem, err := json.Marshal(newVal)
			if err != nil {
				return nil, err
			}
			elems[i] = newElem
		}
		return json.Marshal(elems)
	}

	return jsonData, nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
