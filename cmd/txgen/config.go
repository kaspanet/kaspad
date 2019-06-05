package main

import (
	"errors"
	"github.com/daglabs/btcd/util"
	"path/filepath"

	"github.com/jessevdk/go-flags"
)

const (
	defaultLogFilename = "txgen.log"
)

var (
	// Default configuration options
	defaultHomeDir                      = util.AppDataDir("txgen", false)
	defaultLogFile                      = filepath.Join(defaultHomeDir, defaultLogFilename)
	defaultTargetNumberOfOutputs uint64 = 1
	defaultTargetNumberOfInputs  uint64 = 1
)

type config struct {
	Address               string  `long:"address" description:"An address to a JSON-RPC endpoints" required:"true"`
	PrivateKey            string  `long:"private-key" description:"Private key" required:"true"`
	CertificatePath       string  `long:"cert" description:"Path to certificate accepted by JSON-RPC endpoint"`
	DisableTLS            bool    `long:"notls" description:"Disable TLS"`
	TxInterval            uint64  `long:"tx-interval" description:"Transaction emission interval (in milliseconds)"`
	TargetNumberOfOutputs uint64  `long:"num-outputs" description:"Target number of transaction outputs (with some randomization)"`
	TargetNumberOfInputs  uint64  `long:"num-inputs" description:"Target number of transaction inputs (with some randomization)"`
	AveragePayloadSize    uint64  `long:"payload-size" description:"Average size of transaction payload"`
	AverageGasFraction    float64 `long:"gas-fraction" description:"The average portion of gas from the gas limit"`
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

	if cfg.AverageGasFraction >= 1 {
		return nil, errors.New("--gas-fraction should be between 0 and 1")
	}

	if cfg.TargetNumberOfOutputs == 0 {
		cfg.TargetNumberOfOutputs = defaultTargetNumberOfOutputs
	}

	if cfg.TargetNumberOfInputs == 0 {
		cfg.TargetNumberOfInputs = defaultTargetNumberOfInputs
	}

	initLogRotator(defaultLogFile)

	return cfg, nil
}
