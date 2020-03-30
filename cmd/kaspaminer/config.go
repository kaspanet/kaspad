package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kaspanet/kaspad/config"

	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"

	"github.com/jessevdk/go-flags"
	"github.com/kaspanet/kaspad/version"
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
	Profile        string `long:"profile" description:"Enable HTTP profiling on given port -- NOTE port must be between 1024 and 65536"`
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

	if cfg.Profile != "" {
		profilePort, err := strconv.Atoi(cfg.Profile)
		if err != nil || profilePort < 1024 || profilePort > 65535 {
			return nil, errors.New("The profile port must be between 1024 and 65535")
		}
	}

	initLog(defaultLogFile, defaultErrLogFile)

	return cfg, nil
}
