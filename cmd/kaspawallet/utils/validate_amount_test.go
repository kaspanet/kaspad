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
		"111111111111.11111111", // 12 digits to the left of decimal, 8 digits to the right
		"184467440737.09551615", // Maximum input that can be represented in sompi later
		"184467440737.09551616", // Cannot be represented in sompi, but we'll acccept for "correct format"
		"999999999999.99999999", // Cannot be represented in sompi, but we'll acccept for "correct format"
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
		"0.123456789",           // 9 decimal digits
		".1",                    // decimal but no integer component
		"0a",                    // Extra character
		"0000000000000",         // 13 zeros
		"012",                   // Int padded with zero
		"00.1",                  // Decimal padded with zeros
		"111111111111111111111", // all digits
		"111111111111A11111111", // non-period/non-digit where decimal would be
		"000000000000.00000000", // all zeros
		"kaspa",                 // all text
	}

	for _, testCase := range invalidCases {
		err := ValidateAmountFormat(testCase)

		if err == nil {
			t.Errorf("Expected an error but succeeded validation for test case %s", testCase)
		}
	}
}
