package integration

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/dagconfig"
)

const (
	p2pAddress1 = "127.0.0.1:54321"
	p2pAddress2 = "127.0.0.1:54322"
	p2pAddress3 = "127.0.0.1:54323"

	rpcAddress1 = "127.0.0.1:12345"
	rpcAddress2 = "127.0.0.1:12346"
	rpcAddress3 = "127.0.0.1:12347"

	rpcUser = "user"
	rpcPass = "pass"

	testAddress1   = "kaspasim:qz3tm5pew9lrdpnn8kytgtm6a0mx772j4uw02snetn"
	testAddress1PK = "69f470ff9cd4010de7f4a95161867c49834435423526d9bab83781821cdf95bf"

	testAddress2   = "kaspasim:qqdf0vrh3u576eqzkp0s8qagc04tuj2xnu4sfskhx0"
	testAddress2PK = "aed46ef760223032d2641e086dd48d0b0a4d581811e68ccf15bed2b8fe87348e"

	testAddress3   = "kaspasim:qq2wz0hl73a0qcl8872wr3djplwmyulurscsqxehu2"
	testAddress3PK = "cc94a79bbccca30b0e3edff1895cbdf8d4ddcc119eacfd692970151dcc2881c2"

	defaultTimeout = 10 * time.Second
)

func setConfig(t *testing.T, harness *appHarness) {
	harness.config = commonConfig()
	harness.config.DataDir = randomDirectory(t)
	harness.config.Listeners = []string{harness.p2pAddress}
	harness.config.RPCListeners = []string{harness.rpcAddress}
}

func commonConfig() *config.Config {
	commonConfig := config.DefaultConfig()

	commonConfig.ActiveNetParams = &dagconfig.SimnetParams
	commonConfig.TargetOutboundPeers = 0
	commonConfig.DisableDNSSeed = true
	commonConfig.RPCUser = rpcUser
	commonConfig.RPCPass = rpcPass
	commonConfig.DisableTLS = true

	return commonConfig
}

func randomDirectory(t *testing.T) string {
	dir, err := ioutil.TempDir("", "integration-test")
	if err != nil {
		t.Fatalf("Error creating temporary directory for test: %+v", err)
	}

	return dir
}
