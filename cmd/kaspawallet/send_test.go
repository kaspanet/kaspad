package main

import "testing"

func TestKasToSompi(t *testing.T) {
	type testVector struct {
		originalAmount  float64
		convertedAmount uint64
	}

	testVectors := []testVector{
		{originalAmount: 0, convertedAmount: 0},
		{originalAmount: 1, convertedAmount: 100000000},
		{originalAmount: 33184.1489732, convertedAmount: 3318414897320},
		{originalAmount: 21.35808032, convertedAmount: 2135808032},
		{originalAmount: 184467440737.09551615, convertedAmount: 18446744073709551615},
	}

	for _, currentTestVector := range testVectors {
		if kasToSompi(currentTestVector.originalAmount) != currentTestVector.convertedAmount {
			t.Fail()
			t.Logf("Expected %.8f, to convert to %d. Got: %d", currentTestVector.originalAmount, currentTestVector.convertedAmount, kasToSompi(currentTestVector.originalAmount))
		}
	}
}
