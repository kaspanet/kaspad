package util

import "testing"

// TestToCamelCase tests whether ToCamelCase correctly converts camelCase-ish strings to camelCase.
func TestToCamelCase(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedResult string
	}{
		{
			name:           "single word that's already in camelCase",
			input:          "abc",
			expectedResult: "abc",
		},
		{
			name:           "single word in PascalCase",
			input:          "Abc",
			expectedResult: "abc",
		},
		{
			name:           "single word in all caps",
			input:          "ABC",
			expectedResult: "abc",
		},
		{
			name:           "multiple words that are already in camelCase",
			input:          "aaaBbbCcc",
			expectedResult: "aaaBbbCcc",
		},
		{
			name:           "multiple words in PascalCase",
			input:          "AaaBbbCcc",
			expectedResult: "aaaBbbCcc",
		},
		{
			name:           "acronym in start position",
			input:          "AAABbbCcc",
			expectedResult: "aaaBbbCcc",
		},
		{
			name:           "acronym in middle position",
			input:          "aaaBBBCcc",
			expectedResult: "aaaBbbCcc",
		},
		{
			name:           "acronym in end position",
			input:          "aaaBbbCCC",
			expectedResult: "aaaBbbCcc",
		},
	}

	for _, test := range tests {
		result := ToCamelCase(test.input)
		if result != test.expectedResult {
			t.Errorf("ToCamelCase for test \"%s\" returned an unexpected result. "+
				"Expected: \"%s\", got: \"%s\"", test.name, test.expectedResult, result)
		}
	}
}
