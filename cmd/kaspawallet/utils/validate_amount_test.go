package utils

import (
	"testing"
)

func TestValidateAmountFormat(t *testing.T) {
	validCases := []string{
		"0",
		"1",
		"1.0",
		"0.1",
		"0.12345678",
	}

	for _, testCase := range validCases {
		err := ValidateAmountFormat(testCase)

		if err != nil {
			t.Error(err)
		}
	}

	invalidCases := []string{
		"",
		"a",
		"-1",
		"0.123456789", // 9 decimal digits
		".1",          // decimal but no integer component
	}

	for _, testCase := range invalidCases {
		err := ValidateAmountFormat(testCase)

		if err == nil {
			t.Errorf("Expected an error but succeeded validation for test case %s", testCase)
		}
	}
}
