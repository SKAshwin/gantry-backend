package sha512_test

import (
	"checkin/sha512"
	"checkin/test"
	"testing"
)

func TestCompareHashAndPassword(t *testing.T) {
	hashed, err := sha512.HashMethod{}.HashAndSalt("1234")
	test.Assert(t, err == nil, "sha256.HashAndSalt somehow throws an error: 1234")
	test.Assert(t, sha512.HashMethod{}.CompareHashAndPassword(hashed, "1234"), "sha256.CompareHashAndPassword fails on 1234")

	hashed, err = sha512.HashMethod{}.HashAndSalt("abcd")
	test.Assert(t, err == nil, "sha256.HashAndSalt somehow throws an error: abcd")
	test.Assert(t, sha512.HashMethod{}.CompareHashAndPassword(hashed, "abcd"), "sha256.CompareHashAndPassword fails on abcd")

	hashed, err = sha512.HashMethod{}.HashAndSalt("")
	test.Assert(t, err == nil, "sha256.HashAndSalt somehow throws an error: empty string")
	test.Assert(t, sha512.HashMethod{}.CompareHashAndPassword(hashed, ""), "sha256.CompareHashAndPassword fails on empty string")
}
