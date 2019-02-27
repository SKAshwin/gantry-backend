package bcrypt_test

import (
	"checkin/bcrypt"
	"checkin/test"
	"testing"
)

func TestCompareHashAndPassword(t *testing.T) {
	hashed, err := bcrypt.HashMethod{}.HashAndSalt("1234")
	test.Assert(t, err == nil, "bcrypt.HashAndSalt somehow throws an error: 1234")
	test.Assert(t, bcrypt.HashMethod{}.CompareHashAndPassword(hashed, "1234"), "bcrypt.CompareHashAndPassword fails on 1234")

	hashed, err = bcrypt.HashMethod{}.HashAndSalt("abcd")
	test.Assert(t, err == nil, "bcrypt.HashAndSalt somehow throws an error: abcd")
	test.Assert(t, bcrypt.HashMethod{}.CompareHashAndPassword(hashed, "abcd"), "bcrypt.CompareHashAndPassword fails on abcd")

	hashed, err = bcrypt.HashMethod{}.HashAndSalt("")
	test.Assert(t, err == nil, "bcrypt.HashAndSalt somehow throws an error: empty string")
	test.Assert(t, bcrypt.HashMethod{}.CompareHashAndPassword(hashed, ""), "bcrypt.CompareHashAndPassword fails on empty string")
}
