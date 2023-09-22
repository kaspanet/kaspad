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
