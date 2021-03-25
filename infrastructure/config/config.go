// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package config

import (
	// _ "embed" is necessary for the go:embed feature.
	_ "embed"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/btcsuite/go-socks/socks"
	"github.com/jessevdk/go-flags"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/network"
	"github.com/kaspanet/kaspad/version"
	"github.com/pkg/errors"
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
	defaultBlockMaxMass          = 10000000
	blockMaxMassMin              = 1000
	blockMaxMassMax              = 10000000
	defaultMinRelayTxFee         = 1e-5 // 1 sompi per byte
	defaultMaxOrphanTransactions = 100
	//DefaultMaxOrphanTxSize is the default maximum size for an orphan transaction
	DefaultMaxOrphanTxSize  = 100000
	defaultSigCacheMaxSize  = 100000
	sampleConfigFilename    = "sample-kaspad.conf"
	defaultMaxUTXOCacheSize = 5000000000
)

var (
	// DefaultAppDir is the default home directory for kaspad.
	DefaultAppDir = util.AppDir("kaspad", false)

	defaultConfigFile  = filepath.Join(DefaultAppDir, defaultConfigFilename)
	defaultDataDir     = filepath.Join(DefaultAppDir)
	defaultRPCKeyFile  = filepath.Join(DefaultAppDir, "rpc.key")
	defaultRPCCertFile = filepath.Join(DefaultAppDir, "rpc.cert")
)

//go:embed sample-kaspad.conf
var configurationSampleKaspadString string

// RunServiceCommand is only set to a real function on Windows. It is used
// to parse and execute service commands specified via the -s flag.
var RunServiceCommand func(string) error

// Flags defines the configuration options for kaspad.
//
// See loadConfig for details on the configuration load process.
type Flags struct {
	ShowVersion          bool          `short:"V" long:"version" description:"Display version information and exit"`
	ConfigFile           string        `short:"C" long:"configfile" description:"Path to configuration file"`
	AppDir               string        `short:"b" long:"appdir" description:"Directory to store data"`
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
	RPCListeners         []string      `long:"rpclisten" description:"Add an interface/port to listen for RPC connections (default port: 16110, testnet: 16210)"`
	RPCCert              string        `long:"rpccert" description:"File containing the certificate file"`
	RPCKey               string        `long:"rpckey" description:"File containing the certificate key"`
	RPCMaxClients        int           `long:"rpcmaxclients" description:"Max number of RPC clients for standard connections"`
	RPCMaxWebsockets     int           `long:"rpcmaxwebsockets" description:"Max number of RPC websocket connections"`
	RPCMaxConcurrentReqs int           `long:"rpcmaxconcurrentreqs" description:"Max number of concurrent RPC requests that may be processed concurrently"`
	DisableRPC           bool          `long:"norpc" description:"Disable built-in RPC server"`
	DisableDNSSeed       bool          `long:"nodnsseed" description:"Disable DNS seeding for peers"`
	DNSSeed              string        `long:"dnsseed" description:"Override DNS seeds with specified hostname (Only 1 hostname allowed)"`
	GRPCSeed             string        `long:"grpcseed" description:"Hostname of gRPC server for seeding peers"`
	ExternalIPs          []string      `long:"externalip" description:"Add an ip to the list of local addresses we claim to listen on to peers"`
	Proxy                string        `long:"proxy" description:"Connect via SOCKS5 proxy (eg. 127.0.0.1:9050)"`
	ProxyUser            string        `long:"proxyuser" description:"Username for proxy server"`
	ProxyPass            string        `long:"proxypass" default-mask:"-" description:"Password for proxy server"`
	DbType               string        `long:"dbtype" description:"Database backend to use for the Block DAG"`
	Profile              string        `long:"profile" description:"Enable HTTP profiling on given port -- NOTE port must be between 1024 and 65536"`
	LogLevel             string        `short:"d" long:"loglevel" description:"Logging level for all subsystems {trace, debug, info, warn, error, critical} -- You may also specify <subsystem>=<level>,<subsystem2>=<level>,... to set the log level for individual subsystems -- Use show to list available subsystems"`
	Upnp                 bool          `long:"upnp" description:"Use UPnP to map our listening port outside of NAT"`
	MinRelayTxFee        float64       `long:"minrelaytxfee" description:"The minimum transaction fee in KAS/kB to be considered a non-zero fee."`
	MaxOrphanTxs         int           `long:"maxorphantx" description:"Max number of orphan transactions to keep in memory"`
	BlockMaxMass         uint64        `long:"blockmaxmass" description:"Maximum transaction mass to be used when creating a block"`
	UserAgentComments    []string      `long:"uacomment" description:"Comment to add to the user agent -- See BIP 14 for more information."`
	NoPeerBloomFilters   bool          `long:"nopeerbloomfilters" description:"Disable bloom filtering support"`
	SigCacheMaxSize      uint          `long:"sigcachemaxsize" description:"The maximum number of entries in the signature verification cache"`
	BlocksOnly           bool          `long:"blocksonly" description:"Do not accept transactions from remote peers."`
	RelayNonStd          bool          `long:"relaynonstd" description:"Relay non-standard transactions regardless of the default settings for the active network."`
	RejectNonStd         bool          `long:"rejectnonstd" description:"Reject non-standard transactions regardless of the default settings for the active network."`
	ResetDatabase        bool          `long:"reset-db" description:"Reset database before starting node. It's needed when switching between subnetworks."`
	MaxUTXOCacheSize     uint64        `long:"maxutxocachesize" description:"Max size of loaded UTXO into ram from the disk in bytes"`
	UTXOIndex            bool          `long:"utxoindex" description:"Enable the UTXO index"`
	IsArchivalNode       bool          `long:"archival" description:"Run as an archival node: don't delete old block data when moving the pruning point (Warning: heavy disk usage)'"`
	NetworkFlags
	ServiceOptions *ServiceOptions
}

// Config defines the configuration options for kaspad.
//
// See loadConfig for details on the configuration load process.
type Config struct {
	*Flags
	Lookup        func(string) ([]net.IP, error)
	Dial          func(string, string, time.Duration) (net.Conn, error)
	MiningAddrs   []util.Address
	MinRelayTxFee util.Amount
	Whitelists    []*net.IPNet
	SubnetworkID  *externalapi.DomainSubnetworkID // nil in full nodes
}

// ServiceOptions defines the configuration options for the daemon as a service on
// Windows.
type ServiceOptions struct {
	ServiceCommand string `short:"s" long:"service" description:"Service command {install, remove, start, stop}"`
}

// cleanAndExpandPath expands environment variables and leading ~ in the
// passed path, cleans the result, and returns it.
func cleanAndExpandPath(path string) string {
	// Expand initial ~ to OS specific home directory.
	if strings.HasPrefix(path, "~") {
		homeDir := filepath.Dir(DefaultAppDir)
		path = strings.Replace(path, "~", homeDir, 1)
	}

	// NOTE: The os.ExpandEnv doesn't work with Windows-style %VARIABLE%,
	// but they variables can still be expanded via POSIX-style $VARIABLE.
	return filepath.Clean(os.ExpandEnv(path))
}

// newConfigParser returns a new command line flags parser.
func newConfigParser(cfgFlags *Flags, options flags.Options) *flags.Parser {
	parser := flags.NewParser(cfgFlags, options)
	if runtime.GOOS == "windows" {
		parser.AddGroup("Service Options", "Service Options", cfgFlags.ServiceOptions)
	}
	return parser
}

func defaultFlags() *Flags {
	return &Flags{
		ConfigFile:           defaultConfigFile,
		LogLevel:             defaultLogLevel,
		TargetOutboundPeers:  defaultTargetOutboundPeers,
		MaxInboundPeers:      defaultMaxInboundPeers,
		BanDuration:          defaultBanDuration,
		BanThreshold:         defaultBanThreshold,
		RPCMaxClients:        defaultMaxRPCClients,
		RPCMaxWebsockets:     defaultMaxRPCWebsockets,
		RPCMaxConcurrentReqs: defaultMaxRPCConcurrentReqs,
		AppDir:               defaultDataDir,
		RPCKey:               defaultRPCKeyFile,
		RPCCert:              defaultRPCCertFile,
		BlockMaxMass:         defaultBlockMaxMass,
		MaxOrphanTxs:         defaultMaxOrphanTransactions,
		SigCacheMaxSize:      defaultSigCacheMaxSize,
		MinRelayTxFee:        defaultMinRelayTxFee,
		MaxUTXOCacheSize:     defaultMaxUTXOCacheSize,
		ServiceOptions:       &ServiceOptions{},
	}
}

// DefaultConfig returns the default kaspad configuration
func DefaultConfig() *Config {
	config := &Config{Flags: defaultFlags()}
	config.NetworkFlags.ActiveNetParams = &dagconfig.MainnetParams
	return config
}

// LoadConfig initializes and parses the config using a config file and command
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
func LoadConfig() (*Config, error) {
	cfgFlags := defaultFlags()

	// Pre-parse the command line options to see if an alternative config
	// file or the version flag was specified. Any errors aside from the
	// help message error can be ignored here since they will be caught by
	// the final parse below.
	preCfg := cfgFlags
	preParser := newConfigParser(preCfg, flags.HelpFlag)
	_, err := preParser.Parse()
	if err != nil {
		var flagsErr *flags.Error
		if ok := errors.As(err, &flagsErr); ok && flagsErr.Type == flags.ErrHelp {
			return nil, err
		}
	}

	appName := filepath.Base(os.Args[0])
	appName = strings.TrimSuffix(appName, filepath.Ext(appName))
	usageMessage := fmt.Sprintf("Use %s -h to show usage", appName)

	// Show the version and exit if the version flag was specified.
	if preCfg.ShowVersion {
		fmt.Println(appName, "version", version.Version())
		os.Exit(0)
	}

	// Load additional config from file.
	var configFileError error
	parser := newConfigParser(cfgFlags, flags.Default)
	cfg := &Config{
		Flags: cfgFlags,
	}
	if !preCfg.Simnet || preCfg.ConfigFile != defaultConfigFile {
		if _, err := os.Stat(preCfg.ConfigFile); os.IsNotExist(err) {
			err := createDefaultConfigFile(preCfg.ConfigFile)
			if err != nil {
				return nil, errors.Wrap(err, "Error creating a default config file")
			}
		}

		err := flags.NewIniParser(parser).ParseFile(preCfg.ConfigFile)
		if err != nil {
			if pErr := &(os.PathError{}); !errors.As(err, &pErr) {
				return nil, errors.Wrapf(err, "Error parsing config file: %s\n\n%s", err, usageMessage)
			}
			configFileError = err
		}
	}

	// Parse command line options again to ensure they take precedence.
	_, err = parser.Parse()
	if err != nil {
		var flagsErr *flags.Error
		if ok := errors.As(err, &flagsErr); !ok || flagsErr.Type != flags.ErrHelp {
			return nil, errors.Wrapf(err, "Error parsing command line arguments: %s\n\n%s", err, usageMessage)
		}
		return nil, err
	}

	// Create the home directory if it doesn't already exist.
	funcName := "loadConfig"
	err = os.MkdirAll(DefaultAppDir, 0700)
	if err != nil {
		// Show a nicer error message if it's because a symlink is
		// linked to a directory that does not exist (probably because
		// it's not mounted).
		var e *os.PathError
		if ok := errors.As(err, &e); ok && os.IsExist(err) {
			if link, lerr := os.Readlink(e.Path); lerr == nil {
				str := "is symlink %s -> %s mounted?"
				err = errors.Errorf(str, e.Path, link)
			}
		}

		str := "%s: Failed to create home directory: %s"
		err := errors.Errorf(str, funcName, err)
		return nil, err
	}

	err = cfg.ResolveNetwork(parser)
	if err != nil {
		return nil, err
	}

	// Set the default policy for relaying non-standard transactions
	// according to the default of the active network. The set
	// configuration value takes precedence over the default value for the
	// selected network.
	relayNonStd := cfg.NetParams().RelayNonStdTxs
	switch {
	case cfg.RelayNonStd && cfg.RejectNonStd:
		str := "%s: rejectnonstd and relaynonstd cannot be used " +
			"together -- choose only one"
		err := errors.Errorf(str, funcName)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, err
	case cfg.RejectNonStd:
		relayNonStd = false
	case cfg.RelayNonStd:
		relayNonStd = true
	}
	cfg.RelayNonStd = relayNonStd

	cfg.AppDir = cleanAndExpandPath(cfg.AppDir)
	// Append the network type to the app directory so it is "namespaced"
	// per network.
	// All data is specific to a network, so namespacing the data directory
	// means each individual piece of serialized data does not have to
	// worry about changing names per network and such.
	cfg.AppDir = filepath.Join(cfg.AppDir, cfg.NetParams().Name)

	// Logs directory is usually under the home directory, unless otherwise specified
	if cfg.LogDir == "" {
		cfg.LogDir = filepath.Join(cfg.AppDir, defaultLogDirname)
	}
	cfg.LogDir = cleanAndExpandPath(cfg.LogDir)

	// Special show command to list supported subsystems and exit.
	if cfg.LogLevel == "show" {
		fmt.Println("Supported subsystems", logger.SupportedSubsystems())
		os.Exit(0)
	}

	// Initialize log rotation. After log rotation has been initialized, the
	// logger variables may be used.
	logger.InitLog(filepath.Join(cfg.LogDir, defaultLogFilename), filepath.Join(cfg.LogDir, defaultErrLogFilename))

	// Parse, validate, and set debug log level(s).
	if err := logger.ParseAndSetLogLevels(cfg.LogLevel); err != nil {
		err := errors.Errorf("%s: %s", funcName, err.Error())
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, err
	}

	// Validate profile port number
	if cfg.Profile != "" {
		profilePort, err := strconv.Atoi(cfg.Profile)
		if err != nil || profilePort < 1024 || profilePort > 65535 {
			str := "%s: The profile port must be between 1024 and 65535"
			err := errors.Errorf(str, funcName)
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(os.Stderr, usageMessage)
			return nil, err
		}
	}

	// Don't allow ban durations that are too short.
	if cfg.BanDuration < time.Second {
		str := "%s: The banduration option may not be less than 1s -- parsed [%s]"
		err := errors.Errorf(str, funcName, cfg.BanDuration)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, err
	}

	// Validate any given whitelisted IP addresses and networks.
	if len(cfg.Whitelists) > 0 {
		var ip net.IP
		cfg.Whitelists = make([]*net.IPNet, 0, len(cfg.Flags.Whitelists))

		for _, addr := range cfg.Flags.Whitelists {
			_, ipnet, err := net.ParseCIDR(addr)
			if err != nil {
				ip = net.ParseIP(addr)
				if ip == nil {
					str := "%s: The whitelist value of '%s' is invalid"
					err = errors.Errorf(str, funcName, addr)
					fmt.Fprintln(os.Stderr, err)
					fmt.Fprintln(os.Stderr, usageMessage)
					return nil, err
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
			cfg.Whitelists = append(cfg.Whitelists, ipnet)
		}
	}

	// --addPeer and --connect do not mix.
	if len(cfg.AddPeers) > 0 && len(cfg.ConnectPeers) > 0 {
		str := "%s: the --addpeer and --connect options can not be " +
			"mixed"
		err := errors.Errorf(str, funcName)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, err
	}

	// --proxy or --connect without --listen disables listening.
	if (cfg.Proxy != "" || len(cfg.ConnectPeers) > 0) &&
		len(cfg.Listeners) == 0 {
		cfg.DisableListen = true
	}

	// ConnectPeers means no DNS seeding and no outbound peers
	if len(cfg.ConnectPeers) > 0 {
		cfg.DisableDNSSeed = true
		cfg.TargetOutboundPeers = 0
	}

	// Add the default listener if none were specified. The default
	// listener is all addresses on the listen port for the network
	// we are to connect to.
	if len(cfg.Listeners) == 0 {
		cfg.Listeners = []string{
			net.JoinHostPort("", cfg.NetParams().DefaultPort),
		}
	}

	if cfg.DisableRPC {
		log.Infof("RPC service is disabled")
	}

	// Add the default RPC listener if none were specified. The default
	// RPC listener is all addresses on the RPC listen port for the
	// network we are to connect to.
	if !cfg.DisableRPC && len(cfg.RPCListeners) == 0 {
		cfg.RPCListeners = []string{
			net.JoinHostPort("", cfg.NetParams().RPCPort),
		}
	}

	if cfg.RPCMaxConcurrentReqs < 0 {
		str := "%s: The rpcmaxwebsocketconcurrentrequests option may " +
			"not be less than 0 -- parsed [%d]"
		err := errors.Errorf(str, funcName, cfg.RPCMaxConcurrentReqs)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, err
	}

	// Validate the the minrelaytxfee.
	cfg.MinRelayTxFee, err = util.NewAmount(cfg.Flags.MinRelayTxFee)
	if err != nil {
		str := "%s: invalid minrelaytxfee: %s"
		err := errors.Errorf(str, funcName, err)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, err
	}

	// Disallow 0 and negative min tx fees.
	if cfg.MinRelayTxFee == 0 {
		str := "%s: The minrelaytxfee option must be greater than 0 -- parsed [%d]"
		err := errors.Errorf(str, funcName, cfg.MinRelayTxFee)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, err
	}

	// Limit the max block mass to a sane value.
	if cfg.BlockMaxMass < blockMaxMassMin || cfg.BlockMaxMass >
		blockMaxMassMax {

		str := "%s: The blockmaxmass option must be in between %d " +
			"and %d -- parsed [%d]"
		err := errors.Errorf(str, funcName, blockMaxMassMin,
			blockMaxMassMax, cfg.BlockMaxMass)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, err
	}

	// Limit the max orphan count to a sane value.
	if cfg.MaxOrphanTxs < 0 {
		str := "%s: The maxorphantx option may not be less than 0 " +
			"-- parsed [%d]"
		err := errors.Errorf(str, funcName, cfg.MaxOrphanTxs)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, err
	}

	// Look for illegal characters in the user agent comments.
	for _, uaComment := range cfg.UserAgentComments {
		if strings.ContainsAny(uaComment, "/:()") {
			err := errors.Errorf("%s: The following characters must not "+
				"appear in user agent comments: '/', ':', '(', ')'",
				funcName)
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(os.Stderr, usageMessage)
			return nil, err
		}
	}

	// Add default port to all listener addresses if needed and remove
	// duplicate addresses.
	cfg.Listeners, err = network.NormalizeAddresses(cfg.Listeners,
		cfg.NetParams().DefaultPort)
	if err != nil {
		return nil, err
	}

	// Add default port to all rpc listener addresses if needed and remove
	// duplicate addresses.
	cfg.RPCListeners, err = network.NormalizeAddresses(cfg.RPCListeners,
		cfg.NetParams().RPCPort)
	if err != nil {
		return nil, err
	}

	// Disallow --addpeer and --connect used together
	if len(cfg.AddPeers) > 0 && len(cfg.ConnectPeers) > 0 {
		str := "%s: --addpeer and --connect can not be used together"
		err := errors.Errorf(str, funcName)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, err
	}

	// Add default port to all added peer addresses if needed and remove
	// duplicate addresses.
	cfg.AddPeers, err = network.NormalizeAddresses(cfg.AddPeers,
		cfg.NetParams().DefaultPort)
	if err != nil {
		return nil, err
	}

	cfg.ConnectPeers, err = network.NormalizeAddresses(cfg.ConnectPeers,
		cfg.NetParams().DefaultPort)
	if err != nil {
		return nil, err
	}

	// Setup dial and DNS resolution (lookup) functions depending on the
	// specified options. The default is to use the standard
	// net.DialTimeout function as well as the system DNS resolver. When a
	// proxy is specified, the dial function is set to the proxy specific
	// dial function.
	cfg.Dial = net.DialTimeout
	cfg.Lookup = net.LookupIP
	if cfg.Proxy != "" {
		_, _, err := net.SplitHostPort(cfg.Proxy)
		if err != nil {
			str := "%s: Proxy address '%s' is invalid: %s"
			err := errors.Errorf(str, funcName, cfg.Proxy, err)
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(os.Stderr, usageMessage)
			return nil, err
		}

		proxy := &socks.Proxy{
			Addr:     cfg.Proxy,
			Username: cfg.ProxyUser,
			Password: cfg.ProxyPass,
		}
		cfg.Dial = proxy.DialTimeout
	}

	// Warn about missing config file only after all other configuration is
	// done. This prevents the warning on help messages and invalid
	// options. Note this should go directly before the return.
	if configFileError != nil {
		log.Warnf("%s", configFileError)
	}

	return cfg, nil
}

// createDefaultConfig copies the file sample-kaspad.conf to the given destination path,
// and populates it with some randomly generated RPC username and password.
func createDefaultConfigFile(destinationPath string) error {
	// Create the destination directory if it does not exists
	err := os.MkdirAll(filepath.Dir(destinationPath), 0700)
	if err != nil {
		return err
	}

	dest, err := os.OpenFile(destinationPath,
		os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = dest.WriteString(configurationSampleKaspadString)

	return err
}
