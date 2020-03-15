// Copyright (c) 2015-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"github.com/kaspanet/kaspad/dagconfig"
	_ "github.com/kaspanet/kaspad/database2/drivers/ffldb"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
)

var (
	kaspadHomeDir   = util.AppDataDir("kaspad", false)
	activeNetParams = &dagconfig.MainnetParams

	// Default global config.
	cfg = &config{
		DataDir: filepath.Join(kaspadHomeDir, "data"),
	}
)

// config defines the global configuration options.
type config struct {
	DataDir        string `short:"b" long:"datadir" description:"Location of the kaspad data directory"`
	Testnet        bool   `long:"testnet" description:"Use the test network"`
	RegressionTest bool   `long:"regtest" description:"Use the regression test network"`
	Simnet         bool   `long:"simnet" description:"Use the simulation test network"`
	Devnet         bool   `long:"devnet" description:"Use the development test network"`
}

// fileExists reports whether the named file or directory exists.
func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// setupGlobalConfig examine the global configuration options for any conditions
// which are invalid as well as performs any addition setup necessary after the
// initial parse.
func setupGlobalConfig() error {
	// Multiple networks can't be selected simultaneously.
	// Count number of network flags passed; assign active network params
	// while we're at it
	numNets := 0
	if cfg.Testnet {
		numNets++
		activeNetParams = &dagconfig.TestnetParams
	}
	if cfg.RegressionTest {
		numNets++
		activeNetParams = &dagconfig.RegressionNetParams
	}
	if cfg.Simnet {
		numNets++
		activeNetParams = &dagconfig.SimnetParams
	}
	if cfg.Devnet {
		numNets++
		activeNetParams = &dagconfig.DevnetParams
	}
	if numNets > 1 {
		return errors.New("The testnet, regtest, simnet and devnet params " +
			"can't be used together -- choose one of the four")
	}

	if numNets == 0 {
		return errors.New("Mainnet has not launched yet, use --testnet to run in testnet mode")
	}

	// Append the network type to the data directory so it is "namespaced"
	// per network. In addition to the block database, there are other
	// pieces of data that are saved to disk such as address manager state.
	// All data is specific to a network, so namespacing the data directory
	// means each individual piece of serialized data does not have to
	// worry about changing names per network and such.
	cfg.DataDir = filepath.Join(cfg.DataDir, activeNetParams.Name)

	return nil
}
