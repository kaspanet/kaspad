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

// NetworkFlags holds the network configuration, that is which network is selected.
type NetworkFlags struct {
	TestNet        bool `long:"testnet" description:"Use the test network"`
	RegressionTest bool `long:"regtest" description:"Use the regression test network"`
	SimNet         bool `long:"simnet" description:"Use the simulation test network"`
	DevNet         bool `long:"devnet" description:"Use the development test network"`
}

// ResolveNetwork parses the network command line argument and sets ActiveNetParams accordingly.
// It returns error if more than one network was selected, nil otherwise.
func ResolveNetwork(networkFlags NetworkFlags, parser *flags.Parser) error {
	// Multiple networks can't be selected simultaneously.
	numNets := 0
	// default net is main net
	// Count number of network flags passed; assign active network params
	// while we're at it
	if networkFlags.TestNet {
		numNets++
		ActiveNetParams = &dagconfig.TestNetParams
	}
	if networkFlags.RegressionTest {
		numNets++
		ActiveNetParams = &dagconfig.RegressionNetParams
	}
	if networkFlags.SimNet {
		numNets++
		ActiveNetParams = &dagconfig.SimNetParams
	}
	if networkFlags.DevNet {
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
