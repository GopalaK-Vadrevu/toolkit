package toolkit

import "testing"

func TestTools_RandomString(t *testing.T) {
	var testTools Tools
	randomString := testTools.RandomString(10)
	if len(randomString) != 10 {
		t.Errorf("Expected length of random string to be 10, but got %d", len(randomString))
	}
}
