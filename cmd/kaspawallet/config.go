package main

import (
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/pkg/errors"
	"os"

	"github.com/jessevdk/go-flags"
)

const (
	createSubCmd                    = "create"
	balanceSubCmd                   = "balance"
	sendSubCmd                      = "send"
	createUnsignedTransactionSubCmd = "createUnsignedTransaction"
	signSubCmd                      = "sign"
	broadcastSubCmd                 = "broadcast"
)

type configFlags struct {
	config.NetworkFlags
}

type createConfig struct {
	KeysFile          string `long:"keys-file" short:"f" description:"Keys file location (only if different from default)"`
	MinimumSignatures uint32 `long:"min-signatures" short:"m" description:"Minimum required signatures" default:"1"`
	NumPrivateKeys    uint32 `long:"num-private-keys" short:"k" description:"Number of private keys to generate" default:"1"`
	NumPublicKeys     uint32 `long:"num-public-keys" short:"n" description:"Total number of keys" default:"1"`
	config.NetworkFlags
}

type balanceConfig struct {
	KeysFile  string `long:"keys-file" short:"f" description:"Keys file location (only if different from default)"`
	RPCServer string `long:"rpcserver" short:"s" description:"RPC server to connect to"`
	config.NetworkFlags
}

type sendConfig struct {
	KeysFile   string  `long:"keys-file" short:"f" description:"Keys file location (only if different from default)"`
	RPCServer  string  `long:"rpcserver" short:"s" description:"RPC server to connect to"`
	ToAddress  string  `long:"to-address" short:"t" description:"The public address to send Kaspa to" required:"true"`
	SendAmount float64 `long:"send-amount" short:"v" description:"An amount to send in Kaspa (e.g. 1234.12345678)" required:"true"`
	config.NetworkFlags
}

type createUnsignedTransactionConfig struct {
	KeysFile   string  `long:"keys-file" short:"f" description:"Keys file location (only if different from default)"`
	RPCServer  string  `long:"rpcserver" short:"s" description:"RPC server to connect to"`
	ToAddress  string  `long:"to-address" short:"t" description:"The public address to send Kaspa to" required:"true"`
	SendAmount float64 `long:"send-amount" short:"v" description:"An amount to send in Kaspa (e.g. 1234.12345678)" required:"true"`
	config.NetworkFlags
}

type signConfig struct {
	KeysFile    string `long:"keys-file" short:"f" description:"Keys file location (only if different from default)"`
	Transaction string `long:"transaction" short:"t" description:"The unsigned transaction to sign on (encoded in hex)" required:"true"`
	config.NetworkFlags
}

type broadcastConfig struct {
	RPCServer   string `long:"rpcserver" short:"s" description:"RPC server to connect to"`
	Transaction string `long:"transaction" short:"t" description:"The signed transaction to broadcast (encoded in hex)" required:"true"`
	config.NetworkFlags
}

func parseCommandLine() (subCommand string, config interface{}) {
	cfg := &configFlags{}
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

	createUnsignedTransactionConf := &createUnsignedTransactionConfig{}
	parser.AddCommand(createUnsignedTransactionSubCmd, "Create an unsigned Kaspa transaction",
		"Create an unsigned Kaspa transaction", createUnsignedTransactionConf)

	signConf := &signConfig{}
	parser.AddCommand(signSubCmd, "Sign the given partially signed transaction",
		"Sign the given partially signed transaction", signConf)

	broadcastConf := &broadcastConfig{}
	parser.AddCommand(broadcastSubCmd, "Broadcast the given transaction",
		"Broadcast the given transaction", broadcastConf)

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
		combineNetworkFlags(&createConf.NetworkFlags, &cfg.NetworkFlags)
		err := createConf.ResolveNetwork(parser)
		if err != nil {
			printErrorAndExit(err)
		}
		config = createConf
	case balanceSubCmd:
		combineNetworkFlags(&balanceConf.NetworkFlags, &cfg.NetworkFlags)
		err := balanceConf.ResolveNetwork(parser)
		if err != nil {
			printErrorAndExit(err)
		}
		config = balanceConf
	case sendSubCmd:
		combineNetworkFlags(&sendConf.NetworkFlags, &cfg.NetworkFlags)
		err := sendConf.ResolveNetwork(parser)
		if err != nil {
			printErrorAndExit(err)
		}
		config = sendConf
	case createUnsignedTransactionSubCmd:
		combineNetworkFlags(&createUnsignedTransactionConf.NetworkFlags, &cfg.NetworkFlags)
		err := createUnsignedTransactionConf.ResolveNetwork(parser)
		if err != nil {
			printErrorAndExit(err)
		}
		config = createUnsignedTransactionConf
	case signSubCmd:
		combineNetworkFlags(&signConf.NetworkFlags, &cfg.NetworkFlags)
		err := signConf.ResolveNetwork(parser)
		if err != nil {
			printErrorAndExit(err)
		}
		config = signConf
	case broadcastSubCmd:
		combineNetworkFlags(&broadcastConf.NetworkFlags, &cfg.NetworkFlags)
		err := broadcastConf.ResolveNetwork(parser)
		if err != nil {
			printErrorAndExit(err)
		}
		config = broadcastConf
	}

	return parser.Command.Active.Name, config
}

func combineNetworkFlags(dst, src *config.NetworkFlags) {
	dst.Testnet = dst.Testnet || src.Testnet
	dst.Simnet = dst.Simnet || src.Simnet
	dst.Devnet = dst.Devnet || src.Devnet
	if dst.OverrideDAGParamsFile == "" {
		dst.OverrideDAGParamsFile = src.OverrideDAGParamsFile
	}
}
