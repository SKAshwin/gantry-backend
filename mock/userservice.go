package mock

import (
	"checkin"
)

//UserService represents a mock implementation of the checkin.UserService interface
type UserService struct {
	UserFn      func(username string) (checkin.User, error)
	UserInvoked bool

	UsersFn      func() ([]checkin.User, error)
	UsersInvoked bool

	CreateUserFn      func(u checkin.User) error
	CreateUserInvoked bool

	DeleteUserFn      func(username string) error
	DeleteUserInvoked bool

	UpdateUserFn      func(username string, updateFields map[string]string) (bool, error)
	UpdateUserInvoked bool

	CheckIfExistsFn      func(username string) (bool, error)
	CheckIfExistsInvoked bool

	UpdateLastLoggedInFn      func(username string) error
	UpdateLastLoggedInInvoked bool
}

//User invokes the mock implementation and marks the function as invoked
func (us *UserService) User(username string) (checkin.User, error) {
	us.UserInvoked = true
	return us.UserFn(username)
}

//Users invokes the mock implementation and marks the function as invoked
func (us *UserService) Users() ([]checkin.User, error) {
	us.UsersInvoked = true
	return us.UsersFn()
}

//CreateUser invokes the mock implementation and marks the function as invoked
func (us *UserService) CreateUser(u checkin.User) error {
	us.CreateUserInvoked = true
	return us.CreateUserFn(u)
}

//DeleteUser invokes the mock implementation and marks the function as invoked
func (us *UserService) DeleteUser(username string) error {
	us.DeleteUserInvoked = true
	return us.DeleteUserFn(username)
}

//UpdateUser invokes the mock implementation and marks the function as invoked
func (us *UserService) UpdateUser(username string, updateFields map[string]string) (bool, error) {
	us.UpdateUserInvoked = true
	return us.UpdateUserFn(username, updateFields)
}

//CheckIfExists invokes the mock implementation and marks the function as invoked
func (us *UserService) CheckIfExists(username string) (bool, error) {
	us.CheckIfExistsInvoked = true
	return us.CheckIfExistsFn(username)
}

//UpdateLastLoggedIn invokes the mock implementation and marks the function as invoked
func (us *UserService) UpdateLastLoggedIn(username string) error {
	us.UpdateLastLoggedInInvoked = true
	return us.UpdateLastLoggedInFn(username)
}
