package main

import (
	"fmt"
	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/pow"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/difficulty"
	"math"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/kaspanet/kaspad/stability-tests/common"
	"github.com/kaspanet/kaspad/stability-tests/common/rpc"
	"github.com/kaspanet/kaspad/util/panics"
	"github.com/kaspanet/kaspad/util/profiling"
	"github.com/pkg/errors"
)

const rpcAddress = "localhost:9000"

func main() {
	err := realMain()
	if err != nil {
		log.Criticalf("An error occurred: %+v", err)
		backendLog.Close()
		os.Exit(1)
	}
	backendLog.Close()
}

func realMain() error {
	defer panics.HandlePanic(log, "many-tips-main", nil)
	err := parseConfig()
	if err != nil {
		return errors.Wrap(err, "Error in parseConfig")
	}
	common.UseLogger(backendLog, log.Level())
	cfg := activeConfig()
	if cfg.Profile != "" {
		profiling.Start(cfg.Profile, log)
	}
	teardown, err := startNode()
	if err != nil {
		return errors.Wrap(err, "Error in startNode")
	}
	defer teardown()

	miningAddress, err := generateAddress()
	if err != nil {
		return errors.Wrap(err, "Failed generate a mining address")
	}
	rpcClient, err := rpc.ConnectToRPC(&rpc.Config{
		RPCServer: rpcAddress,
	}, activeConfig().NetParams())
	if err != nil {
		return errors.Wrap(err, "Error connecting to RPC server")
	}
	defer rpcClient.Disconnect()

	// Mine block that its timestamp is one second after the genesis timestamp.
	blockTemplate, err := rpcClient.GetBlockTemplate(miningAddress.EncodeAddress())
	if err != nil {
		return err
	}
	block, err := appmessage.RPCBlockToDomainBlock(blockTemplate.Block)
	if err != nil {
		return err
	}
	mutableHeader := block.Header.ToMutable()
	genesisTimestamp := activeConfig().NetParams().GenesisBlock.Header.TimeInMilliseconds()
	mutableHeader.SetTimeInMilliseconds(genesisTimestamp + 1000)
	block.Header = mutableHeader.ToImmutable()
	solvedBlock := solveBlock(block)
	_, err = rpcClient.SubmitBlock(solvedBlock)
	if err != nil {
		return err
	}
	// mine block at the current time
	err = mineBlock(rpcClient, miningAddress)
	if err != nil {
		return errors.Wrap(err, "Error in mineBlock")
	}
	// Mine on top of it 10k tips.
	numOfTips := 10000
	err = mineTips(numOfTips, miningAddress)
	if err != nil {
		return errors.Wrap(err, "Error in mineTips")
	}
	// Mines until the DAG will have only one tip.
	err = mineLoopUntilHavingOnlyOneTipInDAG(rpcAddress, miningAddress)
	if err != nil {
		return errors.Wrap(err, "Error in mineLoop")
	}
	return nil
}

func startNode() (teardown func(), err error) {
	log.Infof("Starting node")
	dataDir, err := common.TempDir("kaspad-datadir")
	if err != nil {
		panic(errors.Wrapf(err, "Error in Tempdir"))
	}
	log.Infof("kaspad datadir: %s", dataDir)

	kaspadCmd, err := common.StartCmd("KASPAD",
		"kaspad",
		common.NetworkCliArgumentFromNetParams(activeConfig().NetParams()),
		"--appdir", dataDir,
		"--logdir", dataDir,
		"--rpclisten", rpcAddress,
		"--loglevel", "debug",
	)
	if err != nil {
		return nil, err
	}
	shutdown := uint64(0)

	processesStoppedWg := sync.WaitGroup{}
	processesStoppedWg.Add(1)
	spawn("startNode-kaspadCmd.Wait", func() {
		err := kaspadCmd.Wait()
		if err != nil {
			if atomic.LoadUint64(&shutdown) == 0 {
				panics.Exit(log, fmt.Sprintf("kaspadCmd closed unexpectedly: %s. See logs at: %s", err, dataDir))
			}
			if !strings.Contains(err.Error(), "signal: killed") {
				panics.Exit(log, fmt.Sprintf("kaspadCmd closed with an error: %s. See logs at: %s", err, dataDir))
			}
		}
		processesStoppedWg.Done()
	})
	return func() {
		log.Infof("defer start-node")
		atomic.StoreUint64(&shutdown, 1)
		killWithSigterm(kaspadCmd, "kaspadCmd")

		processesStoppedChan := make(chan struct{})
		spawn("startNode-processStoppedWg.Wait", func() {
			processesStoppedWg.Wait()
			processesStoppedChan <- struct{}{}
		})

		const timeout = 10 * time.Second
		select {
		case <-processesStoppedChan:
		case <-time.After(timeout):
			panics.Exit(log, fmt.Sprintf("Processes couldn't be closed after %s", timeout))
		}
	}, nil
}

func generateAddress() (util.Address, error) {
	privateKey, err := secp256k1.GenerateSchnorrKeyPair()
	if err != nil {
		return nil, err
	}
	pubKey, err := privateKey.SchnorrPublicKey()
	if err != nil {
		return nil, err
	}
	pubKeySerialized, err := pubKey.Serialize()
	if err != nil {
		return nil, err
	}
	return util.NewAddressPublicKey(pubKeySerialized[:], activeConfig().ActiveNetParams.Prefix)
}

func mineBlock(rpcClient *rpc.Client, miningAddress util.Address) error {
	blockTemplate, err := rpcClient.GetBlockTemplate(miningAddress.EncodeAddress())
	if err != nil {
		return err
	}
	block, err := appmessage.RPCBlockToDomainBlock(blockTemplate.Block)
	if err != nil {
		return err
	}
	solvedBlock := solveBlock(block)
	_, err = rpcClient.SubmitBlock(solvedBlock)
	if err != nil {
		return err
	}
	return nil
}

func mineTips(numOfTips int, miningAddress util.Address) error {
	rpcClient, err := rpc.ConnectToRPC(&rpc.Config{
		RPCServer: rpcAddress,
	}, activeConfig().NetParams())
	if err != nil {
		return errors.Wrap(err, "Error connecting to RPC server")
	}
	defer rpcClient.Disconnect()

	blockTemplate, err := rpcClient.GetBlockTemplate(miningAddress.EncodeAddress())
	if err != nil {
		return err
	}
	block, err := appmessage.RPCBlockToDomainBlock(blockTemplate.Block)
	if err != nil {
		return err
	}
	for i := 0; i < numOfTips; i++ {
		solvedBlock := solveBlock(block)
		_, err = rpcClient.SubmitBlock(solvedBlock)
		if err != nil {
			return err
		}
		if (i%1000 == 0) && (i != 0) {
			log.Infof("Already mined %d blocks.", i)
		}
	}
	dagInfo, err := rpcClient.GetBlockDAGInfo()
	if err != nil {
		return err
	}
	log.Infof("There are %d tips in the DAG", len(dagInfo.TipHashes))
	return nil
}

// Checks how many blocks were mined and how long it took to get only one tip in the DAG (after having 10k tips in the DAG).
func mineLoopUntilHavingOnlyOneTipInDAG(rpcAddress string, miningAddress util.Address) error {
	rpcClient, err := rpc.ConnectToRPC(&rpc.Config{
		RPCServer: rpcAddress,
	}, activeConfig().NetParams())
	if err != nil {
		panic("Failed create new RPC client")
	}
	defer rpcClient.Disconnect()

	dagInfo, err := rpcClient.GetBlockDAGInfo()
	if err != nil {
		return errors.Wrapf(err, "error in GetBlockDAGInfo")
	}
	numOfBlocksBeforeMining := dagInfo.BlockCount

	kaspaMinerCmd, err := common.StartCmd("MINER",
		"kaspaminer",
		common.NetworkCliArgumentFromNetParams(activeConfig().NetParams()),
		"-s", rpcAddress,
		"--mine-when-not-synced",
		"--miningaddr", miningAddress.EncodeAddress(),
		"--target-blocks-per-second=0",
	)
	if err != nil {
		return err
	}
	startMiningTime := time.Now()
	shutdown := uint64(0)

	spawn("kaspa-miner-Cmd.Wait", func() {
		err := kaspaMinerCmd.Wait()
		if err != nil {
			if atomic.LoadUint64(&shutdown) == 0 {
				panics.Exit(log, fmt.Sprintf("minerCmd closed unexpectedly: %s.", err))
			}
			if !strings.Contains(err.Error(), "signal: killed") {
				panics.Exit(log, fmt.Sprintf("minerCmd closed with an error: %s.", err))
			}
		}
	})

	numOfTips, err := getCurrentTipsLength(rpcClient)
	if err != nil {
		return errors.Wrapf(err, "Error in getCurrentTipsLength")
	}
	for numOfTips > 1 {
		time.Sleep(1 * time.Second)
		numOfTips, err = getCurrentTipsLength(rpcClient)
		if err != nil {
			return errors.Wrapf(err, "Error in getCurrentTipsLength")
		}
	}
	duration := time.Since(startMiningTime)
	log.Infof("It took %s until there was only one tip in the DAG after having 10k tips.", duration)
	dagInfo, err = rpcClient.GetBlockDAGInfo()
	if err != nil {
		return errors.Wrapf(err, "Failed in GetBlockDAGInfo")
	}
	numOfAddedBlocks := dagInfo.BlockCount - numOfBlocksBeforeMining
	log.Infof("Added %d blocks to reach this.", numOfAddedBlocks)
	if duration >= 20*time.Minute {
		return errors.Errorf("Error: Took %s to get only one tip.", duration)
	}
	atomic.StoreUint64(&shutdown, 1)
	killWithSigterm(kaspaMinerCmd, "kaspaMinerCmd")
	return nil
}

func getCurrentTipsLength(rpcClient *rpc.Client) (int, error) {
	dagInfo, err := rpcClient.GetBlockDAGInfo()
	if err != nil {
		return 0, err
	}
	log.Infof("Current number of tips is %d", len(dagInfo.TipHashes))
	return len(dagInfo.TipHashes), nil
}

var latestNonce uint64 = 0 // Use to make the nonce unique.
// The nonce is unique for each block in this function.
func solveBlock(block *externalapi.DomainBlock) *externalapi.DomainBlock {
	targetDifficulty := difficulty.CompactToBig(block.Header.Bits())
	headerForMining := block.Header.ToMutable()
	maxUInt64 := uint64(math.MaxUint64)
	for i := latestNonce; i < maxUInt64; i++ {
		headerForMining.SetNonce(i)
		if pow.CheckProofOfWorkWithTarget(headerForMining, targetDifficulty) {
			block.Header = headerForMining.ToImmutable()
			latestNonce = i + 1
			return block
		}
	}
	panic("Failed to solve block!")
}

func killWithSigterm(cmd *exec.Cmd, commandName string) {
	err := cmd.Process.Signal(syscall.SIGTERM)
	if err != nil {
		log.Criticalf("Error sending SIGKILL to %s", commandName)
	}
}
