// Copyright (c) 2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/kaspanet/kaspad/config"
	"github.com/pkg/errors"

	"github.com/jessevdk/go-flags"
	"github.com/kaspanet/kaspad/util"
)

const (
	defaultConfigFilename = "dnsseeder.conf"
	defaultLogFilename    = "dnsseeder.log"
	defaultErrLogFilename = "dnsseeder_err.log"
	defaultListenPort     = "5354"
)

var (
	// Default configuration options
	defaultHomeDir    = util.AppDataDir("dnsseeder", false)
	defaultConfigFile = filepath.Join(defaultHomeDir, defaultConfigFilename)
	defaultLogFile    = filepath.Join(defaultHomeDir, defaultLogFilename)
	defaultErrLogFile = filepath.Join(defaultHomeDir, defaultErrLogFilename)
)

var activeConfig *ConfigFlags

// ActiveConfig returns the active configuration struct
func ActiveConfig() *ConfigFlags {
	return activeConfig
}

// ConfigFlags holds the configurations set by the command line argument
type ConfigFlags struct {
	Host       string `short:"H" long:"host" description:"Seed DNS address"`
	Listen     string `long:"listen" short:"l" description:"Listen on address:port"`
	Nameserver string `short:"n" long:"nameserver" description:"hostname of nameserver"`
	Seeder     string `short:"s" long:"default seeder" description:"IP address of a  working node"`
	config.NetworkFlags
}

func loadConfig() (*ConfigFlags, error) {
	err := os.MkdirAll(defaultHomeDir, 0700)
	if err != nil {
		// Show a nicer error message if it's because a symlink is
		// linked to a directory that does not exist (probably because
		// it's not mounted).
		if e, ok := err.(*os.PathError); ok && os.IsExist(err) {
			if link, lerr := os.Readlink(e.Path); lerr == nil {
				str := "is symlink %s -> %s mounted?"
				err = errors.Errorf(str, e.Path, link)
			}
		}

		str := "failed to create home directory: %v"
		err := errors.Errorf(str, err)
		fmt.Fprintln(os.Stderr, err)
		return nil, err
	}

	// Default config.
	activeConfig = &ConfigFlags{
		Listen: normalizeAddress("localhost", defaultListenPort),
	}

	preCfg := activeConfig
	preParser := flags.NewParser(preCfg, flags.Default)
	_, err = preParser.Parse()
	if err != nil {
		e, ok := err.(*flags.Error)
		if ok && e.Type == flags.ErrHelp {
			os.Exit(0)
		}
		preParser.WriteHelp(os.Stderr)
		return nil, err
	}

	appName := filepath.Base(os.Args[0])
	appName = strings.TrimSuffix(appName, filepath.Ext(appName))
	usageMessage := fmt.Sprintf("Use %s -h to show usage", appName)

	// Load additional config from file.
	parser := flags.NewParser(activeConfig, flags.Default)
	err = flags.NewIniParser(parser).ParseFile(defaultConfigFile)
	if err != nil {
		if _, ok := err.(*os.PathError); !ok {
			fmt.Fprintf(os.Stderr, "Error parsing ConfigFlags "+
				"file: %v\n", err)
			fmt.Fprintln(os.Stderr, usageMessage)
			return nil, err
		}
	}

	// Parse command line options again to ensure they take precedence.
	_, err = parser.Parse()
	if err != nil {
		if e, ok := err.(*flags.Error); !ok || e.Type != flags.ErrHelp {
			parser.WriteHelp(os.Stderr)
		}
		return nil, err
	}

	if len(activeConfig.Host) == 0 {
		str := "Please specify a hostname"
		err := errors.Errorf(str)
		fmt.Fprintln(os.Stderr, err)
		return nil, err
	}

	if len(activeConfig.Nameserver) == 0 {
		str := "Please specify a nameserver"
		err := errors.Errorf(str)
		fmt.Fprintln(os.Stderr, err)
		return nil, err
	}

	activeConfig.Listen = normalizeAddress(activeConfig.Listen, defaultListenPort)

	err = activeConfig.ResolveNetwork(parser)
	if err != nil {
		return nil, err
	}

	initLog(defaultLogFile, defaultErrLogFile)

	return activeConfig, nil
}

// normalizeAddress returns addr with the passed default port appended if
// there is not already a port specified.
func normalizeAddress(addr, defaultPort string) string {
	_, _, err := net.SplitHostPort(addr)
	if err != nil {
		return net.JoinHostPort(addr, defaultPort)
	}
	return addr
}
