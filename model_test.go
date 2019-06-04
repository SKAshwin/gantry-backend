package checkin_test

import (
	"checkin"
	"checkin/test"
	"testing"
)

func TestIsEmpty(t *testing.T) {
	var guest checkin.Guest
	test.Equals(t, true, guest.IsEmpty())

	guest.Tags = make([]string, 0)
	test.Equals(t, false, guest.IsEmpty())
}
