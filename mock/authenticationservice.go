package mock

//AuthenticationService represents a mock implementation of the checkin.AuthenticationService interface
type AuthenticationService struct {
	AuthenticateFn      func(username string, pwdPlaintext string, isAdmin bool) (bool, error)
	AuthenticateInvoked bool
}

//Authenticate marks the AuthenticateFn as true, and then invokes it with the arguments
func (as *AuthenticationService) Authenticate(username string, pwdPlaintext string, isAdmin bool) (bool, error) {
	as.AuthenticateInvoked = true
	return as.AuthenticateFn(username, pwdPlaintext, isAdmin)
}
