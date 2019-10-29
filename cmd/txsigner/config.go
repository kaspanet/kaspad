package main

import (
	"fmt"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/jessevdk/go-flags"
	"os"
)

var activeNetParams = &dagconfig.MainNetParams

type config struct {
	Transaction    string `long:"transaction" short:"t" description:"Unsigned transaction in HEX format" required:"true"`
	PrivateKey     string `long:"private-key" short:"p" description:"Private key" required:"true"`
	TestNet        bool   `long:"testnet" description:"Use the test network"`
	RegressionTest bool   `long:"regtest" description:"Use the regression test network"`
	SimNet         bool   `long:"simnet" description:"Use the simulation test network"`
	DevNet         bool   `long:"devnet" description:"Use the development test network"`
}

func parseCommandLine() (*config, error) {
	cfg := &config{}
	parser := flags.NewParser(cfg, flags.PrintErrors|flags.HelpFlag)
	_, err := parser.Parse()
	// Multiple networks can't be selected simultaneously.
	funcName := "loadConfig"
	numNets := 0
	// Count number of network flags passed; assign active network params
	// while we're at it
	if cfg.TestNet {
		numNets++
		activeNetParams = &dagconfig.TestNetParams
	}
	if cfg.RegressionTest {
		numNets++
		activeNetParams = &dagconfig.RegressionNetParams
	}
	if cfg.SimNet {
		numNets++
		activeNetParams = &dagconfig.SimNetParams
	}
	if cfg.DevNet {
		numNets++
		activeNetParams = &dagconfig.DevNetParams
	}
	if numNets > 1 {
		str := "%s: The testnet, regtest, simnet and devent params can't be " +
			"used together -- choose one of the four"
		err := fmt.Errorf(str, funcName)
		fmt.Fprintln(os.Stderr, err)
		parser.WriteHelp(os.Stderr)
		return nil, err
	}

	return cfg, err
}
