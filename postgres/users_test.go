package postgres_test

import (
	"checkin"
	"checkin/mock"
	"checkin/postgres"
	"checkin/test"
	"errors"
	"testing"
	"time"

	"github.com/guregu/null"
)

func TestUpdateLastLoggedIn(t *testing.T) {
	us := postgres.UserService{DB: db}

	//try normal functionality (with a formerly never logged in user)
	err := us.UpdateLastLoggedIn("ME5Bob")
	test.Ok(t, err)
	user, err := us.User("ME5Bob")
	test.Ok(t, err)
	test.Assert(t, time.Since(user.LastLoggedIn.Time) < 2*time.Second && time.Since(user.LastLoggedIn.Time) > 0, "Last logged in time not within 2 seconds of now")

	//try normal functionality (with a user which had formerly logged in)
	err = us.UpdateLastLoggedIn("safosscholar")
	test.Ok(t, err)
	user, err = us.User("safosscholar")
	test.Ok(t, err)
	test.Assert(t, time.Since(user.LastLoggedIn.Time) < 2*time.Second && time.Since(user.LastLoggedIn.Time) > 0, "Last logged in time not within 2 seconds of now")

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

func TestCreateUser(t *testing.T) {
	var hm mock.HashMethod
	hm.HashAndSaltFn = hashFnGenerator(nil)
	hm.CompareHashAndPasswordFn = compareHashAndPasswordGenerator()
	us := postgres.UserService{DB: db, HM: &hm}

	//test standard functionality, with fields that should be ignored
	password := "password"
	user := checkin.User{
		Username:          "ElvenAshwin",
		PasswordPlaintext: &password,
		PasswordHash:      "someHash", // should be ignored,
		Name:              "Ashwin",
		CreatedAt:         time.Date(2018, 12, 16, 15, 0, 0, 0, time.UTC),               //should be ignored
		UpdatedAt:         time.Date(2018, 12, 31, 6, 0, 0, 0, time.UTC),                //should be ignored
		LastLoggedIn:      null.TimeFrom(time.Date(2019, 3, 1, 11, 30, 0, 0, time.UTC)), //should be ignored
	}

	err := us.CreateUser(user)
	test.Ok(t, err)
	fetched, err := us.User("ElvenAshwin")
	test.Ok(t, err)

	test.Equals(t, user.Username, fetched.Username)
	test.Equals(t, true, hm.CompareHashAndPassword(fetched.PasswordHash, *user.PasswordPlaintext))
	test.Equals(t, (*string)(nil), fetched.PasswordPlaintext) //the plaintext password should seriously not be around anymore
	test.Equals(t, user.Name, fetched.Name)
	test.Assert(t, time.Since(fetched.CreatedAt) < 2*time.Second && time.Since(fetched.CreatedAt) > 0, "Created at time is not within 2 seconds of now")
	test.Assert(t, time.Since(fetched.UpdatedAt) < 2*time.Second && time.Since(fetched.UpdatedAt) > 0, "Created at time is not within 2 seconds of now")
	test.Equals(t, false, fetched.LastLoggedIn.Valid) //last logged should be null

	err = us.DeleteUser("ElvenAshwin")
	test.Ok(t, err)

	//test supplying only the fields needed
	user = checkin.User{
		Username:          "ElvenAshwin",
		PasswordPlaintext: &password,
		Name:              "Ashwin",
	}

	err = us.CreateUser(user)
	test.Ok(t, err)
	fetched, err = us.User("ElvenAshwin")
	test.Ok(t, err)

	test.Equals(t, user.Username, fetched.Username)
	test.Equals(t, true, hm.CompareHashAndPassword(fetched.PasswordHash, *user.PasswordPlaintext))
	test.Equals(t, (*string)(nil), fetched.PasswordPlaintext) //the plaintext password should seriously not be around anymore
	test.Equals(t, user.Name, fetched.Name)
	test.Assert(t, time.Now().Sub(fetched.CreatedAt) < 2*time.Second, "Created at time is not within 2 seconds of now")
	test.Assert(t, time.Now().Sub(fetched.UpdatedAt) < 2*time.Second, "Created at time is not within 2 seconds of now")
	test.Equals(t, false, fetched.LastLoggedIn.Valid) //last logged should be null

	err = us.DeleteUser("ElvenAshwin")
	test.Ok(t, err)

	//test hash fails
	hm.HashAndSaltFn = hashFnGenerator(errors.New("An error"))
	err = us.CreateUser(user)
	test.Assert(t, err != nil, "Hash failing does not throw an error")

}

func TestUpdateUser(t *testing.T) {
	us := postgres.UserService{DB: db}

	originalUser, err := us.User("AirForceMan")
	test.Ok(t, err)

	//try no new password plaintext first
	newUser := checkin.User{
		Username:     "Jasmine",
		PasswordHash: "ahash", //should be used
		Name:         "Wei Juan",
		CreatedAt:    time.Date(2018, 12, 16, 15, 0, 0, 0, time.UTC),                //should be ignored
		UpdatedAt:    time.Date(2018, 12, 31, 6, 0, 0, 0, time.UTC),                 //should be ignored
		LastLoggedIn: null.TimeFrom(time.Date(2019, 5, 31, 15, 30, 0, 0, time.UTC)), //should be ignored
	}
	err = us.UpdateUser(originalUser.Username, newUser)
	test.Ok(t, err)
	updated, err := us.User("Jasmine")
	test.Ok(t, err)
	test.Equals(t, newUser.Username, updated.Username)
	test.Equals(t, newUser.Name, updated.Name)
	test.Equals(t, newUser.PasswordHash, updated.PasswordHash)
	test.Equals(t, originalUser.LastLoggedIn, updated.LastLoggedIn)
	test.Assert(t, time.Since(updated.UpdatedAt) < 2*time.Second && time.Since(updated.UpdatedAt) > 0, "Updated at time not updated with update")

	//try updating password via plaintext
	var hm mock.HashMethod
	hm.HashAndSaltFn = hashFnGenerator(nil)
	hm.CompareHashAndPasswordFn = compareHashAndPasswordGenerator()
	us.HM = &hm
	password := "mygirlfriendleftme"
	newUser = updated
	newUser.PasswordPlaintext = &password
	newUser.PasswordHash = "ahash" //should be ignored
	err = us.UpdateUser(updated.Username, newUser)
	test.Ok(t, err)
	updated2, err := us.User("Jasmine")
	test.Ok(t, err)
	test.Equals(t, true, hm.CompareHashAndPassword(updated2.PasswordHash, *newUser.PasswordPlaintext))

	//test hashing fails
	hm.HashAndSaltFn = hashFnGenerator(errors.New("An error"))
	err = us.UpdateUser(updated2.Username, newUser)
	test.Assert(t, err != nil, "No error returned even though hashing of plaintext password failed")
}
