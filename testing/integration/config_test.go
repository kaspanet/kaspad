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

	miningAddress1           = "kaspasim:qqqqnc0pxg7qw3qkc7l6sge8kfhsvvyt7mkw8uamtndqup27ftnd6c769gn66"
	miningAddress1PrivateKey = "0d81045b0deb2af36a25403c2154c87aa82d89dd337b575bae27ce7f5de53cee"

	miningAddress2           = "kaspasim:qqqqnc0pxg7qw3qkc7l6sge8kfhsvvyt7mkw8uamtndqup27ftnd6c769gn66"
	miningAddress2PrivateKey = "0d81045b0deb2af36a25403c2154c87aa82d89dd337b575bae27ce7f5de53cee"

	miningAddress3           = "kaspasim:qqq754f2gdcjcnykwuwwr60c82rh5u6mxxe7yqxljnrxz9fu0h95kduq9ezng"
	miningAddress3PrivateKey = "f6c8f31fd359cbb97007034780bc4021f6ad01c6bc10499b79849efd4cc7ca39"

	defaultTimeout = 10 * time.Second
)

func setConfig(t *testing.T, harness *appHarness) {
	harness.config = commonConfig()
	harness.config.AppDir = randomDirectory(t)
	harness.config.Listeners = []string{harness.p2pAddress}
	harness.config.RPCListeners = []string{harness.rpcAddress}
	harness.config.UTXOIndex = harness.utxoIndex
	harness.config.AllowSubmitBlockWhenNotSynced = true

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
