package main

import (
	"fmt"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/utils/pow"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"
	"github.com/kaspanet/kaspad/stability-tests/common"
	"github.com/kaspanet/kaspad/util/difficulty"
	"github.com/kaspanet/kaspad/util/panics"
	"math"
	"math/rand"
	"time"
)

const rpcAddress = "localhost:9000"
const miningAddress = "kaspadev:qrcqat6l9zcjsu7swnaztqzrv0s7hu04skpaezxk43y4etj8ncwfkuhy0zmax"

func main() {
	defer panics.HandlePanic(log, "daa-main", nil)
	err := parseConfig()
	if err != nil {
		panic(err)
	}
	defer backendLog.Close()
	common.UseLogger(backendLog, log.Level())

	machineHashNanoseconds := measureMachineHashNanoseconds()
	log.Infof("Machine hashes per second: %d", hashNanosecondsToHashesPerSecond(machineHashNanoseconds))

	testConstantHashRate(machineHashNanoseconds, 5*time.Minute)
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

func testConstantHashRate(machineHashNanoseconds int64, runDuration time.Duration) {
	log.Infof("testConstantHashRate STARTED")
	defer log.Infof("testConstantHashRate FINISHED")

	targetHashNanoseconds := machineHashNanoseconds * 2

	tearDownKaspad := runKaspad()
	defer tearDownKaspad()

	rpcClient, err := rpcclient.NewRPCClient(rpcAddress)
	if err != nil {
		panic(err)
	}

	var miningDurations []time.Duration

	hashes := int64(0)
	startTime := time.Now()
	runForDuration(runDuration, func() {
		getBlockTemplateResponse, err := rpcClient.GetBlockTemplate(miningAddress)
		if err != nil {
			panic(err)
		}
		templateBlock, err := appmessage.RPCBlockToDomainBlock(getBlockTemplateResponse.Block)
		if err != nil {
			panic(err)
		}
		targetDifficulty := difficulty.CompactToBig(templateBlock.Header.Bits())
		headerForMining := templateBlock.Header.ToMutable()

		miningStartTime := time.Now()
		for i := rand.Uint64(); i < math.MaxUint64; i++ {
			targetElapsedTime := hashes * targetHashNanoseconds
			elapsedTime := time.Since(startTime).Nanoseconds()
			if elapsedTime < targetElapsedTime {
				time.Sleep(time.Duration(targetElapsedTime - elapsedTime))
			}
			hashes++

			headerForMining.SetNonce(i)
			if pow.CheckProofOfWorkWithTarget(headerForMining, targetDifficulty) {
				templateBlock.Header = headerForMining.ToImmutable()
				break
			}
		}
		miningDuration := time.Since(miningStartTime)
		miningDurations = append(miningDurations, miningDuration)

		_, err = rpcClient.SubmitBlock(templateBlock)
		if err != nil {
			panic(err)
		}
	})

	log.Infof("Actual hash rate: %f", float64(hashes)/runDuration.Seconds())
	log.Infof("Mined %d blocks", len(miningDurations))

	lastMiningDurationSampleSize := 60
	sumOfLastMiningDurations := time.Duration(0)
	for _, miningDuration := range miningDurations[len(miningDurations)-lastMiningDurationSampleSize:] {
		sumOfLastMiningDurations += miningDuration
	}
	averageOfLastMiningDurations := sumOfLastMiningDurations / time.Duration(lastMiningDurationSampleSize)
	log.Infof("Average: %s", averageOfLastMiningDurations)

	expectedAverageBlocksPerSecond := float64(1)
	deviation := math.Abs(expectedAverageBlocksPerSecond - averageOfLastMiningDurations.Seconds())
	thresholdDeviation := 0.1
	if deviation > thresholdDeviation {
		panic(fmt.Errorf("block rate deviation %f is higher than threshold %f. Want: %f, got: %f",
			deviation, thresholdDeviation, expectedAverageBlocksPerSecond, averageOfLastMiningDurations.Seconds()))
	}
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
