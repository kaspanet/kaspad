package main

import (
	"fmt"
	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/utils/mining"
	"github.com/kaspanet/kaspad/util"
	"math/rand"
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
	mining.SolveBlock(block, rand.New(rand.NewSource(time.Now().UnixNano())))
	_, err = rpcClient.SubmitBlock(block)
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
	err = mineTips(numOfTips, miningAddress, rpcClient)
	if err != nil {
		return errors.Wrap(err, "Error in mineTips")
	}
	// Mines until the DAG will have only one tip.
	err = mineLoopUntilHavingOnlyOneTipInDAG(rpcClient, miningAddress)
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
		"--allow-submit-block-when-not-synced",
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
				// TODO: Panic here and check why sometimes kaspad closes ungracefully
				log.Errorf("kaspadCmd closed with an error: %s. See logs at: %s", err, dataDir)
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
	mining.SolveBlock(block, rand.New(rand.NewSource(time.Now().UnixNano())))
	_, err = rpcClient.SubmitBlock(block)
	if err != nil {
		return err
	}
	return nil
}

func mineTips(numOfTips int, miningAddress util.Address, rpcClient *rpc.Client) error {
	blockTemplate, err := rpcClient.GetBlockTemplate(miningAddress.EncodeAddress())
	if err != nil {
		return err
	}
	block, err := appmessage.RPCBlockToDomainBlock(blockTemplate.Block)
	if err != nil {
		return err
	}
	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < numOfTips; i++ {
		mining.SolveBlock(block, rd)
		_, err = rpcClient.SubmitBlock(block)
		if err != nil {
			return err
		}
		if (i%1000 == 0) && (i != 0) {
			log.Infof("Mined %d blocks.", i)
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
func mineLoopUntilHavingOnlyOneTipInDAG(rpcClient *rpc.Client, miningAddress util.Address) error {
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
				// TODO: Panic here and check why sometimes miner closes ungracefully
				log.Errorf("minerCmd closed with an error: %s", err)
			}
		}
	})

	numOfTips, err := getCurrentTipsLength(rpcClient)
	if err != nil {
		return errors.Wrapf(err, "Error in getCurrentTipsLength")
	}
	hasTimedOut := false
	spawn("ChecksIfTimeIsUp", func() {
		timer := time.NewTimer(30 * time.Minute)
		<-timer.C
		hasTimedOut = true
	})
	for numOfTips > 1 && !hasTimedOut {
		time.Sleep(1 * time.Second)
		numOfTips, err = getCurrentTipsLength(rpcClient)
		if err != nil {
			return errors.Wrapf(err, "Error in getCurrentTipsLength")
		}
	}

	if hasTimedOut {
		return errors.Errorf("Out of time - the graph still has more than one tip.")
	}
	duration := time.Since(startMiningTime)
	log.Infof("It took %s until there was only one tip in the DAG after having 10k tips.", duration)
	dagInfo, err = rpcClient.GetBlockDAGInfo()
	if err != nil {
		return errors.Wrapf(err, "Failed in GetBlockDAGInfo")
	}
	numOfAddedBlocks := dagInfo.BlockCount - numOfBlocksBeforeMining
	log.Infof("Added %d blocks to reach this.", numOfAddedBlocks)
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

func killWithSigterm(cmd *exec.Cmd, commandName string) {
	err := cmd.Process.Signal(syscall.SIGTERM)
	if err != nil {
		log.Criticalf("Error sending SIGKILL to %s", commandName)
	}
}
