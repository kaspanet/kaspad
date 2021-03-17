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

	miningAddress1           = "kaspasim:qr79e37hxdgkn4xjjmfxvqvayc5gsmsql2660d08u9ej9vnc8lzcywr265u64"
	miningAddress1PrivateKey = "0ec5d7308f65717f3f0c3e4d962d73056c1c255a16593b3989589281b51ad5bc"

	miningAddress2           = "kaspasim:qpvr825ypd2fzq779yl83zvte2r4wlgxwra625rgthk9jj96d4cxgsegwryhg"
	miningAddress2PrivateKey = "2a2e99d4a5c3e6d4add69e7baf66b9c7a2f17e74fad86cbd36a3a6815cecc10e"

	miningAddress3           = "kaspasim:qpvr825ypd2fzq779yl83zvte2r4wlgxwra625rgthk9jj96d4cxgsegwryhg"
	miningAddress3PrivateKey = "2a2e99d4a5c3e6d4add69e7baf66b9c7a2f17e74fad86cbd36a3a6815cecc10e"

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
