package utils

import "testing"

// Takes in a string representation of the Kas value to convert to Sompi
func TestKasToSompi(t *testing.T) {
	type testVector struct {
		originalAmount  string
		convertedAmount uint64
	}

	validCases := []testVector{
		{originalAmount: "0", convertedAmount: 0},
		{originalAmount: "1", convertedAmount: 100000000},
		{originalAmount: "33184.1489732", convertedAmount: 3318414897320},
		{originalAmount: "21.35808032", convertedAmount: 2135808032},
		{originalAmount: "184467440737.09551615", convertedAmount: 18446744073709551615},
	}

	for _, currentTestVector := range validCases {
		convertedAmount, err := KasToSompi(currentTestVector.originalAmount)

		if err != nil {
			t.Error(err)
		} else if convertedAmount != currentTestVector.convertedAmount {
			t.Errorf("Expected %s, to convert to %d. Got: %d", currentTestVector.originalAmount, currentTestVector.convertedAmount, convertedAmount)
		}
	}

	invalidCases := []string{
		"184467440737.09551616", // Bigger than max uint64
		"-1",
		"a",
		"",
	}

	for _, currentTestVector := range invalidCases {
		_, err := KasToSompi(currentTestVector)

		if err == nil {
			t.Errorf("Expected an error but succeeded validation for test case %s", currentTestVector)
		}
	}
}

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
		err := validateAmountFormat(testCase)

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
		err := validateAmountFormat(testCase)

		if err == nil {
			t.Errorf("Expected an error but succeeded validation for test case %s", testCase)
		}
	}
}
