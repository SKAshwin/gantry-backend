package postgres_test

import (
	"checkin"
	"checkin/postgres"
	"checkin/test"
	"math"
	"testing"
	"time"
)

func TestCreateEvent(t *testing.T) {
	es := postgres.EventService{DB: db}

	err := es.CreateEvent(event, "TestUser")
	test.Ok(t, err)
}

func TestUpdateEvent(t *testing.T) {
	//NOTE: This method, for the duration of the execution of the test suite, permanently changes the updatedAt value of one of the events
	//DO NOT RELY ON THAT VALUE IN ANY OTHER TEST
	es := postgres.EventService{DB: db}

	event, err := es.Event("aa19239f-f9f5-4935-b1f7-0edfdceabba7")
	test.Ok(t, err)
	test.Assert(t, math.Abs(event.Radius.Float64-5) > 0.0001, "Event radius was already at updated value")
	test.Assert(t, len(event.TimeTags) == 0 && event.TimeTags != nil, "Time tags already at updated value, should be empty array")
	originalRadius := event.Radius
	test.Assert(t, math.Abs(event.UpdatedAt.Sub(time.Now().UTC()).Seconds()) > 2, "Event last updated time already close to current time")
	originalCreatedAt := event.CreatedAt

	event.Radius.Float64 = 5
	event.CreatedAt = time.Now() //this should not actually be processed as an updatable field
	event.TimeTags["ReLeaSe"] = time.Date(2019, 10, 3, 2, 5, 10, 0, time.UTC)
	//testing adding a time tag, make sure that its not case sensitive (should be set to all lowercase)
	err = es.UpdateEvent(event)
	test.Ok(t, err)

	event, err = es.Event("aa19239f-f9f5-4935-b1f7-0edfdceabba7")
	test.Ok(t, err)
	test.Assert(t, math.Abs(5-event.Radius.Float64) < 0.0001, "Radius was not successfully updated")
	test.Assert(t, math.Abs(event.UpdatedAt.Sub(time.Now().UTC()).Seconds()) < 2, "Event last updated not within 2 seconds of now; i.e. not updated")
	test.Assert(t, event.CreatedAt == originalCreatedAt, "Event created at time was modified; this should not be allowed")
	test.Assert(t, event.TimeTags[0] == checkin.TimeTag{Label: "release", Time: time.Date(2019, 10, 3, 2, 5, 10, 0, time.UTC)}, "Time tags were not properly updated")

	event.Radius = originalRadius
	event.TimeTags = nil

	err = es.UpdateEvent(event)
	test.Ok(t, err)

	event, err = es.Event("aa19239f-f9f5-4935-b1f7-0edfdceabba7")
	test.Ok(t, err)
	test.Equals(t, []checkin.TimeTag{}, event.TimeTags) //should not ever have TimeTags set to nil
	//nil should be equivalent to an empty timetag array
	event.TimeTags = []checkin.TimeTag{}
	err = es.UpdateEvent(event)
	test.Ok(t, err)
	test.Equals(t, []checkin.TimeTag{}, event.TimeTags)

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
