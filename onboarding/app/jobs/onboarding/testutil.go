package onboarding

import (
	"testing"
)

func assert(t *testing.T, result bool, message string, args ...interface{}) {
	if !result {
		t.Errorf(message, args...)
	}
}

func assertEqual(t *testing.T, value1 interface{}, value2 interface{}, message string) {
	if value1 != value2 {
		t.Errorf(message, value1, value2)
	}
}

func assertIsNil(t *testing.T, value interface{}, message string) {
	if value != nil {
		t.Errorf(message, value)
	}
}
