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
	err = json.Unmarshal([]byte(``), &min)
	test.Assert(t, err != nil, "No error when trying to unmarshal empty string into minute")
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
	err = json.Unmarshal([]byte(``), &hour)
	test.Assert(t, err != nil, "No error when trying to unmarshal empty string into Hour")
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
	err = json.Unmarshal([]byte(``), &sz)
	test.Assert(t, err != nil, "No error when trying to unmarshal empty string into TextSize")
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
	err = json.Unmarshal([]byte(`"ayywtf"`), &sz)
	test.Assert(t, err != nil, "No error when trying to unmarshal ayywtf into ButtonSize")
	err = json.Unmarshal([]byte(``), &sz)
	test.Assert(t, err != nil, "No error when trying to unmarshal empty string into ButtonSize")
}

func TestPopUpComponentTypeUnmarshalJSON(t *testing.T) {
	var ppct checkin.PopUpComponentType

	err := json.Unmarshal([]byte(`"text"`), &ppct)
	test.Ok(t, err)
	test.Equals(t, checkin.PopUpComponentType("text"), ppct)

	err = json.Unmarshal([]byte(`"img"`), &ppct)
	test.Ok(t, err)
	test.Equals(t, checkin.PopUpComponentType("img"), ppct)

	//try non valid value
	err = json.Unmarshal([]byte(`"somethingelse"`), &ppct)
	test.Assert(t, err != nil, "No error trying to marshal invalid pop up component type")

	//test invalid types
	err = json.Unmarshal([]byte(`text`), &ppct)
	test.Assert(t, err != nil, "No error when trying to marshal non-JSON string")
	err = json.Unmarshal([]byte(`8`), &ppct) //this is also only of length 1 - might be an edge case
	test.Assert(t, err != nil, "No error when trying to marshal non-JSON string (number)")
	err = json.Unmarshal([]byte(`""`), &ppct)
	test.Assert(t, err != nil, "No error when trying to empty JSON string")
	err = json.Unmarshal([]byte(``), &ppct)
	test.Assert(t, err != nil, "No error when trying to marshall empty JSON")

}

func TestButtonTypeUnmarshalJSON(t *testing.T) {
	var bt checkin.ButtonType

	err := json.Unmarshal([]byte(`"link"`), &bt)
	test.Ok(t, err)
	test.Equals(t, checkin.ButtonType("link"), bt)

	err = json.Unmarshal([]byte(`"popup"`), &bt)
	test.Ok(t, err)
	test.Equals(t, checkin.ButtonType("popup"), bt)

	//try non valid value
	err = json.Unmarshal([]byte(`"img"`), &bt)
	test.Assert(t, err != nil, "No error trying to marshal invalid button type")

	//test invalid types
	err = json.Unmarshal([]byte(`text`), &bt)
	test.Assert(t, err != nil, "No error when trying to marshal non-JSON string")
	err = json.Unmarshal([]byte(`8`), &bt) //this is also only of length 1 - might be an edge case
	test.Assert(t, err != nil, "No error when trying to marshal non-JSON string (number)")
	err = json.Unmarshal([]byte(`""`), &bt)
	test.Assert(t, err != nil, "No error when trying to empty JSON string")
	err = json.Unmarshal([]byte(``), &bt)
	test.Assert(t, err != nil, "No error when trying to marshall empty JSON")
}

func TestQuestionTypeUnmarshalJSON(t *testing.T) {
	var qt checkin.QuestionType

	err := json.Unmarshal([]byte(`"scaled"`), &qt)
	test.Ok(t, err)
	test.Equals(t, checkin.QuestionType("scaled"), qt)

	err = json.Unmarshal([]byte(`"rating"`), &qt)
	test.Ok(t, err)
	test.Equals(t, checkin.QuestionType("rating"), qt)

	err = json.Unmarshal([]byte(`"radio"`), &qt)
	test.Ok(t, err)
	test.Equals(t, checkin.QuestionType("radio"), qt)
	err = json.Unmarshal([]byte(`"open"`), &qt)
	test.Ok(t, err)
	test.Equals(t, checkin.QuestionType("open"), qt)

	//try non valid value
	err = json.Unmarshal([]byte(`"slider"`), &qt)
	test.Assert(t, err != nil, "No error trying to marshal invalid question type")

	//test invalid types
	err = json.Unmarshal([]byte(`text`), &qt)
	test.Assert(t, err != nil, "No error when trying to marshal non-JSON string")
	err = json.Unmarshal([]byte(`8`), &qt) //this is also only of length 1 - might be an edge case
	test.Assert(t, err != nil, "No error when trying to marshal non-JSON string (number)")
	err = json.Unmarshal([]byte(`""`), &qt)
	test.Assert(t, err != nil, "No error when trying to empty JSON string")
	err = json.Unmarshal([]byte(``), &qt)
	test.Assert(t, err != nil, "No error when trying to marshall empty JSON")
}

func TestGuestSiteUnmarshalJSON(t *testing.T) {
	var site checkin.GuestSite

	expectedSite := checkin.GuestSite{
		Details: checkin.DetailsSection{
			Logo: "",
			Title: checkin.TextElement{
				Content: "title",
				Size:    checkin.TextSize(2),
			},
			Tagline: checkin.TextElement{
				Content: "tagline",
				Size:    checkin.TextSize(5),
			},
			Items: []checkin.DetailItem{
				checkin.DetailItem{
					Title:   "Date",
					Content: "1 Jan 2020",
				},
				checkin.DetailItem{
					Title:   "Time",
					Content: "1300 - 1700",
				},
				checkin.DetailItem{
					Title:   "Venue",
					Content: "Kranji Camp 3 Blk 822, 90 Choa Chu Kang Way, Singapore 688264",
				},
				checkin.DetailItem{
					Title:   "Attire",
					Content: "Office Wear",
				},
			},
		},
		Main: checkin.ButtonSection{
			Icon: true,
			Schedule: checkin.ScheduleSection{
				Display: true,
				Pages: []checkin.SchedulePage{
					checkin.SchedulePage{
						Subject: "Menu A",
						Items: []checkin.ScheduleItem{
							checkin.ScheduleItem{
								Topic: "Item A",
								Start: checkin.TimeElement{},
								End:   checkin.TimeElement{},
								Abstract: checkin.OptionalContent{
									Display: true,
									Content: "hello",
								},
								Speaker: checkin.OptionalProfileItem{},
							},
							checkin.ScheduleItem{
								Topic: "Item B",
								Start: checkin.TimeElement{Hour: 12, Minute: 0},
								End:   checkin.TimeElement{Hour: 14, Minute: 0},
								Abstract: checkin.OptionalContent{
									Display: true,
									Content: "hello",
								},
								Speaker: checkin.OptionalProfileItem{
									Display: true,
									Content: checkin.ProfileItem{
										Image: "",
										Name:  "name",
										Title: "title",
										About: "about",
									},
								},
							},
						},
					},
					checkin.SchedulePage{
						Subject: "Menu B",
						Items: []checkin.ScheduleItem{
							checkin.ScheduleItem{
								Topic: "Item A",
								Start: checkin.TimeElement{},
								End:   checkin.TimeElement{},
								Abstract: checkin.OptionalContent{
									Display: false,
									Content: "hello",
								},
								Speaker: checkin.OptionalProfileItem{
									Display: true,
								},
							},
							checkin.ScheduleItem{
								Topic: "Item B",
								Start: checkin.TimeElement{Hour: 12, Minute: 0},
								End:   checkin.TimeElement{Hour: 14, Minute: 0},
								Abstract: checkin.OptionalContent{
									Display: false,
									Content: "hello",
								},
								Speaker: checkin.OptionalProfileItem{
									Display: false,
									Content: checkin.ProfileItem{
										Image: "",
										Name:  "name",
										Title: "title",
										About: "about",
									},
								},
							},
						},
					},
				},
			},
			Size: checkin.ButtonSize(4),
			ButtonRows: checkin.ButtonMatrix([]checkin.ButtonColumn{
				checkin.ButtonColumn{
					Buttons: []checkin.ButtonElement{
						checkin.ButtonElement{
							Title: "Link",
							Type:  checkin.ButtonType("link"),
							Icon:  "a1a.svg",
							Link:  "https://www.google.com",
							PopUp: checkin.PopUp([]checkin.PopUpComponent{
								checkin.PopUpComponent{
									Type:  checkin.PopUpComponentType("text"),
									Text:  "",
									Image: "",
								},
							}),
						},
						checkin.ButtonElement{
							Title: "Popup",
							Type:  checkin.ButtonType("popup"),
							Icon:  "a1b.svg",
							Link:  "",
							PopUp: checkin.PopUp([]checkin.PopUpComponent{
								checkin.PopUpComponent{
									Type:  checkin.PopUpComponentType("text"),
									Text:  "Hello",
									Image: "",
								},
							}),
						},
					},
				},
			}),
		},
		Survey: checkin.SurveySection([]checkin.QuestionElement{
			checkin.QuestionElement{
				Type:     checkin.QuestionType("scaled"),
				Question: "The length of the event was just nice.",
			},
			checkin.QuestionElement{
				Type:     checkin.QuestionType("rating"),
				Question: "How would you rate this event?",
			},
			checkin.QuestionElement{
				Type:     checkin.QuestionType("radio"),
				Question: "Is this your first time attending this event?",
			},
			checkin.QuestionElement{
				Type:     checkin.QuestionType("open"),
				Question: "Any other comments.",
			},
		}),
	}
	err := json.Unmarshal([]byte(`{
		"details": {
			"logo": "",
			"title": {
				"cont": "title",
				"sz": 2
			},
			"tagline": {
				"cont": "tagline",
				"sz": 5
			},
			"items": [
				{
					"title": "Date",
					"cont": "1 Jan 2020"
				},{
					"title": "Time",
					"cont": "1300 - 1700"
				},{
					"title": "Venue",
					"cont": "Kranji Camp 3 Blk 822, 90 Choa Chu Kang Way, Singapore 688264"
				},{
					"title": "Attire",
					"cont": "Office Wear"
				}
			]
		},
		"main":{
			"icon": true,
			"schedule": {
				"check": true,
				"menus": [
					{
						"subject": "Menu A",
						"items": [
							{
								"topic": "Item A",
								"start": {"hour": 0, "mins": 0},
								"end": {"hour": 0, "mins": 0},
								"abstract": {"check": true, "cont": "hello"},
								"speaker": {"check": false, "cont": {"img": "", "name": "", "title": "", "about": ""}}
							},{
								"topic": "Item B",
								"start": {"hour": 12, "mins": 0},
								"end": {"hour": 14, "mins": 0},
								"abstract": {"check": true, "cont": "hello"},
								"speaker": {"check": true, "cont": {"img": "", "name": "name", "title": "title", "about": "about"}}
							}
						]
					},{
						"subject": "Menu B",
						"items": [
							{
								"topic": "Item A",
								"start": {"hour": 0, "mins": 0},
								"end": {"hour": 0, "mins": 0},
								"abstract": {"check": false, "cont": "hello"},
								"speaker": {"check": true, "cont": {"img": "", "name": "", "title": "", "about": ""}}
							},{
								"topic": "Item B",
								"start": {"hour": 12, "mins": 0},
								"end": {"hour": 14, "mins": 0},
								"abstract": {"check": false, "cont": "hello"},
								"speaker": {"check": false, "cont": {"img": "", "name": "name", "title": "title", "about": "about"}}
							}
						]
					}
				]
			},
			"sz": 4,
			"rows": [
				{
					"cols": [{
						"title": "Link",
						"type": "link",
						"icon": "a1a.svg",
						"link": "https://www.google.com",
						"popup": [{"type":"text", "text":"", "img":""}]
						},{
						"title": "Popup",
						"type": "popup",
						"icon": "a1b.svg",
						"link": "",
						"popup": [{"type":"text", "text":"Hello", "img":""}]
					}]
				}
			]
		},
		"survey": [{
			"type": "scaled",
			"question": "The length of the event was just nice."
		},{
			"type": "rating",
			"question": "How would you rate this event?"
		},{
			"type": "radio",
			"question": "Is this your first time attending this event?"
		},{
			"type": "open",
			"question": "Any other comments."
		}]
	}`), &site)
	test.Ok(t, err)
	test.Equals(t, expectedSite, site)

}
