package main

import (
	"github.com/pkg/errors"
	"os"

	"github.com/jessevdk/go-flags"
)

const (
	createSubCmd  = "create"
	balanceSubCmd = "balance"
	sendSubCmd    = "send"
)

type createConfig struct {
}

type balanceConfig struct {
	RPCServer string `long:"rpcserver" short:"s" description:"RPC server to connect to"`
	Address   string `long:"address" short:"d" description:"The public address to check the balance of" required:"true"`
}

type sendConfig struct {
	RPCServer  string  `long:"rpcserver" short:"s" description:"RPC server to connect to"`
	PrivateKey string  `long:"private-key" short:"k" description:"The private key of the sender (encoded in hex)" required:"true"`
	ToAddress  string  `long:"to-address" short:"t" description:"The public address to send Kaspa to" required:"true"`
	SendAmount float64 `long:"send-amount" short:"v" description:"An amount to send in Kaspa (e.g. 1234.12345678)" required:"true"`
}

func parseCommandLine() (subCommand string, config interface{}) {
	cfg := &struct{}{}
	parser := flags.NewParser(cfg, flags.PrintErrors|flags.HelpFlag)

	createConf := &createConfig{}
	parser.AddCommand(createSubCmd, "Creates a new wallet",
		"Creates a private key and 3 public addresses, one for each of MainNet, TestNet and DevNet", createConf)

	balanceConf := &balanceConfig{}
	parser.AddCommand(balanceSubCmd, "Shows the balance of a public address",
		"Shows the balance for a public address in Kaspa", balanceConf)

	sendConf := &sendConfig{}
	parser.AddCommand(sendSubCmd, "Sends a Kaspa transaction to a public address",
		"Sends a Kaspa transaction to a public address", sendConf)

	_, err := parser.Parse()

	if err != nil {
		var flagsErr *flags.Error
		if ok := errors.As(err, &flagsErr); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
		return "", nil
	}

	switch parser.Command.Active.Name {
	case createSubCmd:
		config = createConf
	case balanceSubCmd:
		config = balanceConf
	case sendSubCmd:
		config = sendConf
	}

	return parser.Command.Active.Name, config
}
