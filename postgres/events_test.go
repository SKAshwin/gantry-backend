package postgres_test

import(
	"testing"
	"checkin/postgres"
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

}

