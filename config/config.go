// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package config

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/btcsuite/go-socks/socks"
	"github.com/jessevdk/go-flags"
	"github.com/kaspanet/kaspad/database"
	"github.com/kaspanet/kaspad/logger"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/network"
	"github.com/kaspanet/kaspad/util/subnetworkid"
	"github.com/kaspanet/kaspad/version"
)

const (
	defaultConfigFilename      = "kaspad.conf"
	defaultDataDirname         = "data"
	defaultLogLevel            = "info"
	defaultLogDirname          = "logs"
	defaultLogFilename         = "kaspad.log"
	defaultErrLogFilename      = "kaspad_err.log"
	defaultTargetOutboundPeers = 8
	defaultMaxInboundPeers     = 117
	defaultBanDuration         = time.Hour * 24
	defaultBanThreshold        = 100
	//DefaultConnectTimeout is the default connection timeout when dialing
	DefaultConnectTimeout        = time.Second * 30
	defaultMaxRPCClients         = 10
	defaultMaxRPCWebsockets      = 25
	defaultMaxRPCConcurrentReqs  = 20
	defaultDbType                = "ffldb"
	defaultBlockMaxMass          = 10000000
	blockMaxMassMin              = 1000
	blockMaxMassMax              = 10000000
	defaultMinRelayTxFee         = 1e-5 // 1 sompi per byte
	defaultGenerate              = false
	defaultMaxOrphanTransactions = 100
	//DefaultMaxOrphanTxSize is the default maximum size for an orphan transaction
	DefaultMaxOrphanTxSize = 100000
	defaultSigCacheMaxSize = 100000
	sampleConfigFilename   = "sample-kaspad.conf"
	defaultTxIndex         = false
	defaultAddrIndex       = false
	defaultAcceptanceIndex = false
)

var (
	// DefaultHomeDir is the default home directory for kaspad.
	DefaultHomeDir = util.AppDataDir("kaspad", false)

	defaultConfigFile  = filepath.Join(DefaultHomeDir, defaultConfigFilename)
	defaultDataDir     = filepath.Join(DefaultHomeDir, defaultDataDirname)
	knownDbTypes       = database.SupportedDrivers()
	defaultRPCKeyFile  = filepath.Join(DefaultHomeDir, "rpc.key")
	defaultRPCCertFile = filepath.Join(DefaultHomeDir, "rpc.cert")
	defaultLogDir      = filepath.Join(DefaultHomeDir, defaultLogDirname)
)

var activeConfig *Config

// RunServiceCommand is only set to a real function on Windows. It is used
// to parse and execute service commands specified via the -s flag.
var RunServiceCommand func(string) error

// minUint32 is a helper function to return the minimum of two uint32s.
// This avoids a math import and the need to cast to floats.
func minUint32(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}

// Flags defines the configuration options for kaspad.
//
// See loadConfig for details on the configuration load process.
type Flags struct {
	ShowVersion          bool          `short:"V" long:"version" description:"Display version information and exit"`
	ConfigFile           string        `short:"C" long:"configfile" description:"Path to configuration file"`
	DataDir              string        `short:"b" long:"datadir" description:"Directory to store data"`
	LogDir               string        `long:"logdir" description:"Directory to log output."`
	AddPeers             []string      `short:"a" long:"addpeer" description:"Add a peer to connect with at startup"`
	ConnectPeers         []string      `long:"connect" description:"Connect only to the specified peers at startup"`
	DisableListen        bool          `long:"nolisten" description:"Disable listening for incoming connections -- NOTE: Listening is automatically disabled if the --connect or --proxy options are used without also specifying listen interfaces via --listen"`
	Listeners            []string      `long:"listen" description:"Add an interface/port to listen for connections (default all interfaces port: 16111, testnet: 16211)"`
	TargetOutboundPeers  int           `long:"outpeers" description:"Target number of outbound peers"`
	MaxInboundPeers      int           `long:"maxinpeers" description:"Max number of inbound peers"`
	DisableBanning       bool          `long:"nobanning" description:"Disable banning of misbehaving peers"`
	BanDuration          time.Duration `long:"banduration" description:"How long to ban misbehaving peers. Valid time units are {s, m, h}. Minimum 1 second"`
	BanThreshold         uint32        `long:"banthreshold" description:"Maximum allowed ban score before disconnecting and banning misbehaving peers."`
	Whitelists           []string      `long:"whitelist" description:"Add an IP network or IP that will not be banned. (eg. 192.168.1.0/24 or ::1)"`
	RPCUser              string        `short:"u" long:"rpcuser" description:"Username for RPC connections"`
	RPCPass              string        `short:"P" long:"rpcpass" default-mask:"-" description:"Password for RPC connections"`
	RPCLimitUser         string        `long:"rpclimituser" description:"Username for limited RPC connections"`
	RPCLimitPass         string        `long:"rpclimitpass" default-mask:"-" description:"Password for limited RPC connections"`
	RPCListeners         []string      `long:"rpclisten" description:"Add an interface/port to listen for RPC connections (default port: 16110, testnet: 16210)"`
	RPCCert              string        `long:"rpccert" description:"File containing the certificate file"`
	RPCKey               string        `long:"rpckey" description:"File containing the certificate key"`
	RPCMaxClients        int           `long:"rpcmaxclients" description:"Max number of RPC clients for standard connections"`
	RPCMaxWebsockets     int           `long:"rpcmaxwebsockets" description:"Max number of RPC websocket connections"`
	RPCMaxConcurrentReqs int           `long:"rpcmaxconcurrentreqs" description:"Max number of concurrent RPC requests that may be processed concurrently"`
	DisableRPC           bool          `long:"norpc" description:"Disable built-in RPC server -- NOTE: The RPC server is disabled by default if no rpcuser/rpcpass or rpclimituser/rpclimitpass is specified"`
	DisableTLS           bool          `long:"notls" description:"Disable TLS for the RPC server -- NOTE: This is only allowed if the RPC server is bound to localhost"`
	DisableDNSSeed       bool          `long:"nodnsseed" description:"Disable DNS seeding for peers"`
	DNSSeed              string        `long:"dnsseed" description:"Override DNS seeds with specified hostname (Only 1 hostname allowed)"`
	ExternalIPs          []string      `long:"externalip" description:"Add an ip to the list of local addresses we claim to listen on to peers"`
	Proxy                string        `long:"proxy" description:"Connect via SOCKS5 proxy (eg. 127.0.0.1:9050)"`
	ProxyUser            string        `long:"proxyuser" description:"Username for proxy server"`
	ProxyPass            string        `long:"proxypass" default-mask:"-" description:"Password for proxy server"`
	OnionProxy           string        `long:"onion" description:"Connect to tor hidden services via SOCKS5 proxy (eg. 127.0.0.1:9050)"`
	OnionProxyUser       string        `long:"onionuser" description:"Username for onion proxy server"`
	OnionProxyPass       string        `long:"onionpass" default-mask:"-" description:"Password for onion proxy server"`
	NoOnion              bool          `long:"noonion" description:"Disable connecting to tor hidden services"`
	TorIsolation         bool          `long:"torisolation" description:"Enable Tor stream isolation by randomizing user credentials for each connection."`
	DbType               string        `long:"dbtype" description:"Database backend to use for the Block DAG"`
	Profile              string        `long:"profile" description:"Enable HTTP profiling on given port -- NOTE port must be between 1024 and 65536"`
	CPUProfile           string        `long:"cpuprofile" description:"Write CPU profile to the specified file"`
	DebugLevel           string        `short:"d" long:"debuglevel" description:"Logging level for all subsystems {trace, debug, info, warn, error, critical} -- You may also specify <subsystem>=<level>,<subsystem2>=<level>,... to set the log level for individual subsystems -- Use show to list available subsystems"`
	Upnp                 bool          `long:"upnp" description:"Use UPnP to map our listening port outside of NAT"`
	MinRelayTxFee        float64       `long:"minrelaytxfee" description:"The minimum transaction fee in KAS/kB to be considered a non-zero fee."`
	MaxOrphanTxs         int           `long:"maxorphantx" description:"Max number of orphan transactions to keep in memory"`
	Generate             bool          `long:"generate" description:"Generate (mine) kaspa using the CPU"`
	MiningAddrs          []string      `long:"miningaddr" description:"Add the specified payment address to the list of addresses to use for generated blocks -- At least one address is required if the generate option is set"`
	BlockMaxMass         uint64        `long:"blockmaxmass" description:"Maximum transaction mass to be used when creating a block"`
	UserAgentComments    []string      `long:"uacomment" description:"Comment to add to the user agent -- See BIP 14 for more information."`
	NoPeerBloomFilters   bool          `long:"nopeerbloomfilters" description:"Disable bloom filtering support"`
	SigCacheMaxSize      uint          `long:"sigcachemaxsize" description:"The maximum number of entries in the signature verification cache"`
	BlocksOnly           bool          `long:"blocksonly" description:"Do not accept transactions from remote peers."`
	TxIndex              bool          `long:"txindex" description:"Maintain a full hash-based transaction index which makes all transactions available via the getrawtransaction RPC"`
	DropTxIndex          bool          `long:"droptxindex" description:"Deletes the hash-based transaction index from the database on start up and then exits."`
	AddrIndex            bool          `long:"addrindex" description:"Maintain a full address-based transaction index which makes the searchrawtransactions RPC available"`
	DropAddrIndex        bool          `long:"dropaddrindex" description:"Deletes the address-based transaction index from the database on start up and then exits."`
	AcceptanceIndex      bool          `long:"acceptanceindex" description:"Maintain a full hash-based acceptance index which makes the getChainByBlock RPC available"`
	DropAcceptanceIndex  bool          `long:"dropacceptanceindex" description:"Deletes the hash-based acceptance index from the database on start up and then exits."`
	RelayNonStd          bool          `long:"relaynonstd" description:"Relay non-standard transactions regardless of the default settings for the active network."`
	RejectNonStd         bool          `long:"rejectnonstd" description:"Reject non-standard transactions regardless of the default settings for the active network."`
	Subnetwork           string        `long:"subnetwork" description:"If subnetwork ID is specified, than node will request and process only payloads from specified subnetwork. And if subnetwork ID is ommited, than payloads of all subnetworks are processed. Subnetworks with IDs 2 through 255 are reserved for future use and are currently not allowed."`
	ResetDatabase        bool          `long:"reset-db" description:"Reset database before starting node. It's needed when switching between subnetworks."`
	NetworkFlags
}

// Config defines the configuration options for kaspad.
//
// See loadConfig for details on the configuration load process.
type Config struct {
	*Flags
	Lookup        func(string) ([]net.IP, error)
	OnionDial     func(string, string, time.Duration) (net.Conn, error)
	Dial          func(string, string, time.Duration) (net.Conn, error)
	MiningAddrs   []util.Address
	MinRelayTxFee util.Amount
	Whitelists    []*net.IPNet
	SubnetworkID  *subnetworkid.SubnetworkID // nil in full nodes
}

// serviceOptions defines the configuration options for the daemon as a service on
// Windows.
type serviceOptions struct {
	ServiceCommand string `short:"s" long:"service" description:"Service command {install, remove, start, stop}"`
}

// cleanAndExpandPath expands environment variables and leading ~ in the
// passed path, cleans the result, and returns it.
func cleanAndExpandPath(path string) string {
	// Expand initial ~ to OS specific home directory.
	if strings.HasPrefix(path, "~") {
		homeDir := filepath.Dir(DefaultHomeDir)
		path = strings.Replace(path, "~", homeDir, 1)
	}

	// NOTE: The os.ExpandEnv doesn't work with Windows-style %VARIABLE%,
	// but they variables can still be expanded via POSIX-style $VARIABLE.
	return filepath.Clean(os.ExpandEnv(path))
}

// validDbType returns whether or not dbType is a supported database type.
func validDbType(dbType string) bool {
	for _, knownType := range knownDbTypes {
		if dbType == knownType {
			return true
		}
	}

	return false
}

// newConfigParser returns a new command line flags parser.
func newConfigParser(cfgFlags *Flags, so *serviceOptions, options flags.Options) *flags.Parser {
	parser := flags.NewParser(cfgFlags, options)
	if runtime.GOOS == "windows" {
		parser.AddGroup("Service Options", "Service Options", so)
	}
	return parser
}

//LoadAndSetActiveConfig loads the config that can be afterward be accesible through ActiveConfig()
func LoadAndSetActiveConfig() error {
	tcfg, _, err := loadConfig()
	if err != nil {
		return err
	}
	activeConfig = tcfg
	return nil
}

// ActiveConfig is a getter to the main config
func ActiveConfig() *Config {
	return activeConfig
}

// loadConfig initializes and parses the config using a config file and command
// line options.
//
// The configuration proceeds as follows:
// 	1) Start with a default config with sane settings
// 	2) Pre-parse the command line to check for an alternative config file
// 	3) Load configuration file overwriting defaults with any specified options
// 	4) Parse CLI options and overwrite/add any specified options
//
// The above results in kaspad functioning properly without any config settings
// while still allowing the user to override settings with config files and
// command line options. Command line options always take precedence.
func loadConfig() (*Config, []string, error) {
	// Default config.
	cfgFlags := Flags{
		ConfigFile:           defaultConfigFile,
		DebugLevel:           defaultLogLevel,
		TargetOutboundPeers:  defaultTargetOutboundPeers,
		MaxInboundPeers:      defaultMaxInboundPeers,
		BanDuration:          defaultBanDuration,
		BanThreshold:         defaultBanThreshold,
		RPCMaxClients:        defaultMaxRPCClients,
		RPCMaxWebsockets:     defaultMaxRPCWebsockets,
		RPCMaxConcurrentReqs: defaultMaxRPCConcurrentReqs,
		DataDir:              defaultDataDir,
		LogDir:               defaultLogDir,
		DbType:               defaultDbType,
		RPCKey:               defaultRPCKeyFile,
		RPCCert:              defaultRPCCertFile,
		BlockMaxMass:         defaultBlockMaxMass,
		MaxOrphanTxs:         defaultMaxOrphanTransactions,
		SigCacheMaxSize:      defaultSigCacheMaxSize,
		MinRelayTxFee:        defaultMinRelayTxFee,
		Generate:             defaultGenerate,
		TxIndex:              defaultTxIndex,
		AddrIndex:            defaultAddrIndex,
		AcceptanceIndex:      defaultAcceptanceIndex,
	}

	// Service options which are only added on Windows.
	serviceOpts := serviceOptions{}

	// Pre-parse the command line options to see if an alternative config
	// file or the version flag was specified. Any errors aside from the
	// help message error can be ignored here since they will be caught by
	// the final parse below.
	preCfg := cfgFlags
	preParser := newConfigParser(&preCfg, &serviceOpts, flags.HelpFlag)
	_, err := preParser.Parse()
	if err != nil {
		if e, ok := err.(*flags.Error); ok && e.Type == flags.ErrHelp {
			fmt.Fprintln(os.Stderr, err)
			return nil, nil, err
		}
	}

	// Show the version and exit if the version flag was specified.
	appName := filepath.Base(os.Args[0])
	appName = strings.TrimSuffix(appName, filepath.Ext(appName))
	usageMessage := fmt.Sprintf("Use %s -h to show usage", appName)
	if preCfg.ShowVersion {
		fmt.Println(appName, "version", version.Version())
		os.Exit(0)
	}

	// Perform service command and exit if specified. Invalid service
	// commands show an appropriate error. Only runs on Windows since
	// the RunServiceCommand function will be nil when not on Windows.
	if serviceOpts.ServiceCommand != "" && RunServiceCommand != nil {
		err := RunServiceCommand(serviceOpts.ServiceCommand)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(0)
	}

	// Load additional config from file.
	var configFileError error
	parser := newConfigParser(&cfgFlags, &serviceOpts, flags.Default)
	activeConfig = &Config{
		Flags: &cfgFlags,
	}
	if !(preCfg.RegressionTest || preCfg.SimNet) || preCfg.ConfigFile !=
		defaultConfigFile {

		if _, err := os.Stat(preCfg.ConfigFile); os.IsNotExist(err) {
			err := createDefaultConfigFile(preCfg.ConfigFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating a "+
					"default config file: %s\n", err)
			}
		}

		err := flags.NewIniParser(parser).ParseFile(preCfg.ConfigFile)
		if err != nil {
			if _, ok := err.(*os.PathError); !ok {
				fmt.Fprintf(os.Stderr, "Error parsing config "+
					"file: %s\n", err)
				fmt.Fprintln(os.Stderr, usageMessage)
				return nil, nil, err
			}
			configFileError = err
		}
	}

	// Don't add peers from the config file when in regression test mode.
	if preCfg.RegressionTest && len(activeConfig.AddPeers) > 0 {
		activeConfig.AddPeers = nil
	}

	// Parse command line options again to ensure they take precedence.
	remainingArgs, err := parser.Parse()
	if err != nil {
		if e, ok := err.(*flags.Error); !ok || e.Type != flags.ErrHelp {
			fmt.Fprintln(os.Stderr, usageMessage)
		}
		return nil, nil, err
	}

	// Create the home directory if it doesn't already exist.
	funcName := "loadConfig"
	err = os.MkdirAll(DefaultHomeDir, 0700)
	if err != nil {
		// Show a nicer error message if it's because a symlink is
		// linked to a directory that does not exist (probably because
		// it's not mounted).
		if e, ok := err.(*os.PathError); ok && os.IsExist(err) {
			if link, lerr := os.Readlink(e.Path); lerr == nil {
				str := "is symlink %s -> %s mounted?"
				err = errors.Errorf(str, e.Path, link)
			}
		}

		str := "%s: Failed to create home directory: %s"
		err := errors.Errorf(str, funcName, err)
		fmt.Fprintln(os.Stderr, err)
		return nil, nil, err
	}

	if !activeConfig.DisableRPC {
		if activeConfig.RPCUser == "" {
			str := "%s: rpcuser cannot be empty"
			err := errors.Errorf(str, funcName)
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(os.Stderr, usageMessage)
			return nil, nil, err
		}

		if activeConfig.RPCPass == "" {
			str := "%s: rpcpass cannot be empty"
			err := errors.Errorf(str, funcName)
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(os.Stderr, usageMessage)
			return nil, nil, err
		}
	}

	err = activeConfig.ResolveNetwork(parser)
	if err != nil {
		return nil, nil, err
	}

	// Set the default policy for relaying non-standard transactions
	// according to the default of the active network. The set
	// configuration value takes precedence over the default value for the
	// selected network.
	relayNonStd := activeConfig.NetParams().RelayNonStdTxs
	switch {
	case activeConfig.RelayNonStd && activeConfig.RejectNonStd:
		str := "%s: rejectnonstd and relaynonstd cannot be used " +
			"together -- choose only one"
		err := errors.Errorf(str, funcName)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	case activeConfig.RejectNonStd:
		relayNonStd = false
	case activeConfig.RelayNonStd:
		relayNonStd = true
	}
	activeConfig.RelayNonStd = relayNonStd

	// Append the network type to the data directory so it is "namespaced"
	// per network. In addition to the block database, there are other
	// pieces of data that are saved to disk such as address manager state.
	// All data is specific to a network, so namespacing the data directory
	// means each individual piece of serialized data does not have to
	// worry about changing names per network and such.
	activeConfig.DataDir = cleanAndExpandPath(activeConfig.DataDir)
	activeConfig.DataDir = filepath.Join(activeConfig.DataDir, activeConfig.NetParams().Name)

	// Append the network type to the log directory so it is "namespaced"
	// per network in the same fashion as the data directory.
	activeConfig.LogDir = cleanAndExpandPath(activeConfig.LogDir)
	activeConfig.LogDir = filepath.Join(activeConfig.LogDir, activeConfig.NetParams().Name)

	// Special show command to list supported subsystems and exit.
	if activeConfig.DebugLevel == "show" {
		fmt.Println("Supported subsystems", logger.SupportedSubsystems())
		os.Exit(0)
	}

	// Initialize log rotation. After log rotation has been initialized, the
	// logger variables may be used.
	logger.InitLog(filepath.Join(activeConfig.LogDir, defaultLogFilename), filepath.Join(activeConfig.LogDir, defaultErrLogFilename))

	// Parse, validate, and set debug log level(s).
	if err := logger.ParseAndSetDebugLevels(activeConfig.DebugLevel); err != nil {
		err := errors.Errorf("%s: %s", funcName, err.Error())
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}

	// Validate database type.
	if !validDbType(activeConfig.DbType) {
		str := "%s: The specified database type [%s] is invalid -- " +
			"supported types %s"
		err := errors.Errorf(str, funcName, activeConfig.DbType, knownDbTypes)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}

	// Validate profile port number
	if activeConfig.Profile != "" {
		profilePort, err := strconv.Atoi(activeConfig.Profile)
		if err != nil || profilePort < 1024 || profilePort > 65535 {
			str := "%s: The profile port must be between 1024 and 65535"
			err := errors.Errorf(str, funcName)
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(os.Stderr, usageMessage)
			return nil, nil, err
		}
	}

	// Don't allow ban durations that are too short.
	if activeConfig.BanDuration < time.Second {
		str := "%s: The banduration option may not be less than 1s -- parsed [%s]"
		err := errors.Errorf(str, funcName, activeConfig.BanDuration)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}

	// Validate any given whitelisted IP addresses and networks.
	if len(activeConfig.Whitelists) > 0 {
		var ip net.IP
		activeConfig.Whitelists = make([]*net.IPNet, 0, len(activeConfig.Flags.Whitelists))

		for _, addr := range activeConfig.Flags.Whitelists {
			_, ipnet, err := net.ParseCIDR(addr)
			if err != nil {
				ip = net.ParseIP(addr)
				if ip == nil {
					str := "%s: The whitelist value of '%s' is invalid"
					err = errors.Errorf(str, funcName, addr)
					fmt.Fprintln(os.Stderr, err)
					fmt.Fprintln(os.Stderr, usageMessage)
					return nil, nil, err
				}
				var bits int
				if ip.To4() == nil {
					// IPv6
					bits = 128
				} else {
					bits = 32
				}
				ipnet = &net.IPNet{
					IP:   ip,
					Mask: net.CIDRMask(bits, bits),
				}
			}
			activeConfig.Whitelists = append(activeConfig.Whitelists, ipnet)
		}
	}

	// --addPeer and --connect do not mix.
	if len(activeConfig.AddPeers) > 0 && len(activeConfig.ConnectPeers) > 0 {
		str := "%s: the --addpeer and --connect options can not be " +
			"mixed"
		err := errors.Errorf(str, funcName)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}

	// --proxy or --connect without --listen disables listening.
	if (activeConfig.Proxy != "" || len(activeConfig.ConnectPeers) > 0) &&
		len(activeConfig.Listeners) == 0 {
		activeConfig.DisableListen = true
	}

	// Connect means no DNS seeding.
	if len(activeConfig.ConnectPeers) > 0 {
		activeConfig.DisableDNSSeed = true
	}

	// Add the default listener if none were specified. The default
	// listener is all addresses on the listen port for the network
	// we are to connect to.
	if len(activeConfig.Listeners) == 0 {
		activeConfig.Listeners = []string{
			net.JoinHostPort("", activeConfig.NetParams().DefaultPort),
		}
	}

	// Check to make sure limited and admin users don't have the same username
	if activeConfig.RPCUser == activeConfig.RPCLimitUser && activeConfig.RPCUser != "" {
		str := "%s: --rpcuser and --rpclimituser must not specify the " +
			"same username"
		err := errors.Errorf(str, funcName)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}

	// Check to make sure limited and admin users don't have the same password
	if activeConfig.RPCPass == activeConfig.RPCLimitPass && activeConfig.RPCPass != "" {
		str := "%s: --rpcpass and --rpclimitpass must not specify the " +
			"same password"
		err := errors.Errorf(str, funcName)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}

	// The RPC server is disabled if no username or password is provided.
	if (activeConfig.RPCUser == "" || activeConfig.RPCPass == "") &&
		(activeConfig.RPCLimitUser == "" || activeConfig.RPCLimitPass == "") {
		activeConfig.DisableRPC = true
	}

	if activeConfig.DisableRPC {
		log.Infof("RPC service is disabled")
	}

	// Default RPC to listen on localhost only.
	if !activeConfig.DisableRPC && len(activeConfig.RPCListeners) == 0 {
		addrs, err := net.LookupHost("localhost")
		if err != nil {
			return nil, nil, err
		}
		activeConfig.RPCListeners = make([]string, 0, len(addrs))
		for _, addr := range addrs {
			addr = net.JoinHostPort(addr, activeConfig.NetParams().RPCPort)
			activeConfig.RPCListeners = append(activeConfig.RPCListeners, addr)
		}
	}

	if activeConfig.RPCMaxConcurrentReqs < 0 {
		str := "%s: The rpcmaxwebsocketconcurrentrequests option may " +
			"not be less than 0 -- parsed [%d]"
		err := errors.Errorf(str, funcName, activeConfig.RPCMaxConcurrentReqs)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}

	// Validate the the minrelaytxfee.
	activeConfig.MinRelayTxFee, err = util.NewAmount(activeConfig.Flags.MinRelayTxFee)
	if err != nil {
		str := "%s: invalid minrelaytxfee: %s"
		err := errors.Errorf(str, funcName, err)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}

	// Disallow 0 and negative min tx fees.
	if activeConfig.MinRelayTxFee <= 0 {
		str := "%s: The minrelaytxfee option must be greater than 0 -- parsed [%d]"
		err := errors.Errorf(str, funcName, activeConfig.MinRelayTxFee)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}

	// Limit the max block mass to a sane value.
	if activeConfig.BlockMaxMass < blockMaxMassMin || activeConfig.BlockMaxMass >
		blockMaxMassMax {

		str := "%s: The blockmaxmass option must be in between %d " +
			"and %d -- parsed [%d]"
		err := errors.Errorf(str, funcName, blockMaxMassMin,
			blockMaxMassMax, activeConfig.BlockMaxMass)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}

	// Limit the max orphan count to a sane value.
	if activeConfig.MaxOrphanTxs < 0 {
		str := "%s: The maxorphantx option may not be less than 0 " +
			"-- parsed [%d]"
		err := errors.Errorf(str, funcName, activeConfig.MaxOrphanTxs)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}

	// Look for illegal characters in the user agent comments.
	for _, uaComment := range activeConfig.UserAgentComments {
		if strings.ContainsAny(uaComment, "/:()") {
			err := errors.Errorf("%s: The following characters must not "+
				"appear in user agent comments: '/', ':', '(', ')'",
				funcName)
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(os.Stderr, usageMessage)
			return nil, nil, err
		}
	}

	// --txindex and --droptxindex do not mix.
	if activeConfig.TxIndex && activeConfig.DropTxIndex {
		err := errors.Errorf("%s: the --txindex and --droptxindex "+
			"options may  not be activated at the same time",
			funcName)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}

	// --addrindex and --dropaddrindex do not mix.
	if activeConfig.AddrIndex && activeConfig.DropAddrIndex {
		err := errors.Errorf("%s: the --addrindex and --dropaddrindex "+
			"options may not be activated at the same time",
			funcName)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}

	// --addrindex and --droptxindex do not mix.
	if activeConfig.AddrIndex && activeConfig.DropTxIndex {
		err := errors.Errorf("%s: the --addrindex and --droptxindex "+
			"options may not be activated at the same time "+
			"because the address index relies on the transaction "+
			"index",
			funcName)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}

	// --acceptanceindex and --dropacceptanceindex do not mix.
	if activeConfig.AcceptanceIndex && activeConfig.DropAcceptanceIndex {
		err := errors.Errorf("%s: the --acceptanceindex and --dropacceptanceindex "+
			"options may not be activated at the same time",
			funcName)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}

	// Check mining addresses are valid and saved parsed versions.
	activeConfig.MiningAddrs = make([]util.Address, 0, len(activeConfig.Flags.MiningAddrs))
	for _, strAddr := range activeConfig.Flags.MiningAddrs {
		addr, err := util.DecodeAddress(strAddr, activeConfig.NetParams().Prefix)
		if err != nil {
			str := "%s: mining address '%s' failed to decode: %s"
			err := errors.Errorf(str, funcName, strAddr, err)
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(os.Stderr, usageMessage)
			return nil, nil, err
		}
		if !addr.IsForPrefix(activeConfig.NetParams().Prefix) {
			str := "%s: mining address '%s' is on the wrong network"
			err := errors.Errorf(str, funcName, strAddr)
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(os.Stderr, usageMessage)
			return nil, nil, err
		}
		activeConfig.MiningAddrs = append(activeConfig.MiningAddrs, addr)
	}

	if activeConfig.Flags.Subnetwork != "" {
		activeConfig.SubnetworkID, err = subnetworkid.NewFromStr(activeConfig.Flags.Subnetwork)
		if err != nil {
			return nil, nil, err
		}
	} else {
		activeConfig.SubnetworkID = nil
	}

	// Check that 'generate' and 'subnetwork' flags do not conflict
	if activeConfig.Generate && activeConfig.SubnetworkID != nil {
		str := "%s: both generate flag and subnetwork filtering are set "
		err := errors.Errorf(str, funcName)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}

	// Ensure there is at least one mining address when the generate flag is
	// set.
	if activeConfig.Generate && len(activeConfig.MiningAddrs) == 0 {
		str := "%s: the generate flag is set, but there are no mining " +
			"addresses specified "
		err := errors.Errorf(str, funcName)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}

	// Add default port to all listener addresses if needed and remove
	// duplicate addresses.
	activeConfig.Listeners = network.NormalizeAddresses(activeConfig.Listeners,
		activeConfig.NetParams().DefaultPort)

	// Add default port to all rpc listener addresses if needed and remove
	// duplicate addresses.
	activeConfig.RPCListeners = network.NormalizeAddresses(activeConfig.RPCListeners,
		activeConfig.NetParams().RPCPort)

	// Only allow TLS to be disabled if the RPC is bound to localhost
	// addresses.
	if !activeConfig.DisableRPC && activeConfig.DisableTLS {
		allowedTLSListeners := map[string]struct{}{
			"localhost": {},
			"127.0.0.1": {},
			"::1":       {},
		}
		for _, addr := range activeConfig.RPCListeners {
			host, _, err := net.SplitHostPort(addr)
			if err != nil {
				str := "%s: RPC listen interface '%s' is " +
					"invalid: %s"
				err := errors.Errorf(str, funcName, addr, err)
				fmt.Fprintln(os.Stderr, err)
				fmt.Fprintln(os.Stderr, usageMessage)
				return nil, nil, err
			}
			if _, ok := allowedTLSListeners[host]; !ok {
				str := "%s: the --notls option may not be used " +
					"when binding RPC to non localhost " +
					"addresses: %s"
				err := errors.Errorf(str, funcName, addr)
				fmt.Fprintln(os.Stderr, err)
				fmt.Fprintln(os.Stderr, usageMessage)
				return nil, nil, err
			}
		}
	}

	// Add default port to all added peer addresses if needed and remove
	// duplicate addresses.
	activeConfig.AddPeers = network.NormalizeAddresses(activeConfig.AddPeers,
		activeConfig.NetParams().DefaultPort)
	activeConfig.ConnectPeers = network.NormalizeAddresses(activeConfig.ConnectPeers,
		activeConfig.NetParams().DefaultPort)

	// --noonion and --onion do not mix.
	if activeConfig.NoOnion && activeConfig.OnionProxy != "" {
		err := errors.Errorf("%s: the --noonion and --onion options may "+
			"not be activated at the same time", funcName)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}

	// Tor stream isolation requires either proxy or onion proxy to be set.
	if activeConfig.TorIsolation && activeConfig.Proxy == "" && activeConfig.OnionProxy == "" {
		str := "%s: Tor stream isolation requires either proxy or " +
			"onionproxy to be set"
		err := errors.Errorf(str, funcName)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}

	// Setup dial and DNS resolution (lookup) functions depending on the
	// specified options. The default is to use the standard
	// net.DialTimeout function as well as the system DNS resolver. When a
	// proxy is specified, the dial function is set to the proxy specific
	// dial function and the lookup is set to use tor (unless --noonion is
	// specified in which case the system DNS resolver is used).
	activeConfig.Dial = net.DialTimeout
	activeConfig.Lookup = net.LookupIP
	if activeConfig.Proxy != "" {
		_, _, err := net.SplitHostPort(activeConfig.Proxy)
		if err != nil {
			str := "%s: Proxy address '%s' is invalid: %s"
			err := errors.Errorf(str, funcName, activeConfig.Proxy, err)
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(os.Stderr, usageMessage)
			return nil, nil, err
		}

		// Tor isolation flag means proxy credentials will be overridden
		// unless there is also an onion proxy configured in which case
		// that one will be overridden.
		torIsolation := false
		if activeConfig.TorIsolation && activeConfig.OnionProxy == "" &&
			(activeConfig.ProxyUser != "" || activeConfig.ProxyPass != "") {

			torIsolation = true
			fmt.Fprintln(os.Stderr, "Tor isolation set -- "+
				"overriding specified proxy user credentials")
		}

		proxy := &socks.Proxy{
			Addr:         activeConfig.Proxy,
			Username:     activeConfig.ProxyUser,
			Password:     activeConfig.ProxyPass,
			TorIsolation: torIsolation,
		}
		activeConfig.Dial = proxy.DialTimeout

		// Treat the proxy as tor and perform DNS resolution through it
		// unless the --noonion flag is set or there is an
		// onion-specific proxy configured.
		if !activeConfig.NoOnion && activeConfig.OnionProxy == "" {
			activeConfig.Lookup = func(host string) ([]net.IP, error) {
				return network.TorLookupIP(host, activeConfig.Proxy)
			}
		}
	}

	// Setup onion address dial function depending on the specified options.
	// The default is to use the same dial function selected above. However,
	// when an onion-specific proxy is specified, the onion address dial
	// function is set to use the onion-specific proxy while leaving the
	// normal dial function as selected above. This allows .onion address
	// traffic to be routed through a different proxy than normal traffic.
	if activeConfig.OnionProxy != "" {
		_, _, err := net.SplitHostPort(activeConfig.OnionProxy)
		if err != nil {
			str := "%s: Onion proxy address '%s' is invalid: %s"
			err := errors.Errorf(str, funcName, activeConfig.OnionProxy, err)
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(os.Stderr, usageMessage)
			return nil, nil, err
		}

		// Tor isolation flag means onion proxy credentials will be
		// overridden.
		if activeConfig.TorIsolation &&
			(activeConfig.OnionProxyUser != "" || activeConfig.OnionProxyPass != "") {
			fmt.Fprintln(os.Stderr, "Tor isolation set -- "+
				"overriding specified onionproxy user "+
				"credentials ")
		}

		activeConfig.OnionDial = func(network, addr string, timeout time.Duration) (net.Conn, error) {
			proxy := &socks.Proxy{
				Addr:         activeConfig.OnionProxy,
				Username:     activeConfig.OnionProxyUser,
				Password:     activeConfig.OnionProxyPass,
				TorIsolation: activeConfig.TorIsolation,
			}
			return proxy.DialTimeout(network, addr, timeout)
		}

		// When configured in bridge mode (both --onion and --proxy are
		// configured), it means that the proxy configured by --proxy is
		// not a tor proxy, so override the DNS resolution to use the
		// onion-specific proxy.
		if activeConfig.Proxy != "" {
			activeConfig.Lookup = func(host string) ([]net.IP, error) {
				return network.TorLookupIP(host, activeConfig.OnionProxy)
			}
		}
	} else {
		activeConfig.OnionDial = activeConfig.Dial
	}

	// Specifying --noonion means the onion address dial function results in
	// an error.
	if activeConfig.NoOnion {
		activeConfig.OnionDial = func(a, b string, t time.Duration) (net.Conn, error) {
			return nil, errors.New("tor has been disabled")
		}
	}

	// Warn about missing config file only after all other configuration is
	// done. This prevents the warning on help messages and invalid
	// options. Note this should go directly before the return.
	if configFileError != nil {
		log.Warnf("%s", configFileError)
	}

	return activeConfig, remainingArgs, nil
}

// createDefaultConfig copies the file sample-kaspad.conf to the given destination path,
// and populates it with some randomly generated RPC username and password.
func createDefaultConfigFile(destinationPath string) error {
	// Create the destination directory if it does not exists
	err := os.MkdirAll(filepath.Dir(destinationPath), 0700)
	if err != nil {
		return err
	}

	// We assume sample config file path is same as binary
	path, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return err
	}
	sampleConfigPath := filepath.Join(path, sampleConfigFilename)

	// We generate a random user and password
	randomBytes := make([]byte, 20)
	_, err = rand.Read(randomBytes)
	if err != nil {
		return err
	}
	generatedRPCUser := base64.StdEncoding.EncodeToString(randomBytes)

	_, err = rand.Read(randomBytes)
	if err != nil {
		return err
	}
	generatedRPCPass := base64.StdEncoding.EncodeToString(randomBytes)

	src, err := os.Open(sampleConfigPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dest, err := os.OpenFile(destinationPath,
		os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer dest.Close()

	// We copy every line from the sample config file to the destination,
	// only replacing the two lines for rpcuser and rpcpass
	reader := bufio.NewReader(src)
	for err != io.EOF {
		var line string
		line, err = reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return err
		}

		if strings.Contains(line, "rpcuser=") {
			line = "rpcuser=" + generatedRPCUser + "\n"
		} else if strings.Contains(line, "rpcpass=") {
			line = "rpcpass=" + generatedRPCPass + "\n"
		}

		if _, err := dest.WriteString(line); err != nil {
			return err
		}
	}

	return nil
}
