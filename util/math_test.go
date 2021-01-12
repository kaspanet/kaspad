package util_test

import (
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/util"
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
		actualResult := util.FastLog2Floor(test.n)

		if test.expectedResult != actualResult {
			t.Errorf("TestFastLog2Floor: %d: expected result: %d but got: %d", test.n, test.expectedResult, actualResult)
		}
	}
}

func TestGetHashrateString(t *testing.T) {
	var results = map[string]string{
		"kaspa-mainnet": "2 H/s",
		"kaspa-testnet": "131.07 KH/s",
		"kaspa-devnet":  "131.07 KH/s",
		"kaspa-simnet":  "2.00 KH/s",
	}
	testutils.ForAllNets(t, false, func(t *testing.T, params *dagconfig.Params) {
		targetGenesis := util.CompactToBig(params.GenesisBlock.Header.Bits())
		hashrate := util.GetHashrateString(targetGenesis, params.TargetTimePerBlock)
		if hashrate != results[params.Name] {
			t.Errorf("Expected %s, found %s", results[params.Name], hashrate)
		}
	})
}
