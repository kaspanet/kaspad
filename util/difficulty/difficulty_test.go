package difficulty_test

import (
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/util/difficulty"
	"testing"
)

func TestGetHashrateString(t *testing.T) {
	var results = map[string]string{
		"kaspa-mainnet": "2 H/s",
		"kaspa-testnet": "131.07 KH/s",
		"kaspa-devnet":  "131.07 KH/s",
		"kaspa-simnet":  "2.00 KH/s",
	}
	testutils.ForAllNets(t, false, func(t *testing.T, params *dagconfig.Params) {
		targetGenesis := difficulty.CompactToBig(params.GenesisBlock.Header.Bits())
		hashrate := difficulty.GetHashrateString(targetGenesis, params.TargetTimePerBlock)
		if hashrate != results[params.Name] {
			t.Errorf("Expected %s, found %s", results[params.Name], hashrate)
		}
	})
}
