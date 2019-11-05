package config

import (
	"fmt"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	"os"
)

// ActiveNetParams holds the selected network parameters. Default value is main-net.
var ActiveNetParams = &dagconfig.MainNetParams

// NetConfig holds the network configuration, that is which network is selected.
type NetConfig struct {
	TestNet        bool `long:"testnet" description:"Use the test network"`
	RegressionTest bool `long:"regtest" description:"Use the regression test network"`
	SimNet         bool `long:"simnet" description:"Use the simulation test network"`
	DevNet         bool `long:"devnet" description:"Use the development test network"`
}

// ParseNetConfig parses the network command line argument and sets ActiveNetParams accordingly.
// It returns error if more than one network was selected, nil otherwise.
func ParseNetConfig(netConfig NetConfig, parser *flags.Parser) error {
	// Multiple networks can't be selected simultaneously.
	numNets := 0
	// default net is main net
	// Count number of network flags passed; assign active network params
	// while we're at it
	if netConfig.TestNet {
		numNets++
		ActiveNetParams = &dagconfig.TestNetParams
	}
	if netConfig.RegressionTest {
		numNets++
		ActiveNetParams = &dagconfig.RegressionNetParams
	}
	if netConfig.SimNet {
		numNets++
		ActiveNetParams = &dagconfig.SimNetParams
	}
	if netConfig.DevNet {
		numNets++
		ActiveNetParams = &dagconfig.DevNetParams
	}
	if numNets > 1 {

		message := "Multiple networks parameters (testnet, simnet, devnet, etc.) cannot be used" +
			"together. Please choose only one network"
		err := errors.Errorf(message)
		fmt.Fprintln(os.Stderr, err)
		parser.WriteHelp(os.Stderr)
		return err
	}
	return nil
}
