package daa

import (
	"github.com/kaspanet/kaspad/domain/consensus/utils/pow"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/stability-tests/common"
	"github.com/kaspanet/kaspad/util/difficulty"
	"github.com/kaspanet/kaspad/util/panics"
	"math/rand"
	"testing"
	"time"
)

func TestDAA(t *testing.T) {
	//if os.Getenv("RUN_STABILITY_TESTS") == "" {
	//	t.Skip()
	//}

	defer panics.HandlePanic(log, "daa-main", nil)
	err := parseConfig()
	if err != nil {
		t.Fatalf("error in parseConfig: %s", err)
	}
	defer backendLog.Close()
	common.UseLogger(backendLog, log.Level())

	genesisBlock := dagconfig.DevnetParams.GenesisBlock
	targetDifficulty := difficulty.CompactToBig(genesisBlock.Header.Bits())
	headerForMining := genesisBlock.Header.ToMutable()

	hashesAmountForTiming := 10_000_000
	startTime := time.Now()
	for i := 0; i < hashesAmountForTiming; i++ {
		headerForMining.SetNonce(rand.Uint64())
		pow.CheckProofOfWorkWithTarget(headerForMining, targetDifficulty)
	}
	elapsedTime := time.Since(startTime)
	hashesPerSecond := float64(hashesAmountForTiming) / elapsedTime.Seconds()

	log.Infof("Machine hashes per second: %f", hashesPerSecond)
}
