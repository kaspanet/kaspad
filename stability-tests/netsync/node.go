package main

import (
	"fmt"
	"github.com/kaspanet/kaspad/domain/consensus"
	"os/exec"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/stability-tests/common"
	"github.com/kaspanet/kaspad/stability-tests/common/mine"
	"github.com/kaspanet/kaspad/stability-tests/common/rpc"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/panics"
	"github.com/pkg/errors"
)

const (
	syncerRPCAddress = "localhost:9000"
	syncedRPCAddress = "localhost:9100"

	syncerListen = "localhost:9001"
	syncedListen = "localhost:9101"
)

func startNode(name string, rpcAddress, listen, connect, profilePort, dataDir string) (*exec.Cmd, func(), error) {
	log.Infof("Data directory for %s is %s", name, dataDir)

	args := []string{
		"kaspad",
		common.NetworkCliArgumentFromNetParams(activeConfig().NetParams()),
		"--appdir", dataDir,
		"--logdir", dataDir,
		"--rpclisten", rpcAddress,
		"--listen", listen,
		"--profile", profilePort,
		"--loglevel", "debug",
		"--allow-submit-block-when-not-synced",
	}
	if connect != "" {
		args = append(args, "--connect", connect)
	}

	if activeConfig().OverrideDAGParamsFile != "" {
		args = append(args, "--override-dag-params-file", activeConfig().OverrideDAGParamsFile)
	}

	cmd, err := common.StartCmd(name,
		args...,
	)
	if err != nil {
		return nil, nil, err
	}

	var shutdown uint32
	stopped := make(chan struct{})
	spawn("startNode-cmd.Wait", func() {
		err := cmd.Wait()
		if err != nil {
			if atomic.LoadUint32(&shutdown) == 0 {
				panics.Exit(log, fmt.Sprintf("%s ( %s ) closed unexpectedly: %s", name, cmd, err))
			}
			if !strings.Contains(err.Error(), "signal: killed") {
				panics.Exit(log, fmt.Sprintf("%s ( %s ) closed with an error: %s", name, cmd, err))
			}
		}
		stopped <- struct{}{}
	})

	return cmd, func() {
		atomic.StoreUint32(&shutdown, 1)
		killWithSigkill(cmd, name)
		const timeout = time.Second
		select {
		case <-stopped:
		case <-time.After(timeout):
			panics.Exit(log, fmt.Sprintf("%s couldn't be closed after %s", name, timeout))
		}
	}, nil
}

func killWithSigkill(cmd *exec.Cmd, commandName string) {
	log.Error("SIGKILLED")
	err := cmd.Process.Signal(syscall.SIGKILL)
	if err != nil {
		log.Criticalf("error sending SIGKILL to %s", commandName)
	}
}

func setupNodeWithRPC(name, listen, rpcListen, connect, profilePort, dataDir string) (*rpc.Client, func(), error) {
	_, teardown, err := startNode(name, rpcListen, listen, connect, profilePort, dataDir)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error in startNode")
	}
	defer func() {
		if r := recover(); r != nil {
			teardown()
			panic(r)
		}
	}()

	log.Infof("Waiting for node %s to start...", name)
	const initTime = 2 * time.Second
	time.Sleep(initTime)

	rpcClient, err := rpc.ConnectToRPC(&rpc.Config{
		RPCServer: rpcListen,
	}, activeConfig().NetParams())
	if err != nil {
		return nil, nil, errors.Wrap(err, "error connecting to JSON-RPC server")
	}

	return rpcClient, teardown, nil
}

func setupSyncee() (*rpc.Client, func(), error) {
	const syncedProfilePort = "6061"

	synceeDataDir, err := useDirOrCreateTemp(activeConfig().SynceeDataDirectory, "syncee-kaspad-data-dir")
	if err != nil {
		return nil, nil, err
	}

	return setupNodeWithRPC("SYNCEE", syncedListen, syncedRPCAddress, syncerListen, syncedProfilePort,
		synceeDataDir)
}

func setupSyncer() (*rpc.Client, func(), error) {
	const syncerProfilePort = "6062"

	syncerDataDir, err := useDirOrCreateTemp(activeConfig().SyncerDataDirectory, "syncer-kaspad-data-dir")
	if err != nil {
		return nil, nil, err
	}

	rpcClient, teardown, err := setupNodeWithRPC("SYNCER", syncerListen, syncerRPCAddress, "",
		syncerProfilePort, syncerDataDir)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		if r := recover(); r != nil {
			teardown()
			panic(r)
		}
	}()

	miningDataDir, err := useDirOrCreateTemp(activeConfig().MiningDataDirectory, "syncer-mining-data-dir")
	if err != nil {
		return nil, nil, err
	}

	err = mine.FromFile(cfg.DAGFile, &consensus.Config{Params: *activeConfig().NetParams()}, rpcClient, miningDataDir)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error in mine.FromFile")
	}

	log.Info("Mining on top of syncer tips")
	rejectReason, err := mineOnTips(rpcClient)
	if err != nil {
		panic(err)
	}
	if rejectReason != appmessage.RejectReasonNone {
		panic(fmt.Sprintf("mined block rejected: %s", rejectReason))
	}

	return rpcClient, teardown, nil
}

func useDirOrCreateTemp(dataDir, tempName string) (string, error) {
	if dataDir != "" {
		return dataDir, nil
	}

	return common.TempDir(tempName)
}

func mineOnTips(client *rpc.Client) (appmessage.RejectReason, error) {
	fakePublicKey := make([]byte, util.PublicKeySize)
	addr, err := util.NewAddressPublicKey(fakePublicKey, activeConfig().NetParams().Prefix)
	if err != nil {
		return appmessage.RejectReasonNone, err
	}

	template, err := client.GetBlockTemplate(addr.String(), "")
	if err != nil {
		return appmessage.RejectReasonNone, err
	}

	domainBlock, err := appmessage.RPCBlockToDomainBlock(template.Block)
	if err != nil {
		return appmessage.RejectReasonNone, err
	}

	if !activeConfig().NetParams().SkipProofOfWork {
		mine.SolveBlock(domainBlock)
	}

	return client.SubmitBlockAlsoIfNonDAA(domainBlock)
}
