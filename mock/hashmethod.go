package mock

//HashMethod is a mock implementation of checkin.HashMethod
type HashMethod struct {
	HashAndSaltFn                 func(pwd string) (string, error)
	HashAndSaltInvoked            bool
	CompareHashAndPasswordFn      func(hash string, pwd string) bool
	CompareHashAndPasswordInvoked bool
}

//HashAndSalt invokes the mock function and labels it as such (setting HashAndSaltInvoked to true)
func (hm *HashMethod) HashAndSalt(pwd string) (string, error) {
	hm.HashAndSaltInvoked = true
	return hm.HashAndSaltFn(pwd)
}

//CompareHashAndPassword invokes the mock function and labels it as such (setting HashAndSaltInvoked to true)
func (hm *HashMethod) CompareHashAndPassword(hash string, pwd string) bool {
	hm.CompareHashAndPasswordInvoked = true
	return hm.CompareHashAndPasswordFn(hash, pwd)
}
