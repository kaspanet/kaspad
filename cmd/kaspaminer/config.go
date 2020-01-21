package main

import (
	"fmt"
	"github.com/kaspanet/kaspad/config"
	"os"
	"path/filepath"
	"strings"

	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"

	"github.com/jessevdk/go-flags"
	"github.com/kaspanet/kaspad/cmd/kaspaminer/version"
)

const (
	defaultLogFilename    = "kaspaminer.log"
	defaultErrLogFilename = "kaspaminer_err.log"
)

var (
	// Default configuration options
	defaultHomeDir    = util.AppDataDir("kaspaminer", false)
	defaultLogFile    = filepath.Join(defaultHomeDir, defaultLogFilename)
	defaultErrLogFile = filepath.Join(defaultHomeDir, defaultErrLogFilename)
	defaultRPCServer  = "localhost"
)

type configFlags struct {
	ShowVersion    bool   `short:"V" long:"version" description:"Display version information and exit"`
	RPCUser        string `short:"u" long:"rpcuser" description:"RPC username"`
	RPCPassword    string `short:"P" long:"rpcpass" default-mask:"-" description:"RPC password"`
	RPCServer      string `short:"s" long:"rpcserver" description:"RPC server to connect to"`
	RPCCert        string `short:"c" long:"rpccert" description:"RPC server certificate chain for validation"`
	DisableTLS     bool   `long:"notls" description:"Disable TLS"`
	Verbose        bool   `long:"verbose" short:"v" description:"Enable logging of RPC requests"`
	NumberOfBlocks uint64 `short:"n" long:"numblocks" description:"Number of blocks to mine. If omitted, will mine until the process is interrupted."`
	BlockDelay     uint64 `long:"block-delay" description:"Delay for block submission (in milliseconds). This is used only for testing purposes."`
	config.NetworkFlags
}

func parseConfig() (*configFlags, error) {
	cfg := &configFlags{
		RPCServer: defaultRPCServer,
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

	if cfg.RPCUser == "" {
		return nil, errors.New("--rpcuser is required")
	}
	if cfg.RPCPassword == "" {
		return nil, errors.New("--rpcpass is required")
	}

	if cfg.RPCCert == "" && !cfg.DisableTLS {
		return nil, errors.New("--notls has to be disabled if --cert is used")
	}
	if cfg.RPCCert != "" && cfg.DisableTLS {
		return nil, errors.New("--rpccert should be omitted if --notls is used")
	}

	initLog(defaultLogFile, defaultErrLogFile)

	return cfg, nil
}
