package main

import (
	"fmt"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/stability-tests/common"
	"os"
	"sync/atomic"
	"syscall"
)

func runKaspad() func() {
	dataDir, err := common.TempDir("kaspad-daa-test")
	if err != nil {
		panic(err)
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
		panic(err)
	}
	log.Infof("Kaspad started")

	isShutdown := uint64(0)
	spawn("runKaspad", func() {
		err := kaspadRunCommand.Wait()
		if err != nil {
			if atomic.LoadUint64(&isShutdown) == 0 {
				panic(fmt.Sprintf("kaspad closed unexpectedly: %s. See logs at: %s", err, dataDir))
			}
		}
	})

	return func() {
		err := kaspadRunCommand.Process.Signal(syscall.SIGTERM)
		if err != nil {
			panic(err)
		}
		err = os.RemoveAll(dataDir)
		if err != nil {
			panic(err)
		}
		atomic.StoreUint64(&isShutdown, 1)
		log.Infof("Kaspad stopped")
	}
}
