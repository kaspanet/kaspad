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
		// fdsafdsfsdf
		os.Exit(1)
	}
	backendLog.Close()
}

func realMain() error {
	defer panics.HandlePanic(log, "many-tips-main", nil)

	err := parseConfig()
	if err != nil {
		return errors.Wrap(err, "error in parseConfig")
	}
	common.UseLogger(backendLog, log.Level())
	cfg := activeConfig()
	if cfg.Profile != "" {
		profiling.Start(cfg.Profile, log)
	}

	shutdown := uint64(0)
	teardown, err := startNode()
	if err != nil {
		return errors.Wrap(err, "error in startNode")
	}
	defer teardown()

	miningAddress, err := generateAddress()
	if err != nil {
		return errors.Wrap(err, "error generate a mining address")
	}
	// Mine a chain of 1k blocks
	//chainLength := 1000
	chainLength := 10
	err = mineBlockChains(rpcAddress, miningAddress, chainLength)
	if err != nil {
		return errors.Wrap(err, "Failed mining block chain")
	}

	// Mine on top of the chain 10k tips
	//numOfTips := 10000
	numOfTips := 10
	err = mineTips(numOfTips, miningAddress)
	if err != nil {
		return errors.Wrap(err, "error in mineTips")
	}

	miningAddress2, err := generateAddress()
	if err != nil {
		return errors.Wrap(err, "error generate a mining address")
	}
	err = mineLoop(rpcAddress, miningAddress2)
	if err != nil {
		return errors.Wrap(err, "error in mineLoop")
	}

	log.Infof("finish successfully.")
	atomic.StoreUint64(&shutdown, 1)

	return nil
}

func startNode() (teardown func(), err error) {
	//const connectionListen = "localhost:9001"
	log.Infof("Starting node")
	dataDir, err := common.TempDir("kaspad-datadir")
	if err != nil {
		panic(errors.Wrapf(err, "error in Tempdir"))
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

//todo:tal check if needed tobe public
//no -> remove the comment.
//there is a function like that already dont know how to read for her.
// GenerateAddress generate a new address
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

func mineBlockChains(rpcAddress string, miningAddress util.Address, chainLength int) error {
	for i := 0; i < chainLength; i++ {
		//log.Infof("mine block %d in chain", i)
		kaspaMinerCmd, err := common.StartCmd("MINER",
			"kaspaminer",
			common.NetworkCliArgumentFromNetParams(activeConfig().NetParams()),
			"-s", rpcAddress,
			"--mine-when-not-synced",
			"--miningaddr", miningAddress.EncodeAddress(),
			"--numblocks", "1",
		)
		if err != nil {
			// Ignore error and instead check that the block count changed correctly.
			// TODO: Fix the race condition in kaspaminer so it won't panic (proper shutdown handler)
			log.Warnf("mineBlock returned an err: %s", err)
		}
		//return errors.Wrapf(kaspaMinerCmd.Wait(), "error with command '%s'", kaspaMinerCmd)
		if err = kaspaMinerCmd.Wait(); err != nil {
			return errors.Errorf("error with command '%s' : %s", kaspaMinerCmd, err)
		}
	}
	return nil
}

func mineTips(numOfTips int, miningAddress util.Address) error {

	rpcClient, err := rpc.ConnectToRPC(&rpc.Config{
		RPCServer: rpcAddress,
	}, activeConfig().NetParams())
	if err != nil {
		return errors.Wrap(err, "error connecting to RPC server")
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
		solveBlock(block)
		_, err = rpcClient.SubmitBlock(block)
		if err != nil {
			return err
		}
		log.Infof("im here on mine tips yeah %d", i)
	}
	dagInfo, err := rpcClient.GetBlockDAGInfo()
	if err != nil {
		return err
	}
	log.Infof("There are %d tips in the DAG", len(dagInfo.TipHashes))
	return nil
}

func mineLoop(rpcAddress string, miningAddress util.Address) error {

	kaspaMinerCmd, err := common.StartCmd("MINER",
		"kaspaminer",
		common.NetworkCliArgumentFromNetParams(activeConfig().NetParams()),
		"-s", rpcAddress,
		"--mine-when-not-synced",
		"--miningaddr", miningAddress.EncodeAddress(),
	)
	if err != nil {
		// Ignore error and instead check that the block count changed correctly.
		// TODO: Fix the race condition in kaspaminer so it won't panic (proper shutdown handler)
		log.Warnf("mineBlock returned an err: %s", err)
	}
	shutdown := uint64(0)
	processesStoppedWg := sync.WaitGroup{}
	processesStoppedWg.Add(1)
	//todo:tal to remove all logs prints.
	log.Infof("before spawn")
	spawn("startNode-Cmd.Wait", func() {
		log.Infof("inside first spawn")
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
	spawn("measuringTimeUntilGettingOneTip", func() {
		log.Infof("inside second spawn")
		startTime := time.Now()
		//todo:tal fix string inside

		rpcClientCheckTipsStatus, err := rpc.ConnectToRPC(&rpc.Config{
			RPCServer: rpcAddress,
		}, activeConfig().NetParams())
		if err != nil {
			log.Infof("inside second spawn 010")
			//return errors.Wrap(err, "error connecting to RPC server")
			panic("Failed create new RPC client")
		}
		log.Infof("inside second spawn 01")
		defer rpcClientCheckTipsStatus.Disconnect()
		log.Infof("inside second spawn 1")
		numOfTips, err := getCurrentTipsLength(rpcClientCheckTipsStatus)
		if err != nil {
			panics.Exit(log, fmt.Sprintf("Failed getCurrentTipsLength"))
		}
		log.Infof("inside second spawn 2")
		for numOfTips >= 1 {
			log.Infof("inside second spawn 3")
			time.Sleep(2 * time.Second)
			numOfTips, err = getCurrentTipsLength(rpcClientCheckTipsStatus)
			if err != nil {
				panics.Exit(log, fmt.Sprintf("Failed getCurrentTipsLength"))
			}
		}
		//todo:tal remove all the string in the times print
		duration := time.Since(startTime)
		log.Infof("duration %s", duration)
		if duration >= 2*time.Minute {
			panics.Exit(log, fmt.Sprintf("Took %s to get only one tip after having 10k tips.", duration))
		}
		processesStoppedWg.Done()
	})
	return nil
}

func getCurrentTipsLength(rpcClient *rpc.Client) (int, error) {
	dagInfo, err := rpcClient.GetBlockDAGInfo()
	if err != nil {
		return 0, err
	}
	return len(dagInfo.TipHashes), nil
}

var latestNonce uint64 = 0

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
	panic("Failed to solve block! This should never happen")
}

func killWithSigterm(cmd *exec.Cmd, commandName string) {
	err := cmd.Process.Signal(syscall.SIGTERM)
	if err != nil {
		log.Criticalf("error sending SIGKILL to %s", commandName)
	}
}
