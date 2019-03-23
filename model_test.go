package checkin_test

import (
	"checkin"
	"checkin/test"
	"testing"
	"time"

	"github.com/guregu/null"
)

func TestReleased(t *testing.T) {
	//Event has no release date set
	var event checkin.Event
	test.Equals(t, true, event.Released())

	//Event released in the past
	event = checkin.Event{
		Release: null.Time{Time: time.Now().UTC().Add(-1 * time.Minute), Valid: true},
	}
	test.Equals(t, true, event.Released())

	//Event released in the future
	event = checkin.Event{
		Release: null.Time{Time: time.Now().UTC().Add(5 * time.Minute), Valid: true},
	}
	test.Equals(t, false, event.Released())
}
