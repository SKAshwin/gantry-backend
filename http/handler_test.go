package http_test

import (
	myhttp "checkin/http"
	"checkin/test"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestServeHTTP(t *testing.T) {
	var h myhttp.Handler
	testHandler := func(str string) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reply, _ := json.Marshal(str)
			w.Write(reply)
		})
	}
	testResponse := func(r *http.Response, expected string) {
		var resp string
		err := json.NewDecoder(r.Body).Decode(&resp)
		test.Ok(t, err)
		test.Equals(t, expected, resp)
	}

	h.EventHandler = &myhttp.EventHandler{
		Router: mux.NewRouter(),
	}
	h.EventHandler.Handle("/api/v0/events/blue/da/bu/de", testHandler("I'm an event"))
	r := httptest.NewRequest("POST", "/api/v0/events/blue/da/bu/de", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	testResponse(w.Result(), "I'm an event")

	h.UserHandler = &myhttp.UserHandler{
		Router: mux.NewRouter(),
	}
	h.UserHandler.Handle("/api/v0/users/maybeitsmaybelline", testHandler("I'm a user"))
	r = httptest.NewRequest("POST", "/api/v0/users/maybeitsmaybelline", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	testResponse(w.Result(), "I'm a user")

	h.AuthHandler = &myhttp.AuthHandler{
		Router: mux.NewRouter(),
	}
	h.AuthHandler.Handle("/api/v0/auth/taylorswift", testHandler("I'm a log in service"))
	r = httptest.NewRequest("PATCH", "/api/v0/auth/taylorswift", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	testResponse(w.Result(), "I'm a log in service")

	h.UtilityHandler = &myhttp.UtilityHandler{
		Router: mux.NewRouter(),
	}
	h.UtilityHandler.Handle("/api/v0/utility/henry", testHandler("I'm a utility"))
	r = httptest.NewRequest("DELETE", "/api/v0/utility/henry", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	testResponse(w.Result(), "I'm a utility")

	r = httptest.NewRequest("OPTIONS", "/api/v1/yowhatsthis", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusNotFound, w.Result().StatusCode)
}
