package checkin_test

import (
	"checkin"
	"checkin/test"
	"encoding/json"
	"testing"
)

func TestMinuteUnmarshalJSON(t *testing.T) {
	var min checkin.Minute

	//test successful
	err := json.Unmarshal([]byte("18"), &min)
	test.Ok(t, err)
	test.Equals(t, checkin.Minute(18), min)

	//test boundaries
	err = json.Unmarshal([]byte("0"), &min)
	test.Ok(t, err)
	test.Equals(t, checkin.Minute(0), min)

	err = json.Unmarshal([]byte("59"), &min)
	test.Ok(t, err)
	test.Equals(t, checkin.Minute(59), min)

	//test invalid values
	err = json.Unmarshal([]byte("60"), &min)
	test.Assert(t, err != nil, "No error when trying to unmarshal 60 into minute")

	err = json.Unmarshal([]byte("-1"), &min)
	test.Assert(t, err != nil, "No error when trying to unmarshal -1 into minute")

	err = json.Unmarshal([]byte("100"), &min)
	test.Assert(t, err != nil, "No error when trying to unmarshal 100 into minute")

	err = json.Unmarshal([]byte("-14"), &min)
	test.Assert(t, err != nil, "No error when trying to unmarshal -14 into minute")

	//try not even a number
	err = json.Unmarshal([]byte("aword"), &min)
	test.Assert(t, err != nil, "No error when trying to unmarshal aword into minute")
}

func TestHourUnmarshalJSON(t *testing.T) {
	var hour checkin.Hour

	//test successful
	err := json.Unmarshal([]byte("18"), &hour)
	test.Ok(t, err)
	test.Equals(t, checkin.Hour(18), hour)

	//test boundaries
	err = json.Unmarshal([]byte("0"), &hour)
	test.Ok(t, err)
	test.Equals(t, checkin.Hour(0), hour)

	err = json.Unmarshal([]byte("23"), &hour)
	test.Ok(t, err)
	test.Equals(t, checkin.Hour(23), hour)

	err = json.Unmarshal([]byte("24"), &hour)
	test.Assert(t, err != nil, "No error when trying to unmarshal 60 into Hour")

	err = json.Unmarshal([]byte("-1"), &hour)
	test.Assert(t, err != nil, "No error when trying to unmarshal -1 into Hour")

	err = json.Unmarshal([]byte("30"), &hour)
	test.Assert(t, err != nil, "No error when trying to unmarshal 30 into Hour")

	err = json.Unmarshal([]byte("-3"), &hour)
	test.Assert(t, err != nil, "No error when trying to unmarshal -3 into Hour")

	//try not even a number
	err = json.Unmarshal([]byte("aword"), &hour)
	test.Assert(t, err != nil, "No error when trying to unmarshal aword into Hour")
}

func TestTextSizeUnmarshalJSON(t *testing.T) {
	var sz checkin.TextSize

	//test successful
	err := json.Unmarshal([]byte("4"), &sz)
	test.Ok(t, err)
	test.Equals(t, checkin.TextSize(4), sz)

	//test boundaries
	err = json.Unmarshal([]byte("1"), &sz)
	test.Ok(t, err)
	test.Equals(t, checkin.TextSize(1), sz)

	err = json.Unmarshal([]byte("7"), &sz)
	test.Ok(t, err)
	test.Equals(t, checkin.TextSize(7), sz)

	//test invalid values
	err = json.Unmarshal([]byte("8"), &sz)
	test.Assert(t, err != nil, "No error when trying to unmarshal 60 into TextSize")

	err = json.Unmarshal([]byte("0"), &sz)
	test.Assert(t, err != nil, "No error when trying to unmarshal -1 into TextSize")

	err = json.Unmarshal([]byte("20"), &sz)
	test.Assert(t, err != nil, "No error when trying to unmarshal 100 into TextSize")

	err = json.Unmarshal([]byte("-10"), &sz)
	test.Assert(t, err != nil, "No error when trying to unmarshal -14 into TextSize")

	//try not even a number
	err = json.Unmarshal([]byte(" "), &sz)
	test.Assert(t, err != nil, "No error when trying to unmarshal empty space into TextSize")
}

func TestButtonSizeUnmarshalJSON(t *testing.T) {
	var sz checkin.ButtonSize

	//test successful
	err := json.Unmarshal([]byte("3"), &sz)
	test.Ok(t, err)
	test.Equals(t, checkin.ButtonSize(3), sz)

	//test boundaries
	err = json.Unmarshal([]byte("1"), &sz)
	test.Ok(t, err)
	test.Equals(t, checkin.ButtonSize(1), sz)

	err = json.Unmarshal([]byte("4"), &sz)
	test.Ok(t, err)
	test.Equals(t, checkin.ButtonSize(4), sz)

	//test invalid values
	err = json.Unmarshal([]byte("5"), &sz)
	test.Assert(t, err != nil, "No error when trying to unmarshal 60 into ButtonSize")

	err = json.Unmarshal([]byte("0"), &sz)
	test.Assert(t, err != nil, "No error when trying to unmarshal -1 into ButtonSize")

	err = json.Unmarshal([]byte("10"), &sz)
	test.Assert(t, err != nil, "No error when trying to unmarshal 100 into ButtonSize")

	err = json.Unmarshal([]byte("-2"), &sz)
	test.Assert(t, err != nil, "No error when trying to unmarshal -14 into ButtonSize")

	//try not even a number
	err = json.Unmarshal([]byte("ayywtf"), &sz)
	test.Assert(t, err != nil, "No error when trying to unmarshal ayywtf into ButtonSize")
}

func TestPopUpComponentTypeUnmarshalJSON(t *testing.T) {
	var ppct checkin.PopUpComponent

	err := json.Unmarshal([]byte(`{"type":"text", "text":"hello"}`), &ppct)
	test.Ok(t, err)
	test.Equals(t, checkin.PopUpComponentType("text"), ppct)
}
