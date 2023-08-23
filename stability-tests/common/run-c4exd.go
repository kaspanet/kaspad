package common

import (
	"fmt"
	"os"
	"sync/atomic"
	"syscall"
	"testing"

	"github.com/c4ei/yunseokyeol/domain/dagconfig"
)

// RunKaspadForTesting runs c4exd for testing purposes
func RunKaspadForTesting(t *testing.T, testName string, rpcAddress string) func() {
	appDir, err := TempDir(testName)
	if err != nil {
		t.Fatalf("TempDir: %s", err)
	}

	c4exdRunCommand, err := StartCmd("KASPAD",
		"c4exd",
		NetworkCliArgumentFromNetParams(&dagconfig.DevnetParams),
		"--appdir", appDir,
		"--rpclisten", rpcAddress,
		"--loglevel", "debug",
	)
	if err != nil {
		t.Fatalf("StartCmd: %s", err)
	}
	t.Logf("Kaspad started with --appdir=%s", appDir)

	isShutdown := uint64(0)
	go func() {
		err := c4exdRunCommand.Wait()
		if err != nil {
			if atomic.LoadUint64(&isShutdown) == 0 {
				panic(fmt.Sprintf("Kaspad closed unexpectedly: %s. See logs at: %s", err, appDir))
			}
		}
	}()

	return func() {
		err := c4exdRunCommand.Process.Signal(syscall.SIGTERM)
		if err != nil {
			t.Fatalf("Signal: %s", err)
		}
		err = os.RemoveAll(appDir)
		if err != nil {
			t.Fatalf("RemoveAll: %s", err)
		}
		atomic.StoreUint64(&isShutdown, 1)
		t.Logf("Kaspad stopped")
	}
}
