package http_test

import (
	"checkin"
	myhttp "checkin/http"
	"checkin/test"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHandleCurrentTime(t *testing.T) {
	var qrg checkin.QRGenerator
	h := myhttp.NewUtilityHandler(qrg)

	//Test normal behavior
	r := httptest.NewRequest("GET", "/api/v1-3/utility/time", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	var curTime time.Time
	err := json.NewDecoder(w.Result().Body).Decode(&curTime)
	test.Ok(t, err)
	test.Assert(t, time.Now().Sub(curTime) < 2*time.Second, "Current time... not within 2 seconds of now")
	test.Equals(t, time.UTC, curTime.Location())

	//Test valid locations (using IANA time zone database names)
	r = httptest.NewRequest("GET", "/api/v1-3/utility/time?loc=Asia/Singapore", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	err = json.NewDecoder(w.Result().Body).Decode(&curTime)
	test.Ok(t, err)
	test.Assert(t, time.Now().Sub(curTime) < 2*time.Second, "Current time... not within 2 seconds of now")
	test.Equals(t, "+0800", strings.Fields(curTime.String())[2])

	r = httptest.NewRequest("GET", "/api/v1-3/utility/time?loc=UTC", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	err = json.NewDecoder(w.Result().Body).Decode(&curTime)
	test.Ok(t, err)
	test.Assert(t, time.Now().Sub(curTime) < 2*time.Second, "Current time... not within 2 seconds of now")
	test.Equals(t, time.UTC, curTime.Location())

	//test invalid location
	r = httptest.NewRequest("GET", "/api/v1-3/utility/time?loc=Asia/China", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	test.Equals(t, http.StatusBadRequest, w.Result().StatusCode)

	//test random other form values (should work fine)
	r = httptest.NewRequest("GET", "/api/v1-3/utility/time?lolwut=Asia/Singapore", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	err = json.NewDecoder(w.Result().Body).Decode(&curTime)
	test.Ok(t, err)
	test.Assert(t, time.Now().Sub(curTime) < 2*time.Second, "Current time... not within 2 seconds of now")
	test.Equals(t, time.UTC, curTime.Location())

	r = httptest.NewRequest("GET", "/api/v1-3/utility/time?loc=Asia/Singapore&anything=1", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	err = json.NewDecoder(w.Result().Body).Decode(&curTime)
	test.Ok(t, err)
	test.Assert(t, time.Now().Sub(curTime) < 2*time.Second, "Current time... not within 2 seconds of now")
	test.Equals(t, "+0800", strings.Fields(curTime.String())[2])

}
