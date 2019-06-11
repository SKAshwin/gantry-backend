package postgres_test

import (
	"checkin/postgres"
	"checkin/test"
	"testing"
	"time"
)

func TestUpdateLastLoggedIn(t *testing.T) {
	us := postgres.UserService{DB: db}

	//try normal functionality (with a formerly never logged in user)
	err := us.UpdateLastLoggedIn("ME5Bob")
	test.Ok(t, err)
	user, err := us.User("ME5Bob")
	test.Ok(t, err)
	test.Assert(t, time.Now().Sub(user.LastLoggedIn.Time) < 2*time.Second, "Last logged in time not within 2 seconds of now")

	//try normal functionality (with a user which had formerly logged in)
	err = us.UpdateLastLoggedIn("safosscholar")
	test.Ok(t, err)
	user, err = us.User("safosscholar")
	test.Ok(t, err)
	test.Assert(t, time.Now().Sub(user.LastLoggedIn.Time) < 2*time.Second, "Last logged in time not within 2 seconds of now")

	//make sure other users not changed
	user, err = us.User("TestUser")
	test.Ok(t, err)
	test.Equals(t, false, user.LastLoggedIn.Valid)

	//try user does not exist
	err = us.UpdateLastLoggedIn("lolwut")
	test.Assert(t, err != nil, "No error thrown even though update last log in of user that doesn't exist")

	//try empty string
	err = us.UpdateLastLoggedIn("")
	test.Assert(t, err != nil, "No error thrown even though update last log in of user that doesn't exist (empty string name)")

}
