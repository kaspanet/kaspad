package integration

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/kaspanet/kaspad/config"
)

const (
	kaspad1P2PAddress = "127.0.0.1:54321"
	kaspad2P2PAddress = "127.0.0.1:54322"

	kaspad1RPCAddress = "127.0.0.1:12345"
	kaspad2RPCAddress = "127.0.0.1:12346"
)

func configs(t *testing.T) (kaspad1Config, kaspad2Config *config.Config) {
	kaspad1Config = commonConfig()
	kaspad1Config.DataDir = randomDirectory(t)
	kaspad1Config.Listeners = []string{kaspad1P2PAddress}
	kaspad1Config.RPCListeners = []string{kaspad1RPCAddress}
	kaspad1Config.DisableTLS = true

	kaspad2Config = commonConfig()
	kaspad2Config.DataDir = randomDirectory(t)
	kaspad2Config.Listeners = []string{kaspad2P2PAddress}
	kaspad2Config.ConnectPeers = []string{kaspad1P2PAddress}
	kaspad2Config.RPCListeners = []string{kaspad2RPCAddress}
	kaspad2Config.DisableTLS = true

	return kaspad1Config, kaspad2Config
}

func commonConfig() *config.Config {
	commonConfig := config.DefaultConfig()

	commonConfig.TargetOutboundPeers = 0
	commonConfig.DisableDNSSeed = true

	return commonConfig
}

func randomDirectory(t *testing.T) string {
	dir, err := ioutil.TempDir(os.TempDir(), "integration-test-*")
	if err != nil {
		t.Fatalf("Error creating temporary directory for test: %+v", err)
	}

	return dir
}
