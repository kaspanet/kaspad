package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kaspanet/kaspad/infrastructure/config"

	"github.com/jessevdk/go-flags"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/version"
)

const (
	defaultLogFilename    = "kaspawalletd.log"
	defaultErrLogFilename = "kaspawalletd_err.log"
)

var (
	// Default configuration options
	defaultAppDir     = util.AppDir("kaspawalletd", false)
	defaultLogFile    = filepath.Join(defaultAppDir, defaultLogFilename)
	defaultErrLogFile = filepath.Join(defaultAppDir, defaultErrLogFilename)
	defaultRPCServer  = "localhost"
	defaultListen     = "localhost:8082"
)

type configFlags struct {
	ShowVersion bool   `short:"V" long:"version" description:"Display version information and exit"`
	RPCServer   string `short:"s" long:"rpcserver" description:"RPC server to connect to"`
	Listen      string `short:"l" long:"listen" description:"Address to listen on (default: 0.0.0.0:8082)"`
	KeysFile    string `long:"keys-file" short:"f" description:"Keys file location (default: ~/.kaspawallet/keys.json (*nix), %USERPROFILE%\\AppData\\Local\\Kaspawallet\\key.json (Windows))"`
	config.NetworkFlags
}

func parseConfig() (*configFlags, error) {
	cfg := &configFlags{
		RPCServer: defaultRPCServer,
		Listen:    defaultListen,
	}
	parser := flags.NewParser(cfg, flags.PrintErrors|flags.HelpFlag)
	_, err := parser.Parse()

	// Show the version and exit if the version flag was specified.
	if cfg.ShowVersion {
		appName := filepath.Base(os.Args[0])
		appName = strings.TrimSuffix(appName, filepath.Ext(appName))
		fmt.Println(appName, "version", version.Version())
		os.Exit(0)
	}

	if err != nil {
		return nil, err
	}

	err = cfg.ResolveNetwork(parser)
	if err != nil {
		return nil, err
	}

	initLog(defaultLogFile, defaultErrLogFile)

	return cfg, nil
}
