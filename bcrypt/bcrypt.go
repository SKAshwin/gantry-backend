package bcrypt

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

//HashMethod An implementation of the HashMethod interface
//Using the bcrypt library
type HashMethod struct {
	HashCost int //must be between 4 to 20
	//bigger hash cost takes longer to run
}

//HashAndSalt Hashes a function (with a salt) using the bcrypt algorithm
//With the hashcost specified in the bcrypt.HashMethod instance
func (hm *HashMethod) HashAndSalt(pwd string) (string, error) {
	//Use GenerateFromPassword to hash & salt pwd.
	//cost must be above 4
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), hm.HashCost)
	if err != nil {
		return "", errors.New("Failed to hash password: " + err.Error())
	}
	// GenerateFromPassword returns a byte slice so we need to
	// convert the bytes to a string and return it
	return string(hash), nil
}

//CompareHashAndPassword Returns true if the hashed string matches the plaintex
//Returns false otherwise
func (hm *HashMethod) CompareHashAndPassword(hash string, pwd string) bool {
	//method returns error if hash does not match
	//returns nil otherwise
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pwd)) == nil
}
