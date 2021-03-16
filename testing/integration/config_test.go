package integration

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/infrastructure/config"
)

const (
	p2pAddress1 = "127.0.0.1:54321"
	p2pAddress2 = "127.0.0.1:54322"
	p2pAddress3 = "127.0.0.1:54323"

	rpcAddress1 = "127.0.0.1:12345"
	rpcAddress2 = "127.0.0.1:12346"
	rpcAddress3 = "127.0.0.1:12347"

	miningAddress1           = "kaspasim:qzpj2cfa9m40w9m2cmr8pvfuqpp32mzzwsuw6ukhfd"
	miningAddress1PrivateKey = "be9e9884f03e687166479e22d21b064db7903d69b5a46878aae66521c01a6094"

	miningAddress2           = "kaspasim:qr7w7nqsdnc3zddm6u8s9fex4ysk95hm3v30q353ym"
	miningAddress2PrivateKey = "98bd8d8e1f7078abefd017839f83edd0e3c8226ed4989e4d7a8bceb5935de193"

	miningAddress3           = "kaspasim:qz7n8gfak3j2wt9vywy5ljhs3v3xu7lgmutfjqaay5"
	miningAddress3PrivateKey = "eb0af684f2cdbb4ed2d85fbfe0b7f40654a7777fb2c47f142ffb5543b594d1e4"

	defaultTimeout = 10 * time.Second
)

func setConfig(t *testing.T, harness *appHarness) {
	harness.config = commonConfig()
	harness.config.HomeDir = randomDirectory(t)
	harness.config.Listeners = []string{harness.p2pAddress}
	harness.config.RPCListeners = []string{harness.rpcAddress}
	harness.config.UTXOIndex = harness.utxoIndex

	if harness.overrideDAGParams != nil {
		harness.config.ActiveNetParams = harness.overrideDAGParams
	}
}

func commonConfig() *config.Config {
	commonConfig := config.DefaultConfig()

	*commonConfig.ActiveNetParams = dagconfig.SimnetParams // Copy so that we can make changes safely
	commonConfig.ActiveNetParams.BlockCoinbaseMaturity = 10
	commonConfig.TargetOutboundPeers = 0
	commonConfig.DisableDNSSeed = true
	commonConfig.Simnet = true

	return commonConfig
}

func randomDirectory(t *testing.T) string {
	dir, err := ioutil.TempDir("", "integration-test")
	if err != nil {
		t.Fatalf("Error creating temporary directory for test: %+v", err)
	}

	return dir
}
