package goesi

import (
	"testing"
	"time"
)

func TestGetExpiration(t *testing.T) {
	s := "Thu, 09 Nov 2017 17:27:14 GMT"
	e, err := getExpiration(s)
	if err != nil {
		t.Fail()
	}
	loc, err := time.LoadLocation("GMT")
	if err != nil {
		t.Fail()
	}
	expected := time.Date(2017, time.November, 9, 17, 27, 14, 0, loc)
	if !e.Equal(expected) {
		t.Fatalf("Dates are not equal. Expected: %s, actual: %s", expected, e)
	}
}
