package main

import (
	"fmt"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"
	"github.com/kaspanet/kaspad/stability-tests/common"
	"github.com/kaspanet/kaspad/util/panics"
	"github.com/kaspanet/kaspad/util/profiling"
	"github.com/pkg/errors"
	"os/exec"
)

func main() {
	defer panics.HandlePanic(log, "mempool-limits-main", nil)
	err := parseConfig()
	if err != nil {
		panic(errors.Wrap(err, "error in parseConfig"))
	}
	defer backendLog.Close()
	common.UseLogger(backendLog, log.Level())

	cfg := activeConfig()
	if cfg.Profile != "" {
		profiling.Start(cfg.Profile, log)
	}

	defer func() {
		if r := recover(); r != nil {
			log.Criticalf("mempool-limits failed")
		}
	}()

	rpcPort := 29587
	kaspadErrChan := runKaspad(rpcPort)
	rpcClient := buildRPCClient(rpcPort)
	fillUpMempool(rpcClient, kaspadErrChan)

	log.Infof("mempool-limits passed")
}

func runKaspad(rpcPort int) chan error {
	cmd := exec.Command("kaspad", "--devnet", fmt.Sprintf("--rpclisten=0.0.0.0:%d", rpcPort))
	cmd.Stdout = common.NewLogWriter(log, logger.LevelTrace, "KASPAD-STDOUT")
	cmd.Stderr = common.NewLogWriter(log, logger.LevelWarn, "KASPAD-STDERR")

	log.Infof("Running `%s`", cmd)
	errChan := make(chan error)
	spawn("kaspad-run", func() {
		errChan <- cmd.Run()
	})
	return errChan
}

func buildRPCClient(rpcPort int) *rpcclient.RPCClient {
	rpcAddress := fmt.Sprintf("127.0.0.1:%d", rpcPort)
	client, err := rpcclient.NewRPCClient(rpcAddress)
	if err != nil {
		panic(errors.Wrapf(err, "error connecting to %s", rpcAddress))
	}
	return client
}

func fillUpMempool(rpcClient *rpcclient.RPCClient, kaspadErrChan chan error) {
	blockDAGInfo, err := rpcClient.GetBlockDAGInfo()
	if err != nil {
		panic(errors.Wrapf(err, "error getting blockDAGInfo"))
	}
	log.Infof("blockDAGInfo: %v", blockDAGInfo)

	select {
	case err := <-kaspadErrChan:
		log.Errorf("Kaspad closed unexpectedly: %s", err)
	}
}
