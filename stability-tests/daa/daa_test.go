package daa

import (
	"fmt"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/utils/pow"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"
	"github.com/kaspanet/kaspad/stability-tests/common"
	"github.com/kaspanet/kaspad/util/difficulty"
	"math"
	"math/rand"
	"os"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

const rpcAddress = "localhost:9000"
const miningAddress = "kaspadev:qrcqat6l9zcjsu7swnaztqzrv0s7hu04skpaezxk43y4etj8ncwfkuhy0zmax"
const blockRateDeviationThreshold = 0.5
const averageBlockRateSampleSize = 60

func TestDAA(t *testing.T) {
	machineHashNanoseconds := measureMachineHashNanoseconds(t)
	t.Logf("Machine hashes per second: %d", hashNanosecondsToHashesPerSecond(machineHashNanoseconds))

	tests := []struct {
		name                     string
		runDuration              time.Duration
		throttleDurationFunction func(hashes int64, elapsedTimeNanoseconds int64) int64
	}{
		{
			name:        "constant hash rate",
			runDuration: 5 * time.Minute,
			throttleDurationFunction: func(hashes int64, elapsedTimeNanoseconds int64) int64 {
				targetHashNanoseconds := machineHashNanoseconds * 2
				targetElapsedTimeNanoseconds := hashes * targetHashNanoseconds
				return targetElapsedTimeNanoseconds - elapsedTimeNanoseconds
			},
		},
	}

	for _, test := range tests {
		runDAATest(t, test.name, test.runDuration, test.throttleDurationFunction)
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
	runForDuration(machineHashesPerSecondMeasurementDuration, func() {
		headerForMining.SetNonce(rand.Uint64())
		pow.CheckProofOfWorkWithTarget(headerForMining, targetDifficulty)
		hashes++
	})

	return machineHashesPerSecondMeasurementDuration.Nanoseconds() / hashes
}

func runDAATest(t *testing.T, testName string, runDuration time.Duration,
	throttleDurationFunction func(hashes int64, elapsedTimeNanoseconds int64) int64) {

	t.Logf("TEST STARTED: %s", testName)
	defer t.Logf("TEST FINISHED: %s", testName)

	tearDownKaspad := runKaspad(t)
	defer tearDownKaspad()

	rpcClient, err := rpcclient.NewRPCClient(rpcAddress)
	if err != nil {
		t.Fatalf("NewRPCClient: %s", err)
	}

	var miningDurations []time.Duration

	hashes := int64(0)
	startTime := time.Now()
	runForDuration(runDuration, func() {
		getBlockTemplateResponse, err := rpcClient.GetBlockTemplate(miningAddress)
		if err != nil {
			t.Fatalf("GetBlockTemplate: %s", err)
		}
		templateBlock, err := appmessage.RPCBlockToDomainBlock(getBlockTemplateResponse.Block)
		if err != nil {
			t.Fatalf("RPCBlockToDomainBlock: %s", err)
		}
		targetDifficulty := difficulty.CompactToBig(templateBlock.Header.Bits())
		headerForMining := templateBlock.Header.ToMutable()

		miningStartTime := time.Now()
		for i := rand.Uint64(); i < math.MaxUint64; i++ {
			elapsedTimeNanoseconds := time.Since(startTime).Nanoseconds()
			throttleDuration := throttleDurationFunction(hashes, elapsedTimeNanoseconds)
			time.Sleep(time.Duration(throttleDuration))
			hashes++

			headerForMining.SetNonce(i)
			if pow.CheckProofOfWorkWithTarget(headerForMining, targetDifficulty) {
				templateBlock.Header = headerForMining.ToImmutable()
				break
			}
		}
		miningDuration := time.Since(miningStartTime)
		miningDurations = append(miningDurations, miningDuration)

		averageBlocksPerSecond := calculateAverageBlocksPerSecond(miningDurations)
		t.Logf("Mined block. Took: %s, average blocks per second: %f, time elapsed: %s",
			miningDuration, averageBlocksPerSecond, time.Since(startTime))

		_, err = rpcClient.SubmitBlock(templateBlock)
		if err != nil {
			t.Fatalf("SubmitBlock: %s", err)
		}
	})

	averageBlocksPerSecond := calculateAverageBlocksPerSecond(miningDurations)
	expectedAverageBlocksPerSecond := float64(1)
	deviation := math.Abs(expectedAverageBlocksPerSecond - averageBlocksPerSecond)
	if deviation > blockRateDeviationThreshold {
		t.Fatalf("Block rate deviation %f is higher than threshold %f. Want: %f, got: %f",
			deviation, blockRateDeviationThreshold, expectedAverageBlocksPerSecond, averageBlocksPerSecond)
	}
}

func hashNanosecondsToHashesPerSecond(hashNanoseconds int64) int64 {
	return time.Second.Nanoseconds() / hashNanoseconds
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

func runKaspad(t *testing.T) func() {
	dataDir, err := common.TempDir("kaspad-daa-test")
	if err != nil {
		t.Fatalf("TempDir: %s", err)
	}

	kaspadRunCommand, err := common.StartCmd("KASPAD",
		"kaspad",
		common.NetworkCliArgumentFromNetParams(&dagconfig.DevnetParams),
		"--appdir", dataDir,
		"--logdir", dataDir,
		"--rpclisten", rpcAddress,
		"--loglevel", "debug",
	)
	if err != nil {
		t.Fatalf("StartCmd: %s", err)
	}
	t.Logf("Kaspad started")

	isShutdown := uint64(0)
	go func() {
		err := kaspadRunCommand.Wait()
		if err != nil {
			if atomic.LoadUint64(&isShutdown) == 0 {
				panic(fmt.Sprintf("Kaspad closed unexpectedly: %s. See logs at: %s", err, dataDir))
			}
		}
	}()

	return func() {
		err := kaspadRunCommand.Process.Signal(syscall.SIGTERM)
		if err != nil {
			t.Fatalf("Signal: %s", err)
		}
		err = os.RemoveAll(dataDir)
		if err != nil {
			t.Fatalf("RemoveAll: %s", err)
		}
		atomic.StoreUint64(&isShutdown, 1)
		t.Logf("Kaspad stopped")
	}
}

func calculateAverageBlocksPerSecond(miningDurations []time.Duration) float64 {
	sumOfLastMiningDurations := time.Duration(0)
	startIndex := max(0, len(miningDurations)-averageBlockRateSampleSize)
	for _, miningDuration := range miningDurations[startIndex:] {
		sumOfLastMiningDurations += miningDuration
	}
	averageOfMiningDurations := sumOfLastMiningDurations / time.Duration(averageBlockRateSampleSize)
	return averageOfMiningDurations.Seconds()
}

func max(left int, right int) int {
	if left > right {
		return left
	}
	return right
}
