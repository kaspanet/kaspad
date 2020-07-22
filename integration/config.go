package integration

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/kaspanet/kaspad/dagconfig"

	"github.com/kaspanet/kaspad/config"
)

const (
	kaspad1P2PAddress = "127.0.0.1:54321"
	kaspad2P2PAddress = "127.0.0.1:54322"

	kaspad1RPCAddress = "127.0.0.1:12345"
	kaspad2RPCAddress = "127.0.0.1:12346"

	rpcUser = "user"
	rpcPass = "pass"

	testAddress1   = "kaspasim:qz3tm5pew9lrdpnn8kytgtm6a0mx772j4uw02snetn"
	testAddress1PK = "69f470ff9cd4010de7f4a95161867c49834435423526d9bab83781821cdf95bf"

	testAddress2   = "kaspasim:qqdf0vrh3u576eqzkp0s8qagc04tuj2xnu4sfskhx0"
	testAddress2PK = "aed46ef760223032d2641e086dd48d0b0a4d581811e68ccf15bed2b8fe87348e"

	defaultTimeout = 10 * time.Second
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
	kaspad2Config.RPCListeners = []string{kaspad2RPCAddress}
	kaspad2Config.DisableTLS = true

	return kaspad1Config, kaspad2Config
}

func commonConfig() *config.Config {
	commonConfig := config.DefaultConfig()

	commonConfig.ActiveNetParams = &dagconfig.SimnetParams
	commonConfig.TargetOutboundPeers = 0
	commonConfig.DisableDNSSeed = true
	commonConfig.RPCUser = rpcUser
	commonConfig.RPCPass = rpcPass

	return commonConfig
}

func randomDirectory(t *testing.T) string {
	dir, err := ioutil.TempDir(os.TempDir(), "integration-test-*")
	if err != nil {
		t.Fatalf("Error creating temporary directory for test: %+v", err)
	}

	return dir
}
