package integration

import (
	"os"

	"github.com/kaspanet/kaspad/config"
)

const kaspad1Address = "127.0.0.1:6543"
const kaspad2Address = "127.0.0.1:6544"

func configs() (kaspad1Config, kaspad2Config *config.Config) {
	kaspad1Config = commonConfig()
	kaspad1Config.DataDir = os.TempDir()
	kaspad1Config.Listeners = []string{kaspad1Address}

	kaspad2Config = commonConfig()
	kaspad2Config.DataDir = os.TempDir()
	kaspad2Config.Listeners = []string{kaspad2Address}
	kaspad2Config.ConnectPeers = []string{kaspad1Address}

	return kaspad1Config, kaspad2Config
}

func commonConfig() *config.Config {
	commonConfig := config.DefaultConfig()

	commonConfig.TargetOutboundPeers = 0
	commonConfig.DisableDNSSeed = true

	return commonConfig
}
