// Copyright (c) 2013-2015 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/version"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/kaspanet/kaspad/network/rpc/model"
	"github.com/kaspanet/kaspad/util"
)

const (
	// unusableFlags are the command usage flags which this utility are not
	// able to use. In particular it doesn't support websockets and
	// consequently notifications.
	unusableFlags = model.UFWebsocketOnly | model.UFNotification
)

var (
	kaspadHomeDir      = util.AppDataDir("kaspad", false)
	kaspactlHomeDir    = util.AppDataDir("kaspactl", false)
	defaultConfigFile  = filepath.Join(kaspactlHomeDir, "kaspactl.conf")
	defaultRPCServer   = "localhost"
	defaultRPCCertFile = filepath.Join(kaspadHomeDir, "rpc.cert")
	activeConfig       *ConfigFlags
)

// listCommands categorizes and lists all of the usable commands along with
// their one-line usage.
func listCommands() {
	const (
		categoryDAG uint8 = iota
		numCategories
	)

	// Get a list of registered commands and categorize and filter them.
	cmdMethods := model.RegisteredCmdMethods()
	categorized := make([][]string, numCategories)
	for _, method := range cmdMethods {
		flags, err := model.MethodUsageFlags(method)
		if err != nil {
			// This should never happen since the method was just
			// returned from the package, but be safe.
			continue
		}

		// Skip the commands that aren't usable from this utility.
		if flags&unusableFlags != 0 {
			continue
		}

		usage, err := model.MethodUsageText(method)
		if err != nil {
			// This should never happen since the method was just
			// returned from the package, but be safe.
			continue
		}

		// Categorize the command based on the usage flags.
		category := categoryDAG
		categorized[category] = append(categorized[category], usage)
	}

	// Display the command according to their categories.
	categoryTitles := make([]string, numCategories)
	categoryTitles[categoryDAG] = "DAG Server Commands:"
	for category := uint8(0); category < numCategories; category++ {
		fmt.Println(categoryTitles[category])
		for _, usage := range categorized[category] {
			fmt.Println(usage)
		}
		fmt.Println()
	}
}

// ConfigFlags defines the configuration options for kaspactl.
//
// See loadConfig for details on the configuration load process.
type ConfigFlags struct {
	ShowVersion   bool   `short:"V" long:"version" description:"Display version information and exit"`
	ListCommands  bool   `short:"l" long:"listcommands" description:"List all of the supported commands and exit"`
	ConfigFile    string `short:"C" long:"configfile" description:"Path to configuration file"`
	RPCUser       string `short:"u" long:"rpcuser" description:"RPC username"`
	RPCPassword   string `short:"P" long:"rpcpass" default-mask:"-" description:"RPC password"`
	RPCServer     string `short:"s" long:"rpcserver" description:"RPC server to connect to"`
	RPCCert       string `short:"c" long:"rpccert" description:"RPC server certificate chain for validation"`
	NoTLS         bool   `long:"notls" description:"Disable TLS"`
	Proxy         string `long:"proxy" description:"Connect via SOCKS5 proxy (eg. 127.0.0.1:9050)"`
	ProxyUser     string `long:"proxyuser" description:"Username for proxy server"`
	ProxyPass     string `long:"proxypass" default-mask:"-" description:"Password for proxy server"`
	TLSSkipVerify bool   `long:"skipverify" description:"Do not verify tls certificates (not recommended!)"`
	config.NetworkFlags
}

// cleanAndExpandPath expands environement variables and leading ~ in the
// passed path, cleans the result, and returns it.
func cleanAndExpandPath(path string) string {
	// Expand initial ~ to OS specific home directory.
	if strings.HasPrefix(path, "~") {
		homeDir := filepath.Dir(kaspactlHomeDir)
		path = strings.Replace(path, "~", homeDir, 1)
	}

	// NOTE: The os.ExpandEnv doesn't work with Windows-style %VARIABLE%,
	// but they variables can still be expanded via POSIX-style $VARIABLE.
	return filepath.Clean(os.ExpandEnv(path))
}

// loadConfig initializes and parses the config using a config file and command
// line options.
//
// The configuration proceeds as follows:
// 	1) Start with a default config with sane settings
// 	2) Pre-parse the command line to check for an alternative config file
// 	3) Load configuration file overwriting defaults with any specified options
// 	4) Parse CLI options and overwrite/add any specified options
//
// The above results in functioning properly without any config settings
// while still allowing the user to override settings with config files and
// command line options. Command line options always take precedence.
func loadConfig() (*ConfigFlags, []string, error) {
	// Default config.
	activeConfig = &ConfigFlags{
		ConfigFile: defaultConfigFile,
		RPCServer:  defaultRPCServer,
		RPCCert:    defaultRPCCertFile,
	}

	// Pre-parse the command line options to see if an alternative config
	// file, the version flag, or the list commands flag was specified. Any
	// errors aside from the help message error can be ignored here since
	// they will be caught by the final parse below.
	preCfg := activeConfig
	preParser := flags.NewParser(preCfg, flags.HelpFlag)
	_, err := preParser.Parse()
	if err != nil {
		var flagsErr *flags.Error
		if ok := errors.As(err, &flagsErr); ok && flagsErr.Type == flags.ErrHelp {
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, "The special parameter `-` "+
				"indicates that a parameter should be read "+
				"from the\nnext unread line from standard "+
				"input.")
			return nil, nil, err
		}
	}

	// Show the version and exit if the version flag was specified.
	appName := filepath.Base(os.Args[0])
	appName = strings.TrimSuffix(appName, filepath.Ext(appName))
	usageMessage := fmt.Sprintf("Use %s -h to show options", appName)
	if preCfg.ShowVersion {
		fmt.Println(appName, "version", version.Version())
		os.Exit(0)
	}

	// Show the available commands and exit if the associated flag was
	// specified.
	if preCfg.ListCommands {
		listCommands()
		os.Exit(0)
	}

	// If no rpc user and password were configured, create
	// a kaspactl default config file based on the rpc login
	// details written in the RPC server configuration file
	if preCfg.RPCUser == "" && preCfg.RPCPassword == "" {
		if _, err := os.Stat(preCfg.ConfigFile); os.IsNotExist(err) {
			serverConfigPath := filepath.Join(kaspadHomeDir, "kaspad.conf")
			err := createDefaultConfigFile(preCfg.ConfigFile, serverConfigPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating a default config file: %s\n", err)
			}
		}
	}

	// Load additional config from file.
	parser := flags.NewParser(activeConfig, flags.Default)
	err = flags.NewIniParser(parser).ParseFile(preCfg.ConfigFile)
	if err != nil {
		if pErr := &(os.PathError{}); !errors.As(err, &pErr) {
			fmt.Fprintf(os.Stderr, "Error parsing config file: %s\n",
				err)
			fmt.Fprintln(os.Stderr, usageMessage)
			return nil, nil, err
		}
	}

	// Parse command line options again to ensure they take precedence.
	remainingArgs, err := parser.Parse()
	if err != nil {
		var flagsErr *flags.Error
		if ok := errors.As(err, &flagsErr); !ok || flagsErr.Type != flags.ErrHelp {
			fmt.Fprintln(os.Stderr, usageMessage)
		}
		return nil, nil, err
	}

	err = activeConfig.ResolveNetwork(parser)
	if err != nil {
		return nil, nil, err
	}
	// Handle environment variable expansion in the RPC certificate path.
	activeConfig.RPCCert = cleanAndExpandPath(activeConfig.RPCCert)

	// Add default port to RPC server based on --testnet and --simnet flags
	// if needed.
	activeConfig.RPCServer, err = activeConfig.NetParams().NormalizeRPCServerAddress(activeConfig.RPCServer)
	if err != nil {
		return nil, nil, err
	}

	return activeConfig, remainingArgs, nil
}

// createDefaultConfig creates a basic config file at the given destination path.
// For this it tries to read the config file for the RPC server and extract the
// RPC user and password from it.
func createDefaultConfigFile(destinationPath, serverConfigPath string) error {
	// Read the RPC server config
	serverConfigFile, err := os.Open(serverConfigPath)
	if os.IsNotExist(err) {
		return errors.Errorf("the RPC server configuration file could not be found at %s", serverConfigPath)
	}
	if err != nil {
		return err
	}
	defer serverConfigFile.Close()
	content, err := ioutil.ReadAll(serverConfigFile)
	if err != nil {
		return err
	}

	// Extract the rpcuser
	rpcUserRegexp, err := regexp.Compile(`(?m)^\s*rpcuser=([^\s]+)`)
	if err != nil {
		return err
	}
	userSubmatches := rpcUserRegexp.FindSubmatch(content)
	if userSubmatches == nil {
		// No user found, nothing to do
		return nil
	}

	// Extract the rpcpass
	rpcPassRegexp, err := regexp.Compile(`(?m)^\s*rpcpass=([^\s]+)`)
	if err != nil {
		return err
	}
	passSubmatches := rpcPassRegexp.FindSubmatch(content)
	if passSubmatches == nil {
		// No password found, nothing to do
		return nil
	}

	// Extract the notls
	noTLSRegexp, err := regexp.Compile(`(?m)^\s*notls=(0|1)(?:\s|$)`)
	if err != nil {
		return err
	}
	noTLSSubmatches := noTLSRegexp.FindSubmatch(content)

	// Create the destination directory if it does not exists
	err = os.MkdirAll(filepath.Dir(destinationPath), 0700)
	if err != nil {
		return err
	}

	// Create the destination file and write the rpcuser and rpcpass to it
	dest, err := os.OpenFile(destinationPath,
		os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer dest.Close()

	destString := fmt.Sprintf("rpcuser=%s\nrpcpass=%s\n",
		string(userSubmatches[1]), string(passSubmatches[1]))
	if noTLSSubmatches != nil {
		destString += fmt.Sprintf("notls=%s\n", noTLSSubmatches[1])
	}

	dest.WriteString(destString)

	return nil
}
