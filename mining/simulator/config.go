package main

import (
	"path/filepath"

	"github.com/daglabs/kaspad/util"
	"github.com/pkg/errors"

	"github.com/jessevdk/go-flags"
)

const (
	defaultLogFilename    = "miningsimulator.log"
	defaultErrLogFilename = "miningsimulator_err.log"
)

var (
	// Default configuration options
	defaultHomeDir    = util.AppDataDir("miningsimulator", false)
	defaultLogFile    = filepath.Join(defaultHomeDir, defaultLogFilename)
	defaultErrLogFile = filepath.Join(defaultHomeDir, defaultErrLogFilename)
)

type config struct {
	AutoScalingGroup string `long:"autoscaling" description:"AWS AutoScalingGroup to use for address list"`
	Region           string `long:"region" description:"AWS region to use for address list"`
	AddressListPath  string `long:"addresslist" description:"Path to a list of nodes' JSON-RPC endpoints"`
	CertificatePath  string `long:"cert" description:"Path to certificate accepted by JSON-RPC endpoint"`
	DisableTLS       bool   `long:"notls" description:"Disable TLS"`
	Verbose          bool   `long:"verbose" short:"v" description:"Enable logging of RPC requests"`
	BlockDelay       uint64 `long:"block-delay" description:"Delay for block submission (in milliseconds)"`
}

func parseConfig() (*config, error) {
	cfg := &config{}
	parser := flags.NewParser(cfg, flags.PrintErrors|flags.HelpFlag)
	_, err := parser.Parse()

	if err != nil {
		return nil, err
	}

	if cfg.CertificatePath == "" && !cfg.DisableTLS {
		return nil, errors.New("--notls has to be disabled if --cert is used")
	}

	if cfg.CertificatePath != "" && cfg.DisableTLS {
		return nil, errors.New("--cert should be omitted if --notls is used")
	}

	if (cfg.AutoScalingGroup == "" || cfg.Region == "") && cfg.AddressListPath == "" {
		return nil, errors.New("Either (--autoscaling and --region) or --addresslist must be specified")
	}

	if (cfg.AutoScalingGroup != "" || cfg.Region != "") && cfg.AddressListPath != "" {
		return nil, errors.New("Both (--autoscaling and --region) and --addresslist can't be specified at the same time")
	}

	if cfg.AutoScalingGroup != "" && cfg.Region == "" {
		return nil, errors.New("If --autoscaling is specified --region must be specified as well")
	}

	initLog(defaultLogFile, defaultErrLogFile)

	return cfg, nil
}
