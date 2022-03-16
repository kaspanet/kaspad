package main

import (
	"os"

	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/pkg/errors"

	"github.com/jessevdk/go-flags"
)

const (
	createSubCmd                    = "create"
	balanceSubCmd                   = "balance"
	sendSubCmd                      = "send"
	createUnsignedTransactionSubCmd = "create-unsigned-transaction"
	signSubCmd                      = "sign"
	broadcastSubCmd                 = "broadcast"
	parseSubCmd                     = "parse"
	showAddressesSubCmd             = "show-addresses"
	newAddressSubCmd                = "new-address"
	dumpUnencryptedDataSubCmd       = "dump-unencrypted-data"
	startDaemonSubCmd               = "start-daemon"
)

const (
	defaultListen    = "localhost:8082"
	defaultRPCServer = "localhost"
)

type configFlags struct {
	config.NetworkFlags
}

type createConfig struct {
	KeysFile          string `long:"keys-file" short:"f" description:"Keys file location (default: ~/.kaspawallet/keys.json (*nix), %USERPROFILE%\\AppData\\Local\\Kaspawallet\\key.json (Windows))"`
	Password          string `long:"password" short:"p" description:"Wallet password"`
	Yes               bool   `long:"yes" short:"y" description:"Assume \"yes\" to all questions"`
	MinimumSignatures uint32 `long:"min-signatures" short:"m" description:"Minimum required signatures" default:"1"`
	NumPrivateKeys    uint32 `long:"num-private-keys" short:"k" description:"Number of private keys" default:"1"`
	NumPublicKeys     uint32 `long:"num-public-keys" short:"n" description:"Total number of keys" default:"1"`
	ECDSA             bool   `long:"ecdsa" description:"Create an ECDSA wallet"`
	Import            bool   `long:"import" short:"i" description:"Import private keys (as opposed to generating them)"`
	config.NetworkFlags
}

type balanceConfig struct {
	DaemonAddress string `long:"daemonaddress" short:"d" description:"Wallet daemon server to connect to (default: localhost:8082)"`
	Verbose       bool   `long:"verbose" short:"v" description:"Verbose: show addresses with balance"`
	config.NetworkFlags
}

type sendConfig struct {
	KeysFile      string  `long:"keys-file" short:"f" description:"Keys file location (default: ~/.kaspawallet/keys.json (*nix), %USERPROFILE%\\AppData\\Local\\Kaspawallet\\key.json (Windows))"`
	Password      string  `long:"password" short:"p" description:"Wallet password"`
	DaemonAddress string  `long:"daemonaddress" short:"d" description:"Wallet daemon server to connect to (default: localhost:8082)"`
	ToAddress     string  `long:"to-address" short:"t" description:"The public address to send Kaspa to" required:"true"`
	SendAmount    float64 `long:"send-amount" short:"v" description:"An amount to send in Kaspa (e.g. 1234.12345678)" required:"true"`
	config.NetworkFlags
}

type createUnsignedTransactionConfig struct {
	DaemonAddress string  `long:"daemonaddress" short:"d" description:"Wallet daemon server to connect to (default: localhost:8082)"`
	ToAddress     string  `long:"to-address" short:"t" description:"The public address to send Kaspa to" required:"true"`
	SendAmount    float64 `long:"send-amount" short:"v" description:"An amount to send in Kaspa (e.g. 1234.12345678)" required:"true"`
	config.NetworkFlags
}

type signConfig struct {
	KeysFile        string `long:"keys-file" short:"f" description:"Keys file location (default: ~/.kaspawallet/keys.json (*nix), %USERPROFILE%\\AppData\\Local\\Kaspawallet\\key.json (Windows))"`
	Password        string `long:"password" short:"p" description:"Wallet password"`
	Transaction     string `long:"transaction" short:"t" description:"The unsigned transaction(s) to sign on (encoded in hex)"`
	TransactionFile string `long:"transaction-file" short:"F" description:"The file containing the unsigned transaction(s) to sign on (encoded in hex)"`
	config.NetworkFlags
}

type broadcastConfig struct {
	DaemonAddress    string `long:"daemonaddress" short:"d" description:"Wallet daemon server to connect to (default: localhost:8082)"`
	Transactions     string `long:"transaction" short:"t" description:"The signed transaction to broadcast (encoded in hex)"`
	TransactionsFile string `long:"transaction-file" short:"F" description:"The file containing the unsigned transaction to sign on (encoded in hex)"`
	config.NetworkFlags
}

type parseConfig struct {
	Transaction     string `long:"transaction" short:"t" description:"The transaction to parse (encoded in hex)"`
	TransactionFile string `long:"transaction-file" short:"F" description:"The file containing the transaction to parse (encoded in hex)"`
	Verbose         bool   `long:"verbose" short:"v" description:"Verbose: show transaction inputs"`
	config.NetworkFlags
}

type showAddressesConfig struct {
	DaemonAddress string `long:"daemonaddress" short:"d" description:"Wallet daemon server to connect to (default: localhost:8082)"`
	config.NetworkFlags
}

type newAddressConfig struct {
	DaemonAddress string `long:"daemonaddress" short:"d" description:"Wallet daemon server to connect to (default: localhost:8082)"`
	config.NetworkFlags
}

type startDaemonConfig struct {
	KeysFile  string `long:"keys-file" short:"f" description:"Keys file location (default: ~/.kaspawallet/keys.json (*nix), %USERPROFILE%\\AppData\\Local\\Kaspawallet\\key.json (Windows))"`
	Password  string `long:"password" short:"p" description:"Wallet password"`
	RPCServer string `long:"rpcserver" short:"s" description:"RPC server to connect to"`
	Listen    string `short:"l" long:"listen" description:"Address to listen on (default: 0.0.0.0:8082)"`
	Profile   string `long:"profile" description:"Enable HTTP profiling on given port -- NOTE port must be between 1024 and 65536"`
	config.NetworkFlags
}

type dumpUnencryptedDataConfig struct {
	KeysFile string `long:"keys-file" short:"f" description:"Keys file location (default: ~/.kaspawallet/keys.json (*nix), %USERPROFILE%\\AppData\\Local\\Kaspawallet\\key.json (Windows))"`
	Password string `long:"password" short:"p" description:"Wallet password"`
	Yes      bool   `long:"yes" short:"y" description:"Assume \"yes\" to all questions"`
	config.NetworkFlags
}

func parseCommandLine() (subCommand string, config interface{}) {
	cfg := &configFlags{}
	parser := flags.NewParser(cfg, flags.PrintErrors|flags.HelpFlag)

	createConf := &createConfig{}
	parser.AddCommand(createSubCmd, "Creates a new wallet",
		"Creates a private key and 3 public addresses, one for each of MainNet, TestNet and DevNet", createConf)

	balanceConf := &balanceConfig{DaemonAddress: defaultListen}
	parser.AddCommand(balanceSubCmd, "Shows the balance of a public address",
		"Shows the balance for a public address in Kaspa", balanceConf)

	sendConf := &sendConfig{DaemonAddress: defaultListen}
	parser.AddCommand(sendSubCmd, "Sends a Kaspa transaction to a public address",
		"Sends a Kaspa transaction to a public address", sendConf)

	createUnsignedTransactionConf := &createUnsignedTransactionConfig{DaemonAddress: defaultListen}
	parser.AddCommand(createUnsignedTransactionSubCmd, "Create an unsigned Kaspa transaction",
		"Create an unsigned Kaspa transaction", createUnsignedTransactionConf)

	signConf := &signConfig{}
	parser.AddCommand(signSubCmd, "Sign the given partially signed transaction",
		"Sign the given partially signed transaction", signConf)

	broadcastConf := &broadcastConfig{DaemonAddress: defaultListen}
	parser.AddCommand(broadcastSubCmd, "Broadcast the given transaction",
		"Broadcast the given transaction", broadcastConf)

	parseConf := &parseConfig{}
	parser.AddCommand(parseSubCmd, "Parse the given transaction and print its contents",
		"Parse the given transaction and print its contents", parseConf)

	showAddressesConf := &showAddressesConfig{DaemonAddress: defaultListen}
	parser.AddCommand(showAddressesSubCmd, "Shows all generated public addresses of the current wallet",
		"Shows all generated public addresses of the current wallet", showAddressesConf)

	newAddressConf := &newAddressConfig{DaemonAddress: defaultListen}
	parser.AddCommand(newAddressSubCmd, "Generates new public address of the current wallet and shows it",
		"Generates new public address of the current wallet and shows it", newAddressConf)

	dumpUnencryptedDataConf := &dumpUnencryptedDataConfig{}
	parser.AddCommand(dumpUnencryptedDataSubCmd, "Prints the unencrypted wallet data",
		"Prints the unencrypted wallet data including its private keys. Anyone that sees it can access "+
			"the funds. Use only on safe environment.", dumpUnencryptedDataConf)

	startDaemonConf := &startDaemonConfig{
		RPCServer: defaultRPCServer,
		Listen:    defaultListen,
	}
	parser.AddCommand(startDaemonSubCmd, "Start the wallet daemon", "Start the wallet daemon", startDaemonConf)

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
	case parseSubCmd:
		combineNetworkFlags(&parseConf.NetworkFlags, &cfg.NetworkFlags)
		err := parseConf.ResolveNetwork(parser)
		if err != nil {
			printErrorAndExit(err)
		}
		config = parseConf
	case showAddressesSubCmd:
		combineNetworkFlags(&showAddressesConf.NetworkFlags, &cfg.NetworkFlags)
		err := showAddressesConf.ResolveNetwork(parser)
		if err != nil {
			printErrorAndExit(err)
		}
		config = showAddressesConf
	case newAddressSubCmd:
		combineNetworkFlags(&newAddressConf.NetworkFlags, &cfg.NetworkFlags)
		err := newAddressConf.ResolveNetwork(parser)
		if err != nil {
			printErrorAndExit(err)
		}
		config = newAddressConf
	case dumpUnencryptedDataSubCmd:
		combineNetworkFlags(&dumpUnencryptedDataConf.NetworkFlags, &cfg.NetworkFlags)
		err := dumpUnencryptedDataConf.ResolveNetwork(parser)
		if err != nil {
			printErrorAndExit(err)
		}
		config = dumpUnencryptedDataConf
	case startDaemonSubCmd:
		combineNetworkFlags(&startDaemonConf.NetworkFlags, &cfg.NetworkFlags)
		err := startDaemonConf.ResolveNetwork(parser)
		if err != nil {
			printErrorAndExit(err)
		}
		config = startDaemonConf
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
