package daa

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/pow"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"
	"github.com/kaspanet/kaspad/stability-tests/common"
	"github.com/kaspanet/kaspad/util/difficulty"
	"math"
	"math/rand"
	"os"
	"testing"
	"time"
)

const rpcAddress = "localhost:9000"
const miningAddress = "kaspadev:qrcqat6l9zcjsu7swnaztqzrv0s7hu04skpaezxk43y4etj8ncwfkuhy0zmax"
const blockRateDeviationThreshold = 0.5
const averageBlockRateSampleSize = 60
const averageHashRateSampleSize = 100_000

func TestDAA(t *testing.T) {
	if os.Getenv("RUN_STABILITY_TESTS") == "" {
		t.Skip()
	}

	machineHashNanoseconds := measureMachineHashNanoseconds(t)
	t.Logf("Machine hashes per second: %d", hashNanosecondsToHashesPerSecond(machineHashNanoseconds))

	tests := []struct {
		name        string
		runDuration time.Duration

		// targetHashNanosecondsFunction receives the duration of time between now and the start
		// of the run (moments before the first hash has been calculated). It returns the target
		// duration of a single hash operation in nanoseconds (greater return value = lesser hash rate)
		targetHashNanosecondsFunction func(totalElapsedDuration time.Duration) int64
	}{
		{
			name:        "constant hash rate",
			runDuration: 10 * time.Minute,
			targetHashNanosecondsFunction: func(totalElapsedDuration time.Duration) int64 {
				return machineHashNanoseconds * 2
			},
		},
		{
			name:        "sudden hash rate drop",
			runDuration: 15 * time.Minute,
			targetHashNanosecondsFunction: func(totalElapsedDuration time.Duration) int64 {
				if totalElapsedDuration < 5*time.Minute {
					return machineHashNanoseconds * 2
				}
				return machineHashNanoseconds * 10
			},
		},
		{
			name:        "sudden hash rate jump",
			runDuration: 15 * time.Minute,
			targetHashNanosecondsFunction: func(totalElapsedDuration time.Duration) int64 {
				if totalElapsedDuration < 5*time.Minute {
					return machineHashNanoseconds * 10
				}
				return machineHashNanoseconds * 2
			},
		},
		{
			name:        "hash rate peak",
			runDuration: 10 * time.Minute,
			targetHashNanosecondsFunction: func(totalElapsedDuration time.Duration) int64 {
				if totalElapsedDuration > 4*time.Minute && totalElapsedDuration < 5*time.Minute {
					return machineHashNanoseconds * 2
				}
				return machineHashNanoseconds * 10
			},
		},
		{
			name:        "hash rate valley",
			runDuration: 10 * time.Minute,
			targetHashNanosecondsFunction: func(totalElapsedDuration time.Duration) int64 {
				if totalElapsedDuration > 4*time.Minute && totalElapsedDuration < 5*time.Minute {
					return machineHashNanoseconds * 10
				}
				return machineHashNanoseconds * 2
			},
		},
		{
			name:        "periodic hash rate peaks",
			runDuration: 10 * time.Minute,
			targetHashNanosecondsFunction: func(totalElapsedDuration time.Duration) int64 {
				if int(totalElapsedDuration.Seconds())%30 == 0 {
					return machineHashNanoseconds * 2
				}
				return machineHashNanoseconds * 10
			},
		},
		{
			name:        "periodic hash rate valleys",
			runDuration: 10 * time.Minute,
			targetHashNanosecondsFunction: func(totalElapsedDuration time.Duration) int64 {
				if int(totalElapsedDuration.Seconds())%30 == 0 {
					return machineHashNanoseconds * 10
				}
				return machineHashNanoseconds * 2
			},
		},
		{
			name:        "constant exponential hash rate increase",
			runDuration: 15 * time.Minute,
			targetHashNanosecondsFunction: func(totalElapsedDuration time.Duration) int64 {
				fromHashNanoseconds := machineHashNanoseconds * 10
				toHashNanoseconds := machineHashNanoseconds * 2

				if totalElapsedDuration < 10*time.Minute {
					exponentialIncreaseDuration := 10 * time.Minute
					timeElapsedFraction := float64(totalElapsedDuration.Nanoseconds()) / float64(exponentialIncreaseDuration.Nanoseconds())

					return fromHashNanoseconds -
						int64(math.Pow(float64(fromHashNanoseconds-toHashNanoseconds), timeElapsedFraction))
				}

				// 5 minute cooldown

				return toHashNanoseconds
			},
		},
		{
			name:        "constant exponential hash rate decrease",
			runDuration: 15 * time.Minute,
			targetHashNanosecondsFunction: func(totalElapsedDuration time.Duration) int64 {
				fromHashNanoseconds := machineHashNanoseconds * 2
				toHashNanoseconds := machineHashNanoseconds * 10

				if totalElapsedDuration < 10*time.Minute {
					exponentialDecreaseDuration := 10 * time.Minute
					timeElapsedFraction := float64(totalElapsedDuration.Nanoseconds()) / float64(exponentialDecreaseDuration.Nanoseconds())

					return fromHashNanoseconds +
						int64(math.Pow(float64(toHashNanoseconds-fromHashNanoseconds), timeElapsedFraction))
				}

				// 5 minute cooldown

				return toHashNanoseconds
			},
		},
	}

	for _, test := range tests {
		runDAATest(t, test.name, test.runDuration, test.targetHashNanosecondsFunction)
	}
}

func measureMachineHashNanoseconds(t *testing.T) int64 {
	t.Logf("Measuring machine hash rate")
	defer t.Logf("Finished measuring machine hash rate")

	genesisBlock := dagconfig.DevnetParams.GenesisBlock
	targetDifficulty := difficulty.CompactToBig(genesisBlock.Header.Bits())
	headerForMining := genesisBlock.Header.ToMutable()

	machineHashesPerSecondMeasurementDuration := 10 * time.Second
	hashes := int64(0)
	nonce := rand.Uint64()
	loopForDuration(machineHashesPerSecondMeasurementDuration, func(isFinished *bool) {
		headerForMining.SetNonce(nonce)
		pow.CheckProofOfWorkWithTarget(headerForMining, targetDifficulty)
		hashes++
		nonce++
	})

	return machineHashesPerSecondMeasurementDuration.Nanoseconds() / hashes
}

func runDAATest(t *testing.T, testName string, runDuration time.Duration,
	targetHashNanosecondsFunction func(totalElapsedDuration time.Duration) int64) {

	t.Logf("DAA TEST STARTED: %s", testName)
	defer t.Logf("DAA TEST FINISHED: %s", testName)

	tearDownKaspad := common.RunKaspadForTesting(t, "kaspad-daa-test", rpcAddress)
	defer tearDownKaspad()

	rpcClient, err := rpcclient.NewRPCClient(rpcAddress)
	if err != nil {
		t.Fatalf("NewRPCClient: %s", err)
	}

	hashDurations := make([]time.Duration, 0, averageHashRateSampleSize+1)
	miningDurations := make([]time.Duration, 0, averageBlockRateSampleSize+1)
	previousDifficulty := float64(0)
	blocksMined := 0

	startTime := time.Now()
	loopForDuration(runDuration, func(isFinished *bool) {
		templateBlock := fetchBlockForMining(t, rpcClient)
		targetDifficulty := difficulty.CompactToBig(templateBlock.Header.Bits())
		headerForMining := templateBlock.Header.ToMutable()

		miningStartTime := time.Now()
		nonce := rand.Uint64()
		for {
			hashStartTime := time.Now()

			headerForMining.SetNonce(nonce)
			if pow.CheckProofOfWorkWithTarget(headerForMining, targetDifficulty) {
				templateBlock.Header = headerForMining.ToImmutable()
				break
			}
			nonce++

			waitUntilTargetHashDurationHadElapsed(startTime, hashStartTime, targetHashNanosecondsFunction)

			hashDuration := time.Since(hashStartTime)
			hashDurations = append(hashDurations, hashDuration)
			if len(hashDurations) > averageHashRateSampleSize {
				hashDurations = hashDurations[1:]
			}

			if *isFinished {
				return
			}
		}
		miningDuration := time.Since(miningStartTime)
		miningDurations = append(miningDurations, miningDuration)
		if len(miningDurations) > averageBlockRateSampleSize {
			miningDurations = miningDurations[1:]
		}

		if *isFinished {
			return
		}

		averageMiningDuration := calculateAverageDuration(miningDurations)
		averageHashNanoseconds := calculateAverageDuration(hashDurations).Nanoseconds()
		averageHashesPerSecond := hashNanosecondsToHashesPerSecond(averageHashNanoseconds)
		blockDAGInfoResponse, err := rpcClient.GetBlockDAGInfo()
		if err != nil {
			t.Fatalf("GetBlockDAGInfo: %s", err)
		}
		difficultyDelta := blockDAGInfoResponse.Difficulty - previousDifficulty
		previousDifficulty = blockDAGInfoResponse.Difficulty
		blocksMined++
		t.Logf("Mined block. Took: %s, average block mining duration: %s, "+
			"average hashes per second: %d, difficulty delta: %f, time elapsed: %s, blocks mined: %d",
			miningDuration, averageMiningDuration, averageHashesPerSecond, difficultyDelta, time.Since(startTime), blocksMined)

		if *isFinished {
			return
		}

		_, err = rpcClient.SubmitBlock(templateBlock)
		if err != nil {
			t.Fatalf("SubmitBlock: %s", err)
		}
	})

	averageBlocksPerSecond := calculateAverageDuration(miningDurations).Seconds()
	expectedAverageBlocksPerSecond := float64(1)
	deviation := math.Abs(expectedAverageBlocksPerSecond - averageBlocksPerSecond)
	if deviation > blockRateDeviationThreshold {
		t.Errorf("Block rate deviation %f is higher than threshold %f. Want: %f, got: %f",
			deviation, blockRateDeviationThreshold, expectedAverageBlocksPerSecond, averageBlocksPerSecond)
	}
}

func fetchBlockForMining(t *testing.T, rpcClient *rpcclient.RPCClient) *externalapi.DomainBlock {
	getBlockTemplateResponse, err := rpcClient.GetBlockTemplate(miningAddress)
	if err != nil {
		t.Fatalf("GetBlockTemplate: %s", err)
	}
	templateBlock, err := appmessage.RPCBlockToDomainBlock(getBlockTemplateResponse.Block)
	if err != nil {
		t.Fatalf("RPCBlockToDomainBlock: %s", err)
	}
	return templateBlock
}

func waitUntilTargetHashDurationHadElapsed(startTime time.Time, hashStartTime time.Time,
	targetHashNanosecondsFunction func(totalElapsedDuration time.Duration) int64) {

	// Yielding a thread in Go takes up to a few milliseconds whereas hashing once
	// takes a few hundred nanoseconds, so we spin in place instead of e.g. calling time.Sleep()
	for {
		targetHashNanoseconds := targetHashNanosecondsFunction(time.Since(startTime))
		hashElapsedDurationNanoseconds := time.Since(hashStartTime).Nanoseconds()
		if hashElapsedDurationNanoseconds >= targetHashNanoseconds {
			break
		}
	}
}

func hashNanosecondsToHashesPerSecond(hashNanoseconds int64) int64 {
	return time.Second.Nanoseconds() / hashNanoseconds
}

func loopForDuration(duration time.Duration, runFunction func(isFinished *bool)) {
	isFinished := false
	go func() {
		for !isFinished {
			runFunction(&isFinished)
		}
	}()
	time.Sleep(duration)
	isFinished = true
}

func calculateAverageDuration(durations []time.Duration) time.Duration {
	sumOfDurations := time.Duration(0)
	for _, duration := range durations {
		sumOfDurations += duration
	}
	averageOfDurations := sumOfDurations / time.Duration(len(durations))
	return averageOfDurations
}
