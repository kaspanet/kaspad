package main

import (
	"fmt"
	"github.com/kaspanet/kaspad/domain/consensus/utils/pow"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/stability-tests/common"
	"github.com/kaspanet/kaspad/util/difficulty"
	"github.com/kaspanet/kaspad/util/panics"
	"math/rand"
	"time"
)

func main() {
	defer panics.HandlePanic(log, "daa-main", nil)
	err := parseConfig()
	if err != nil {
		panic(fmt.Errorf("error in parseConfig: %s", err))
	}
	defer backendLog.Close()
	common.UseLogger(backendLog, log.Level())

	machineHashNanoseconds := measureMachineHashNanoseconds()
	log.Infof("Machine hashes per second: %d", hashNanosecondsToHashesPerSecond(machineHashNanoseconds))

	targetHashNanoseconds := machineHashNanoseconds * 10
	runHashes(targetHashNanoseconds, 10*time.Second)
}

func hashNanosecondsToHashesPerSecond(hashNanoseconds int64) int64 {
	return time.Second.Nanoseconds() / hashNanoseconds
}

func measureMachineHashNanoseconds() int64 {
	genesisBlock := dagconfig.DevnetParams.GenesisBlock
	targetDifficulty := difficulty.CompactToBig(genesisBlock.Header.Bits())
	headerForMining := genesisBlock.Header.ToMutable()

	machineHashesPerSecondMeasurementDuration := 10 * time.Second
	hashes := int64(0)
	runForDuration(machineHashesPerSecondMeasurementDuration, func() {
		headerForMining.SetNonce(rand.Uint64())
		pow.CheckProofOfWorkWithTarget(headerForMining, targetDifficulty)
		hashes++
	})

	return machineHashesPerSecondMeasurementDuration.Nanoseconds() / hashes
}

func runHashes(targetHashNanoseconds int64, runDuration time.Duration) {
	genesisBlock := dagconfig.DevnetParams.GenesisBlock
	targetDifficulty := difficulty.CompactToBig(genesisBlock.Header.Bits())
	headerForMining := genesisBlock.Header.ToMutable()

	hashes := int64(0)
	startTime := time.Now()
	runForDuration(runDuration, func() {
		targetElapsedTime := hashes * targetHashNanoseconds
		elapsedTime := time.Since(startTime).Nanoseconds()
		if elapsedTime < targetElapsedTime {
			time.Sleep(time.Duration(targetElapsedTime - elapsedTime))
		}

		headerForMining.SetNonce(rand.Uint64())
		pow.CheckProofOfWorkWithTarget(headerForMining, targetDifficulty)
		hashes++
	})

	log.Infof("aaaa %f", float64(hashes)/runDuration.Seconds())
}

func runForDuration(duration time.Duration, runFunction func()) {
	isFinished := false
	go func() {
		for !isFinished {
			runFunction()
		}
	}()
	time.Sleep(duration)
	isFinished = true
}
