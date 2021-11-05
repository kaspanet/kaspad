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
	"math/big"
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
		// duration of a single hash operation in nanoseconds (greater return value = lower hash rate)
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

				// 5 minute cooldown. We expect the DAA to still be "catching up" at the end
				// of the exponential increase so, for the sake of testing, we wait a while for
				// the hash rate to stabilize

				return toHashNanoseconds
			},
		},
		{
			name:        "constant exponential hash rate decrease",
			runDuration: 45 * time.Minute,
			targetHashNanosecondsFunction: func(totalElapsedDuration time.Duration) int64 {
				fromHashNanoseconds := machineHashNanoseconds * 2
				toHashNanoseconds := machineHashNanoseconds * 10

				if totalElapsedDuration < 10*time.Minute {
					exponentialDecreaseDuration := 10 * time.Minute
					timeElapsedFraction := float64(totalElapsedDuration.Nanoseconds()) / float64(exponentialDecreaseDuration.Nanoseconds())

					return fromHashNanoseconds +
						int64(math.Pow(float64(toHashNanoseconds-fromHashNanoseconds), timeElapsedFraction))
				}

				// 5 minute cooldown. We expect the DAA to still be "catching up" at the end
				// of the exponential decrease so, for the sake of testing, we wait a while for
				// the hash rate to stabilize

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

	// These variables are for gathering stats. Useful mostly for debugging
	averageHashDuration := newAverageDuration(averageHashRateSampleSize)
	averageMiningDuration := newAverageDuration(averageBlockRateSampleSize)
	previousDifficulty := float64(0)
	blocksMined := 0

	// Mine blocks the same way a CPU miner mines blocks until `runDuration` elapses
	startTime := time.Now()
	loopForDuration(runDuration, func(isFinished *bool) {
		templateBlock := fetchBlockForMining(t, rpcClient)
		targetDifficulty := difficulty.CompactToBig(templateBlock.Header.Bits())
		headerForMining := templateBlock.Header.ToMutable()

		// Try hashes until we find a valid block
		miningStartTime := time.Now()
		nonce := rand.Uint64()
		for {
			hashStartTime := time.Now()

			blockFound := tryNonceForMiningAndIncrementNonce(headerForMining, &nonce, targetDifficulty, templateBlock)
			if blockFound {
				break
			}

			// Throttle the hash rate by waiting until the target hash duration elapses
			waitUntilTargetHashDurationHadElapsed(startTime, hashStartTime, targetHashNanosecondsFunction)

			// Collect stats about hash rate
			hashDuration := time.Since(hashStartTime)
			averageHashDuration.add(hashDuration)

			// Exit early if the test is finished
			if *isFinished {
				return
			}
		}

		// Collect stats about block rate
		miningDuration := time.Since(miningStartTime)
		averageMiningDuration.add(miningDuration)

		logMinedBlockStatsAndUpdateStatFields(t, rpcClient, averageMiningDuration, averageHashDuration, startTime,
			miningDuration, &previousDifficulty, &blocksMined)

		// Exit early if the test is finished
		if *isFinished {
			return
		}

		submitMinedBlock(t, rpcClient, templateBlock)
	})

	averageMiningDurationInSeconds := averageMiningDuration.toDuration().Seconds()
	expectedAverageMiningDurationInSeconds := float64(1)
	deviation := math.Abs(expectedAverageMiningDurationInSeconds - averageMiningDurationInSeconds)
	if deviation > blockRateDeviationThreshold {
		t.Errorf("Block rate deviation %f is higher than threshold %f. Want: %f, got: %f",
			deviation, blockRateDeviationThreshold, expectedAverageMiningDurationInSeconds, averageMiningDurationInSeconds)
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

func tryNonceForMiningAndIncrementNonce(headerForMining externalapi.MutableBlockHeader, nonce *uint64,
	targetDifficulty *big.Int, templateBlock *externalapi.DomainBlock) bool {

	headerForMining.SetNonce(*nonce)
	if pow.CheckProofOfWorkWithTarget(headerForMining, targetDifficulty) {
		templateBlock.Header = headerForMining.ToImmutable()
		return true
	}

	*nonce++
	return false
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

func logMinedBlockStatsAndUpdateStatFields(t *testing.T, rpcClient *rpcclient.RPCClient,
	averageMiningDuration *averageDuration, averageHashDurations *averageDuration,
	startTime time.Time, miningDuration time.Duration, previousDifficulty *float64, blocksMined *int) {

	averageMiningDurationAsDuration := averageMiningDuration.toDuration()
	averageHashNanoseconds := averageHashDurations.toDuration().Nanoseconds()
	averageHashesPerSecond := hashNanosecondsToHashesPerSecond(averageHashNanoseconds)
	blockDAGInfoResponse, err := rpcClient.GetBlockDAGInfo()
	if err != nil {
		t.Fatalf("GetBlockDAGInfo: %s", err)
	}
	difficultyDelta := blockDAGInfoResponse.Difficulty - *previousDifficulty
	*previousDifficulty = blockDAGInfoResponse.Difficulty
	*blocksMined++
	t.Logf("Mined block. Took: %s, average block mining duration: %s, "+
		"average hashes per second: %d, difficulty delta: %f, time elapsed: %s, blocks mined: %d",
		miningDuration, averageMiningDurationAsDuration, averageHashesPerSecond, difficultyDelta, time.Since(startTime), *blocksMined)
}

func submitMinedBlock(t *testing.T, rpcClient *rpcclient.RPCClient, block *externalapi.DomainBlock) {
	_, err := rpcClient.SubmitBlock(block)
	if err != nil {
		t.Fatalf("SubmitBlock: %s", err)
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
