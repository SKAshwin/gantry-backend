package postgres_test

import(
	"testing"
	"checkin/postgres"
	"checkin"
	"checkin/test"
	"time"
	"math"
)

func TestUpdateEvent(t *testing.T) {
	//NOTE: This method, for the duration of the execution of the test suite, permanently changes the updatedAt value of one of the events
	//DO NOT RELY ON THAT VALUE IN ANY OTHER TEST
	es := postgres.EventService{DB: db}

	event, err := es.Event("aa19239f-f9f5-4935-b1f7-0edfdceabba7")
	test.Ok(t, err)
	test.Assert(t, math.Abs(event.Radius.Float64-5)>0.0001, "Event radius was already at updated value")
	originalRadius := event.Radius
	test.Assert(t, math.Abs(event.UpdatedAt.Sub(time.Now().UTC()).Seconds())>2, "Event last updated time already close to current time")
	originalCreatedAt := event.CreatedAt

	event.Radius.Float64 = 5
	event.CreatedAt = time.Now() //this should not actually be processed as an updatable field
	err = es.UpdateEvent(event)
	test.Ok(t, err)

	event, err = es.Event("aa19239f-f9f5-4935-b1f7-0edfdceabba7")
	test.Ok(t, err)
	test.Assert(t, math.Abs(5 - event.Radius.Float64)<0.0001, "Radius was not successfully updated")
	test.Assert(t, math.Abs(event.UpdatedAt.Sub(time.Now().UTC()).Seconds())<2, "Event last updated not within 2 seconds of now; i.e. not updated")
	test.Assert(t, event.CreatedAt == originalCreatedAt, "Event created at time was modified; this should not be allowed")

	event.Radius = originalRadius
	
	err = es.UpdateEvent(event)
	test.Ok(t, err)

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
			ID: "ec5c5f6f-5384-4406-9beb-73b9effbdf50",
			NRIC: "A1234",
			Survey: []checkin.FeedbackFormItem{
				checkin.FeedbackFormItem{
					Question: "A",
					Answer: "AA1",
				},
				checkin.FeedbackFormItem{
					Question: "B",
					Answer: "BB1",
				},
				checkin.FeedbackFormItem{
					Question: "C",
					Answer: "CC1",
				},	
			},
			SubmitTime: time.Date(2019,time.April, 11, 8, 18, 14, 0, time.UTC),
		},
		checkin.FeedbackForm{
			ID: "663fd6e1-b781-49e7-b1ed-dd0e3c6ff28e",
			NRIC: "B5678",
			Survey: []checkin.FeedbackFormItem{
				checkin.FeedbackFormItem{
					Question: "A",
					Answer: "AA2",
				},
				checkin.FeedbackFormItem{
					Question: "B",
					Answer: "BB2",
				},
				checkin.FeedbackFormItem{
					Question: "C",
					Answer: "CC2",
				},	
			},
			SubmitTime: time.Date(2019,time.April, 11, 9, 32, 4, 0, time.UTC),
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
