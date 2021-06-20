package daa

import (
	"fmt"
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

	machineHashNanoseconds := measureMachineHashNanoseconds()
	log.Infof("Machine hash nanoseconds: %d", machineHashNanoseconds)

	targetHashesPerSecond := int64(200_000)
	targetHashNanoseconds := time.Second.Nanoseconds() / targetHashesPerSecond
	runHashes(machineHashNanoseconds, targetHashNanoseconds, 10*time.Second)
}

func measureMachineHashNanoseconds() int64 {
	genesisBlock := dagconfig.DevnetParams.GenesisBlock
	targetDifficulty := difficulty.CompactToBig(genesisBlock.Header.Bits())
	headerForMining := genesisBlock.Header.ToMutable()

	machineHashesPerSecondMeasurementDuration := 10 * time.Second
	isFinished := false
	hashes := int64(0)
	go func() {
		for !isFinished {
			headerForMining.SetNonce(rand.Uint64())
			pow.CheckProofOfWorkWithTarget(headerForMining, targetDifficulty)
			hashes++
		}
	}()
	time.Sleep(machineHashesPerSecondMeasurementDuration)
	isFinished = true

	return machineHashesPerSecondMeasurementDuration.Nanoseconds() / hashes
}

func runHashes(machineHashNanoseconds int64, targetHashNanoseconds int64, runDuration time.Duration) {
	if targetHashNanoseconds < machineHashNanoseconds {
		panic(fmt.Errorf("targetHashNanoseconds (%dns) is faster than "+
			"what the machine is capable of (%dns)", targetHashNanoseconds, machineHashNanoseconds))
	}

	genesisBlock := dagconfig.DevnetParams.GenesisBlock
	targetDifficulty := difficulty.CompactToBig(genesisBlock.Header.Bits())
	headerForMining := genesisBlock.Header.ToMutable()

	isFinished := false
	hashes := int64(0)
	startTime := time.Now()
	go func() {
		for !isFinished {
			targetElapsedTime := hashes * targetHashNanoseconds
			elapsedTime := time.Since(startTime).Nanoseconds()
			if elapsedTime < targetElapsedTime {
				time.Sleep(time.Duration(targetElapsedTime - elapsedTime))
			}

			headerForMining.SetNonce(rand.Uint64())
			pow.CheckProofOfWorkWithTarget(headerForMining, targetDifficulty)
			hashes++
		}
	}()
	time.Sleep(runDuration)
	isFinished = true

	log.Infof("aaaa %f", float64(hashes)/runDuration.Seconds())
}
