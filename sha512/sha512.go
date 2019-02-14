package sha512

import (
	"crypto/sha512"
)

//HashMethod a SHA512 implementation of HashMethod
type HashMethod struct {
}

//HashAndSalt Hashes a function using SHA512, will never
//return an error
//no salt
func (hm HashMethod) HashAndSalt(pwd string) (string, error) {
	bytes := sha512.Sum512([]byte(pwd)) //needs to be two lines
	//cant call [:] on a non-addressable value
	return string(bytes[:]), nil
}

//CompareHashAndPassword Returns true if the hashed string matches the plaintext
//Returns false otherwise
func (hm HashMethod) CompareHashAndPassword(hash string, pwd string) bool {
	bytes := sha512.Sum512([]byte(pwd))
	return hash == string(bytes[:])
}
