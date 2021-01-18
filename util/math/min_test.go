package math_test

import (
	utilMath "github.com/kaspanet/kaspad/util/math"
	"math"
	"testing"
)

const (
	MaxInt = int(^uint(0) >> 1)
	MinInt = -MaxInt - 1
)

func TestMinInt(t *testing.T) {
	tests := []struct {
		inputs   [2]int
		expected int
	}{
		{[2]int{MaxInt, 0}, 0},
		{[2]int{1, 2}, 1},
		{[2]int{MaxInt, MaxInt}, MaxInt},
		{[2]int{MaxInt, MaxInt - 1}, MaxInt - 1},
		{[2]int{MaxInt, MinInt}, MinInt},
		{[2]int{MinInt, 0}, MinInt},
		{[2]int{MinInt, MinInt}, MinInt},
		{[2]int{0, MinInt + 1}, MinInt + 1},
		{[2]int{0, MinInt}, MinInt},
	}

	for i, test := range tests {
		result := utilMath.MinInt(test.inputs[0], test.inputs[1])
		if result != test.expected {
			t.Fatalf("%d: Expected %d, instead found: %d", i, test.expected, result)
		}
		reverseResult := utilMath.MinInt(test.inputs[1], test.inputs[0])
		if result != reverseResult {
			t.Fatalf("%d: Expected result and reverseResult to be the same, instead: %d!=%d", i, result, reverseResult)
		}
	}
}

func TestMinUint32(t *testing.T) {
	tests := []struct {
		inputs   [2]uint32
		expected uint32
	}{
		{[2]uint32{math.MaxUint32, 0}, 0},
		{[2]uint32{1, 2}, 1},
		{[2]uint32{math.MaxUint32, math.MaxUint32}, math.MaxUint32},
		{[2]uint32{math.MaxUint32, math.MaxUint32 - 1}, math.MaxUint32 - 1},
	}

	for _, test := range tests {
		result := utilMath.MinUint32(test.inputs[0], test.inputs[1])
		if result != test.expected {
			t.Fatalf("Expected %d, instead found: %d", test.expected, result)

		}
		reverseResult := utilMath.MinUint32(test.inputs[1], test.inputs[0])
		if result != reverseResult {
			t.Fatalf("Expected result and reverseResult to be the same, instead: %d!=%d", result, reverseResult)
		}
	}
}
