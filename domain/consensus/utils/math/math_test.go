package math

import (
	"math"
	"math/big"
	"testing"
)

func TestFastLog2Floor(t *testing.T) {
	tests := []struct {
		n              uint64
		expectedResult uint8
	}{
		{1, 0},
		{2, 1},
		{3, 1},
		{4, 2},
		{5, 2},
		{16, 4},
		{31, 4},
		{1684234, 20},
		{4294967295, 31}, // math.MaxUint32 (2^32 - 1)
		{4294967296, 32}, // 2^32
		{4294967297, 32}, // 2^32 + 1
		{4611686018427387904, 62},
		{9223372036854775808, 63},  // 2^63
		{18446744073709551615, 63}, // math.MaxUint64 (2^64 - 1).
	}

	for _, test := range tests {
		actualResult := FastLog2Floor(test.n)

		if test.expectedResult != actualResult {
			t.Errorf("TestFastLog2Floor: %d: expected result: %d but got: %d", test.n, test.expectedResult, actualResult)
		}
	}
}

// TestBigToCompact ensures BigToCompact converts big integers to the expected
// compact representation.
func TestBigToCompact(t *testing.T) {
	tests := []struct {
		in  string
		out uint32
	}{
		{"0", 0},
		{"-1", 25231360},
		{"9223372036854775807", 142606335},
		{"922337203685477580712312312123487", 237861256},
	}

	for x, test := range tests {
		n := new(big.Int)
		n.SetString(test.in, 10)
		r := BigToCompact(n)
		if r != test.out {
			t.Errorf("TestBigToCompact test #%d failed: got %d want %d\n",
				x, r, test.out)
			return
		}
	}
}

// TestCompactToBig ensures CompactToBig converts numbers using the compact
// representation to the expected big integers.
func TestCompactToBig(t *testing.T) {
	tests := []struct {
		in  uint32
		out string
	}{
		{0, "0"},
		{10000000, "0"},
		{math.MaxUint32, "-6311914495863998658485429352026283268468573753812676234178171506285465200675957" +
			"87397376951158770808349115367298981082112562162319027637583517246275967980671962038665775867893645140" +
			"22856089959012026469381002722748489975264028415685723882208353467651862351803217528553851158828320170" +
			"89832330727351553686808317476632783024236208492771822246700842318520468733521003756809213629548010354" +
			"33865968377930773213939300289069292503211567790599147939718451689002543278625341832829837474611074167" +
			"86700705915281593002614032021233542099318559748885883681365573294332856023451874423425211080847063825" +
			"199113186681992371681311588352",
		},
		{142606335, "9223370937343148032"},
		{25231360, "-1"},
		{237861256, "922337129789886856855791696084992"},
	}

	for i, test := range tests {
		n := CompactToBig(test.in)
		if n.String() != test.out {
			t.Errorf("TestCompactToBig test #%d failed: got %s want %s",
				i, n, test.out)
			return
		}
	}
}
