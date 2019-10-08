// Copyright (c) 2015-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/database"
	_ "github.com/daglabs/btcd/database/ffldb"
	"github.com/daglabs/btcd/util"
)

var (
	btcdHomeDir     = util.AppDataDir("btcd", false)
	knownDbTypes    = database.SupportedDrivers()
	activeNetParams = &dagconfig.MainNetParams

	// Default global config.
	cfg = &config{
		DataDir: filepath.Join(btcdHomeDir, "data"),
		DbType:  "ffldb",
	}
)

// config defines the global configuration options.
type config struct {
	DataDir        string `short:"b" long:"datadir" description:"Location of the btcd data directory"`
	DbType         string `long:"dbtype" description:"Database backend to use for the Block Chain"`
	TestNet        bool   `long:"testnet" description:"Use the test network"`
	RegressionTest bool   `long:"regtest" description:"Use the regression test network"`
	SimNet         bool   `long:"simnet" description:"Use the simulation test network"`
	DevNet         bool   `long:"devnet" description:"Use the development test network"`
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

// validDbType returns whether or not dbType is a supported database type.
func validDbType(dbType string) bool {
	for _, knownType := range knownDbTypes {
		if dbType == knownType {
			return true
		}
	}

	return false
}

// setupGlobalConfig examine the global configuration options for any conditions
// which are invalid as well as performs any addition setup necessary after the
// initial parse.
func setupGlobalConfig() error {
	// Multiple networks can't be selected simultaneously.
	// Count number of network flags passed; assign active network params
	// while we're at it
	numNets := 0
	if cfg.TestNet {
		numNets++
		activeNetParams = &dagconfig.TestNetParams
	}
	if cfg.RegressionTest {
		numNets++
		activeNetParams = &dagconfig.RegressionNetParams
	}
	if cfg.SimNet {
		numNets++
		activeNetParams = &dagconfig.SimNetParams
	}
	if cfg.DevNet {
		numNets++
		activeNetParams = &dagconfig.DevNetParams
	}
	if numNets > 1 {
		return errors.New("The testnet, regtest, simnet and devnet params " +
			"can't be used together -- choose one of the four")
	}

	// Validate database type.
	if !validDbType(cfg.DbType) {
		str := "The specified database type [%s] is invalid -- " +
			"supported types: %s"
		return fmt.Errorf(str, cfg.DbType, strings.Join(knownDbTypes, ", "))
	}

	// Append the network type to the data directory so it is "namespaced"
	// per network.  In addition to the block database, there are other
	// pieces of data that are saved to disk such as address manager state.
	// All data is specific to a network, so namespacing the data directory
	// means each individual piece of serialized data does not have to
	// worry about changing names per network and such.
	cfg.DataDir = filepath.Join(cfg.DataDir, activeNetParams.Name)

	return nil
}
