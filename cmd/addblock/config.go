// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	flags "github.com/jessevdk/go-flags"
	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
)

const (
	defaultDataFile = "bootstrap.dat"
	defaultProgress = 10
)

var (
	kaspadHomeDir  = util.AppDataDir("kaspad", false)
	defaultDataDir = filepath.Join(kaspadHomeDir, "data")
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
	DataDir         string `short:"b" long:"datadir" description:"Location of the kaspad data directory"`
	InFile          string `short:"i" long:"infile" description:"File containing the block(s)"`
	Progress        int    `short:"p" long:"progress" description:"Show a progress message each time this number of seconds have passed -- Use 0 to disable progress announcements"`
	AcceptanceIndex bool   `long:"acceptanceindex" description:"Maintain a full hash-based acceptance index which makes the getChainFromBlock RPC available"`
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

// loadConfig initializes and parses the config using command line options.
func loadConfig() (*ConfigFlags, []string, error) {
	// Default config.
	activeConfig = &ConfigFlags{
		DataDir:  defaultDataDir,
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
