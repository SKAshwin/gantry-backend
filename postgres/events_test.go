package postgres_test

import (
	"checkin"
	"checkin/postgres"
	"checkin/test"
	"log"
	"math"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/guregu/null"
)

func TestEvent(t *testing.T) {
	es := postgres.EventService{DB: db}

	//test normal functionality
	event, err := es.Event("2c59b54d-3422-4bdb-824c-4125775b44c8")
	test.Ok(t, err)
	event.UpdatedAt = time.Time{}
	event.CreatedAt = time.Time{}
	test.Equals(t, checkin.Event{
		ID:       "2c59b54d-3422-4bdb-824c-4125775b44c8",
		Name:     "Data Science CoP",
		URL:      "cop2018",
		TimeTags: map[string]time.Time{"release": time.Date(2019, 4, 12, 9, 0, 0, 0, time.UTC)},
	}, event)

	//test multiple timetags
	event, err = es.Event("3820a980-a207-4738-b82b-45808fe7aba8")
	test.Ok(t, err)
	event.UpdatedAt = time.Time{}
	event.CreatedAt = time.Time{}
	test.Equals(t, checkin.Event{
		ID:       "3820a980-a207-4738-b82b-45808fe7aba8",
		Name:     "SDB Cohesion",
		URL:      "sdbcohesionnovember",
		TimeTags: map[string]time.Time{"release": time.Date(2019, 10, 8, 12, 0, 0, 0, time.UTC), "formrelease": time.Date(2019, 10, 8, 16, 30, 12, 0, time.UTC)},
	}, event)

	//test no time tags
	event, err = es.Event("03293b3b-df83-407e-b836-fb7d4a3c4966")
	test.Ok(t, err)
	event.UpdatedAt = time.Time{}
	event.CreatedAt = time.Time{}
	test.Equals(t, checkin.Event{
		ID:       "03293b3b-df83-407e-b836-fb7d4a3c4966",
		Name:     "CSSCOM Planning Seminar",
		URL:      "csscom",
		TimeTags: map[string]time.Time{},
	}, event)

	//test no such event exists
	_, err = es.Event("810b6bf0-a29b-405b-82ee-e482924f8faa")
	test.Assert(t, err != nil, "No error thrown when trying to fetch event that does not exist")

	//test invalid UUID
	_, err = es.Event("")
	test.Assert(t, err != nil, "No error thrown when trying to fetch event that does not exist (invalid UUID)")
}

func TestEventsBy(t *testing.T) {
	es := postgres.EventService{DB: db}

	//test normal functionality
	events, err := es.EventsBy("TestUser")
	test.Ok(t, err)
	sort.Slice(events, func(i, j int) bool {
		return events[i].Name < events[j].Name
	})
	test.Equals(t, []checkin.Event{
		checkin.Event{
			ID:        "aa19239f-f9f5-4935-b1f7-0edfdceabba7",
			Name:      "Data Science Department Talk",
			URL:       "dsdjan2019",
			Start:     null.TimeFrom(time.Date(2019, 1, 10, 15, 0, 0, 0, time.UTC)),
			End:       null.TimeFrom(time.Date(2019, 1, 10, 18, 0, 0, 0, time.UTC)),
			TimeTags:  map[string]time.Time{},
			CreatedAt: time.Date(2019, 4, 1, 4, 5, 36, 0, time.UTC),
			UpdatedAt: time.Date(2019, 4, 10, 3, 2, 11, 0, time.UTC),
			Lat:       null.FloatFrom(1.335932),
			Long:      null.FloatFrom(103.744708),
			Radius:    null.FloatFrom(0.5),
		},
		checkin.Event{
			ID:        "3820a980-a207-4738-b82b-45808fe7aba8",
			Name:      "SDB Cohesion",
			URL:       "sdbcohesionnovember",
			TimeTags:  map[string]time.Time{"release": time.Date(2019, 10, 8, 12, 0, 0, 0, time.UTC), "formrelease": time.Date(2019, 10, 8, 16, 30, 12, 0, time.UTC)},
			CreatedAt: time.Date(2019, 5, 31, 02, 15, 22, 0, time.UTC),
			UpdatedAt: time.Date(2019, 5, 31, 13, 02, 11, 0, time.UTC),
		},
	}, events)

	//test user hosts no events
	events, err = es.EventsBy("ME6Alice")
	test.Ok(t, err)
	test.Equals(t, []checkin.Event{}, events)

	//test user does not exist
	events, err = es.EventsBy("wdqdwqd")
	test.Ok(t, err)
	test.Equals(t, []checkin.Event{}, events)
}

func TestCreateEvent(t *testing.T) {
	es := postgres.EventService{DB: db}

	singapore, err := time.LoadLocation("Asia/Singapore")
	test.Ok(t, err)
	event := checkin.Event{
		ID:        uuid.New().String(),
		Name:      "Test",
		TimeTags:  make(map[string]time.Time),
		URL:       "hello",
		Long:      null.FloatFrom(10.0),
		Start:     null.TimeFrom(time.Date(2019, 10, 11, 11, 2, 1, 0, singapore)), //test how non-UTC times are handled
		UpdatedAt: time.Date(2015, 2, 3, 1, 2, 0, 0, time.UTC),                    //these fields should be ignored by the create funtion
		CreatedAt: time.Date(2014, 2, 3, 1, 2, 0, 0, time.UTC),                    //^
	}
	event.TimeTags["releAsE"] = time.Date(2019, 5, 4, 3, 2, 1, 0, time.UTC)
	event.TimeTags["FORMRELEASE"] = time.Date(2019, 6, 5, 12, 3, 2, 0, singapore) //test non-UTC times in the timetag
	err = es.CreateEvent(event, "TestUser")
	test.Ok(t, err)

	fetched, err := es.Event(event.ID)
	test.Ok(t, err)
	log.Println(fetched.Start)
	test.Equals(t, time.Date(2019, 5, 4, 3, 2, 1, 0, time.UTC), fetched.TimeTags["release"])
	test.Equals(t, time.Date(2019, 6, 5, 4, 3, 2, 0, time.UTC), fetched.TimeTags["formrelease"])
	test.Assert(t, time.Since(fetched.CreatedAt) < 2*time.Second && time.Since(fetched.CreatedAt) > 0, "Create time not within 2 seconds before now; not properly set")
	test.Assert(t, fetched.CreatedAt == fetched.UpdatedAt, "UpdatedAt time not set to be equal to the CreatedAt time upon creation")
	event.UpdatedAt = fetched.UpdatedAt
	event.CreatedAt = fetched.CreatedAt
	event.TimeTags = fetched.TimeTags
	event.Start = null.TimeFrom(time.Date(2019, 10, 11, 3, 2, 1, 0, time.UTC))
	test.Equals(t, event, fetched)

	err = es.DeleteEvent(event.ID)
	test.Ok(t, err)

	//test nil timetags
	event.TimeTags = nil
	err = es.CreateEvent(event, "TestUser")
	test.Ok(t, err)
	fetched, err = es.Event(event.ID)
	test.Ok(t, err)
	test.Equals(t, make(map[string]time.Time), fetched.TimeTags)

	err = es.DeleteEvent(event.ID)
	test.Ok(t, err)

	//test empty map time tags, should be same result as nil
	event.TimeTags = make(map[string]time.Time)
	err = es.CreateEvent(event, "TestUser")
	test.Ok(t, err)
	fetched, err = es.Event(event.ID)
	test.Ok(t, err)
	test.Equals(t, make(map[string]time.Time), fetched.TimeTags)

	err = es.DeleteEvent(event.ID)
	test.Ok(t, err)

	//test missing ID
	event.ID = ""
	err = es.CreateEvent(event, "TestUser")
	test.Assert(t, err != nil, "No error when creating an event without an ID")

	//test host does not exist
	err = es.CreateEvent(event, "Notauser")
	test.Assert(t, err != nil, "No error when creating an event without an ID")
}

func TestUpdateEvent(t *testing.T) {
	//NOTE: This method, for the duration of the execution of the test suite, permanently changes the updatedAt value of one of the events
	//DO NOT RELY ON THAT VALUE IN ANY OTHER TEST
	es := postgres.EventService{DB: db}

	asmara, err := time.LoadLocation("Africa/Asmara") //off by 3
	test.Ok(t, err)

	event, err := es.Event("aa19239f-f9f5-4935-b1f7-0edfdceabba7")
	test.Ok(t, err)
	test.Assert(t, math.Abs(event.Radius.Float64-5) > 0.0001, "Event radius was already at updated value")
	test.Assert(t, len(event.TimeTags) == 0 && event.TimeTags != nil, "Time tags already at updated value, should be empty array")
	originalRadius := event.Radius
	test.Assert(t, math.Abs(event.UpdatedAt.Sub(time.Now().UTC()).Seconds()) > 2, "Event last updated time already close to current time")
	originalCreatedAt := event.CreatedAt

	event.Radius.Float64 = 5
	event.CreatedAt = time.Now()                                                  //this should not actually be processed as an updatable field
	event.TimeTags["ReLeaSe"] = time.Date(2019, 10, 3, 2, 5, 10, 0, time.UTC)     //testing adding a time tag, make sure that its not case sensitive (should be set to all lowercase)
	event.TimeTags["formrelease"] = time.Date(2019, 10, 3, 12, 15, 30, 0, asmara) //test non-UTC time
	err = es.UpdateEvent(event)
	test.Ok(t, err)

	event, err = es.Event("aa19239f-f9f5-4935-b1f7-0edfdceabba7")
	test.Ok(t, err)
	test.Assert(t, math.Abs(5-event.Radius.Float64) < 0.0001, "Radius was not successfully updated")
	test.Assert(t, math.Abs(event.UpdatedAt.Sub(time.Now().UTC()).Seconds()) < 2, "Event last updated not within 2 seconds of now; i.e. not updated")
	test.Assert(t, event.CreatedAt == originalCreatedAt, "Event created at time was modified; this should not be allowed")
	test.Assert(t, event.TimeTags["release"] == time.Date(2019, 10, 3, 2, 5, 10, 0, time.UTC) && len(event.TimeTags) == 2, "Time tags were not properly updated")
	test.Assert(t, event.TimeTags["formrelease"] == time.Date(2019, 10, 3, 9, 15, 30, 0, time.UTC) && len(event.TimeTags) == 2, "Time tags were not properly updated (non-UTC timezone issue)")

	event.Radius = originalRadius
	event.TimeTags = nil

	err = es.UpdateEvent(event)
	test.Ok(t, err)

	event, err = es.Event("aa19239f-f9f5-4935-b1f7-0edfdceabba7")
	test.Ok(t, err)
	test.Equals(t, make(map[string]time.Time, 0), event.TimeTags) //should not ever have TimeTags set to nil
	//nil should be equivalent to an empty map
	event.TimeTags = make(map[string]time.Time)
	err = es.UpdateEvent(event)
	test.Ok(t, err)
	test.Equals(t, make(map[string]time.Time), event.TimeTags)

	//test no such event with that event ID
	event = checkin.Event{ID: "a6db3963-5389-4dbe-8fc6-bbd7f7ce66b8"}
	err = es.UpdateEvent(event)
	test.Assert(t, err != nil, "No event with given UUID fails to throw error")

}

func TestFeedbackForms(t *testing.T) {
	es := postgres.EventService{DB: db}
	ff, err := es.FeedbackForms("aa19239f-f9f5-4935-b1f7-0edfdceabba7")
	test.Ok(t, err)
	expected := []checkin.FeedbackForm{
		checkin.FeedbackForm{
			ID:   "ec5c5f6f-5384-4406-9beb-73b9effbdf50",
			Name: "Alice",
			Survey: []checkin.FeedbackFormItem{
				checkin.FeedbackFormItem{
					Question: "A",
					Answer:   "AA1",
				},
				checkin.FeedbackFormItem{
					Question: "B",
					Answer:   "BB1",
				},
				checkin.FeedbackFormItem{
					Question: "C",
					Answer:   "CC1",
				},
			},
			SubmitTime: time.Date(2019, time.April, 11, 8, 18, 14, 0, time.UTC),
		},
		checkin.FeedbackForm{
			ID:   "663fd6e1-b781-49e7-b1ed-dd0e3c6ff28e",
			Name: "Bob",
			Survey: []checkin.FeedbackFormItem{
				checkin.FeedbackFormItem{
					Question: "A",
					Answer:   "AA2",
				},
				checkin.FeedbackFormItem{
					Question: "B",
					Answer:   "BB2",
				},
				checkin.FeedbackFormItem{
					Question: "C",
					Answer:   "CC2",
				},
			},
			SubmitTime: time.Date(2019, time.April, 11, 9, 32, 4, 0, time.UTC),
		},
	}
	test.Equals(t, expected, ff)

	//try event does not exist
	ff, err = es.FeedbackForms("663fd6e1-b781-49e7-b1ed-dd0e3c6ff28e")
	test.Assert(t, err != nil, "Failed to throw error when non-existent event accessed")
	ff, err = es.FeedbackForms("fdfdsfdsf")
	test.Assert(t, err != nil, "Failed to throw error when non-existent event accessed (invalid UUID")

	//try event has no feedback forms
	ff, err = es.FeedbackForms("03293b3b-df83-407e-b836-fb7d4a3c4966")
	test.Equals(t, []checkin.FeedbackForm{}, ff)
}

func TestSubmitFeedback(t *testing.T) {
	es := postgres.EventService{DB: db}

	ff := checkin.FeedbackForm{
		ID:   "3c7381e2-2459-401e-9b0d-763a2d9cd93d",
		Name: "Caleb",
		Survey: []checkin.FeedbackFormItem{
			checkin.FeedbackFormItem{
				Question: "A",
				Answer:   "A",
			},
			checkin.FeedbackFormItem{
				Question: "B",
				Answer:   "B",
			},
		},
		SubmitTime: time.Date(2019, time.April, 11, 8, 18, 14, 0, time.UTC), //this should be ignored
	}

	//test submitting to an event with no existing forms
	err := es.SubmitFeedback("03293b3b-df83-407e-b836-fb7d4a3c4966", ff)
	test.Ok(t, err)
	readFF, err := es.FeedbackForms("03293b3b-df83-407e-b836-fb7d4a3c4966")
	test.Ok(t, err)
	test.Assert(t, len(readFF) == 1, "Event has more than one feedback form, which is not expected")
	test.Assert(t, readFF[0].ID != "", "No UUID was generated for submitted feedback form ID")
	test.Assert(t, readFF[0].ID != ff.ID, "Provided UUID was used; all UUIDs should be generated in database, not provided externally")
	test.Equals(t, ff.Name, readFF[0].Name)
	test.Equals(t, ff.Survey, readFF[0].Survey)
	test.Assert(t, math.Abs(readFF[0].SubmitTime.Sub(time.Now().UTC()).Seconds()) < 2,
		"Form not submitted within 2 seconds of now; i.e. submit time set incorrectly") //possibly because submit time was not ignored
	//in implementation

	//test submitting to an event with an existing form
	err = es.SubmitFeedback("2c59b54d-3422-4bdb-824c-4125775b44c8", ff)
	test.Ok(t, err)
	readFF, err = es.FeedbackForms("2c59b54d-3422-4bdb-824c-4125775b44c8")
	var insertedFF checkin.FeedbackForm
	for _, form := range readFF {
		if form.ID != "a6db3963-5389-4dbe-8fc6-bbd7f7ce66b8" { //the only form before we inserted this one
			insertedFF = form
		}
	}
	test.Assert(t, insertedFF.ID != "", "Submitted feedback form could not be found")
	test.Equals(t, ff.Name, insertedFF.Name)
	test.Equals(t, ff.Survey, insertedFF.Survey)
	test.Assert(t, math.Abs(insertedFF.SubmitTime.Sub(time.Now().UTC()).Seconds()) < 2,
		"Form not submitted within 2 seconds of now; i.e. submit time set incorrectly")

	//test event does not exist
	err = es.SubmitFeedback("1440a8c0-2212-430c-bb71-4d7bf3a42862", ff)
	test.Assert(t, err != nil, "Failed to throw error when submitting feedback to non-existent event")

	//test survey is nil - should throw an error
	ff.Survey = nil
	err = es.SubmitFeedback("03293b3b-df83-407e-b836-fb7d4a3c4966", ff)
	test.Assert(t, err != nil, "Nil survey does not throw an error")

	//test empty survey - should be rejected, almost definitely never intended behavior
	ff.Survey = []checkin.FeedbackFormItem{}
	err = es.SubmitFeedback("03293b3b-df83-407e-b836-fb7d4a3c4966", ff)
	test.Assert(t, err != nil, "Empty survey does not throw an error")

}

func TestCheckIfExists(t *testing.T) {
	es := postgres.EventService{DB: db}

	//test event exists
	exists, err := es.CheckIfExists("c14a592c-950d-44ba-b173-bbb9e4f5c8b4")
	test.Ok(t, err)
	test.Equals(t, true, exists)

	//test event does not exist
	exists, err = es.CheckIfExists("812d513d-8bb1-4216-93f5-17bd3056fff4")
	test.Ok(t, err)
	test.Equals(t, false, exists)

	//test invalid UUID
	exists, err = es.CheckIfExists("helloworld")
	test.Ok(t, err)
	test.Equals(t, false, exists)

	//test empty string
	exists, err = es.CheckIfExists("")
	test.Ok(t, err)
	test.Equals(t, false, exists)

}
