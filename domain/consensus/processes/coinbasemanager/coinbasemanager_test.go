package coinbasemanager

import "testing"

func TestPowInt64(t *testing.T) {
	tests := []struct {
		base           int64
		exponent       int64
		expectedResult int64
	}{
		{
			base:           0,
			exponent:       0,
			expectedResult: 1,
		},
		{
			base:           0,
			exponent:       1,
			expectedResult: 0,
		},
		{
			base:           1,
			exponent:       1,
			expectedResult: 1,
		},
		{
			base:           1,
			exponent:       2,
			expectedResult: 1,
		},
		{
			base:           2,
			exponent:       1,
			expectedResult: 2,
		},
		{
			base:           2,
			exponent:       2,
			expectedResult: 4,
		},
		{
			base:           3,
			exponent:       2,
			expectedResult: 9,
		},
		{
			base:           2,
			exponent:       3,
			expectedResult: 8,
		},
		{
			base:           3,
			exponent:       3,
			expectedResult: 27,
		},
		{
			base:           5,
			exponent:       11,
			expectedResult: 48828125,
		},
	}

	for _, test := range tests {
		result := powInt64(test.base, test.exponent)
		if result != test.expectedResult {
			t.Errorf("Unexpected result from powInt64. Want: %d, got: %d", test.expectedResult, result)
		}
	}
}
