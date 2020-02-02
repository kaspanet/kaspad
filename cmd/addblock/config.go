// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/kaspanet/kaspad/config"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strings"

	flags "github.com/jessevdk/go-flags"
	"github.com/kaspanet/kaspad/database"
	_ "github.com/kaspanet/kaspad/database/ffldb"
	"github.com/kaspanet/kaspad/util"
)

const (
	defaultDbType   = "ffldb"
	defaultDataFile = "bootstrap.dat"
	defaultProgress = 10
)

var (
	kaspadHomeDir  = util.AppDataDir("kaspad", false)
	defaultDataDir = filepath.Join(kaspadHomeDir, "data")
	knownDbTypes   = database.SupportedDrivers()
	activeConfig   *ConfigFlags
)

// ActiveConfig returns the active configuration struct
func ActiveConfig() *ConfigFlags {
	return activeConfig
}

// ConfigFlags defines the configuration options for addblock.
//
// See loadConfig for details on the configuration load process.
type ConfigFlags struct {
	DataDir   string `short:"b" long:"datadir" description:"Location of the kaspad data directory"`
	DbType    string `long:"dbtype" description:"Database backend to use for the Block DAG"`
	InFile    string `short:"i" long:"infile" description:"File containing the block(s)"`
	TxIndex   bool   `long:"txindex" description:"Build a full hash-based transaction index which makes all transactions available via the getrawtransaction RPC"`
	AddrIndex bool   `long:"addrindex" description:"Build a full address-based transaction index which makes the searchrawtransactions RPC available"`
	Progress  int    `short:"p" long:"progress" description:"Show a progress message each time this number of seconds have passed -- Use 0 to disable progress announcements"`
	config.NetworkFlags
}

// filesExists reports whether the named file or directory exists.
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

// loadConfig initializes and parses the config using command line options.
func loadConfig() (*ConfigFlags, []string, error) {
	// Default config.
	activeConfig = &ConfigFlags{
		DataDir:  defaultDataDir,
		DbType:   defaultDbType,
		InFile:   defaultDataFile,
		Progress: defaultProgress,
	}

	// Parse command line options.
	parser := flags.NewParser(&activeConfig, flags.Default)
	remainingArgs, err := parser.Parse()
	if err != nil {
		var flagsErr *flags.Error
		if ok := errors.As(err, &flagsErr); !ok || flagsErr.Type != flags.ErrHelp {
			parser.WriteHelp(os.Stderr)
		}
		return nil, nil, err
	}

	err = activeConfig.ResolveNetwork(parser)
	if err != nil {
		return nil, nil, err
	}

	// Validate database type.
	if !validDbType(activeConfig.DbType) {
		str := "%s: The specified database type [%s] is invalid -- " +
			"supported types %s"
		err := errors.Errorf(str, "loadConfig", activeConfig.DbType, strings.Join(knownDbTypes, ", "))
		fmt.Fprintln(os.Stderr, err)
		parser.WriteHelp(os.Stderr)
		return nil, nil, err
	}

	// Append the network type to the data directory so it is "namespaced"
	// per network. In addition to the block database, there are other
	// pieces of data that are saved to disk such as address manager state.
	// All data is specific to a network, so namespacing the data directory
	// means each individual piece of serialized data does not have to
	// worry about changing names per network and such.
	cfg.DataDir = filepath.Join(cfg.DataDir, ActiveConfig().NetParams().Name)

	// Ensure the specified block file exists.
	if !fileExists(cfg.InFile) {
		str := "%s: The specified block file [%s] does not exist"
		err := errors.Errorf(str, "loadConfig", cfg.InFile)
		fmt.Fprintln(os.Stderr, err)
		parser.WriteHelp(os.Stderr)
		return nil, nil, err
	}

	return cfg, remainingArgs, nil
}
