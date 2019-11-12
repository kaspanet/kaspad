// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/daglabs/btcd/config"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/daglabs/btcd/database"
	_ "github.com/daglabs/btcd/database/ffldb"
	"github.com/daglabs/btcd/util"
	flags "github.com/jessevdk/go-flags"
)

const (
	minCandidates        = 1
	maxCandidates        = 20
	defaultNumCandidates = 5
	defaultDbType        = "ffldb"
)

var (
	btcdHomeDir    = util.AppDataDir("btcd", false)
	defaultDataDir = filepath.Join(btcdHomeDir, "data")
	knownDbTypes   = database.SupportedDrivers()
	activeConfig   *ConfigFlags
)

// ActiveConfig returns the active configuration struct
func ActiveConfig() *ConfigFlags {
	return activeConfig
}

// ConfigFlags defines the configuration options for findcheckpoint.
//
// See loadConfig for details on the configuration load process.
type ConfigFlags struct {
	DataDir       string `short:"b" long:"datadir" description:"Location of the btcd data directory"`
	DbType        string `long:"dbtype" description:"Database backend to use for the Block Chain"`
	NumCandidates int    `short:"n" long:"numcandidates" description:"Max num of checkpoint candidates to show {1-20}"`
	UseGoOutput   bool   `short:"g" long:"gooutput" description:"Display the candidates using Go syntax that is ready to insert into the btcchain checkpoint list"`
	config.NetworkFlags
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
		DataDir:       defaultDataDir,
		DbType:        defaultDbType,
		NumCandidates: defaultNumCandidates,
	}

	// Parse command line options.
	parser := flags.NewParser(&activeConfig, flags.Default)
	remainingArgs, err := parser.Parse()
	if err != nil {
		if e, ok := err.(*flags.Error); !ok || e.Type != flags.ErrHelp {
			parser.WriteHelp(os.Stderr)
		}
		return nil, nil, err
	}

	funcName := "loadConfig"

	err = activeConfig.ResolveNetwork(parser)
	if err != nil {
		return nil, nil, err
	}
	// Validate database type.
	if !validDbType(activeConfig.DbType) {
		str := "%s: The specified database type [%s] is invalid -- " +
			"supported types %s"
		err := errors.Errorf(str, funcName, activeConfig.DbType, strings.Join(knownDbTypes, ", "))
		fmt.Fprintln(os.Stderr, err)
		parser.WriteHelp(os.Stderr)
		return nil, nil, err
	}

	// Append the network type to the data directory so it is "namespaced"
	// per network.  In addition to the block database, there are other
	// pieces of data that are saved to disk such as address manager state.
	// All data is specific to a network, so namespacing the data directory
	// means each individual piece of serialized data does not have to
	// worry about changing names per network and such.
	activeConfig.DataDir = filepath.Join(activeConfig.DataDir, activeConfig.NetParams().Name)

	// Validate the number of candidates.
	if activeConfig.NumCandidates < minCandidates || activeConfig.NumCandidates > maxCandidates {
		str := "%s: The specified number of candidates is out of " +
			"range -- parsed [%d]"
		err = errors.Errorf(str, funcName, activeConfig.NumCandidates)
		fmt.Fprintln(os.Stderr, err)
		parser.WriteHelp(os.Stderr)
		return nil, nil, err
	}

	return activeConfig, remainingArgs, nil
}
