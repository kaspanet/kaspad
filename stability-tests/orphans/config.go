package main

import (
	"os"
	"path/filepath"

	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/stability-tests/common"
	"github.com/kaspanet/kaspad/stability-tests/common/rpc"

	"github.com/jessevdk/go-flags"
)

const (
	defaultLogFilename    = "orphans.log"
	defaultErrLogFilename = "orphans_err.log"
)

var (
	// Default configuration options
	defaultLogFile    = filepath.Join(common.DefaultAppDir, defaultLogFilename)
	defaultErrLogFile = filepath.Join(common.DefaultAppDir, defaultErrLogFilename)
)

type configFlags struct {
	rpc.Config
	NodeP2PAddress    string `long:"addr" short:"a" description:"node's P2P address"`
	Profile           string `long:"profile" description:"Enable HTTP profiling on given port -- NOTE port must be between 1024 and 65536"`
	OrphanChainLength int    `long:"num-orphans" short:"n" description:"Desired length of orphan chain"`
	config.NetworkFlags
}

var cfg *configFlags

func activeConfig() *configFlags {
	return cfg
}

func parseConfig() error {
	cfg = &configFlags{}
	parser := flags.NewParser(cfg, flags.PrintErrors|flags.HelpFlag)

	_, err := parser.Parse()

	if err != nil {
		if err, ok := err.(*flags.Error); ok && err.Type == flags.ErrHelp {
			os.Exit(0)
		}
		return err
	}

	err = cfg.ResolveNetwork(parser)
	if err != nil {
		return err
	}
	log.SetLevel(logger.LevelInfo)
	common.InitBackend(backendLog, defaultLogFile, defaultErrLogFile)

	return nil
}
