// Copyright (c) 2013-2017 The btcsuite developers
// Copyright (c) 2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package logger

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

// Loggers per subsystem. A single backend logger is created and all subsytem
// loggers created from it will write to the backend. When adding new
// subsystems, add the subsystem logger variable here and to the
// subsystemLoggers map.
//
// Loggers can not be used before the log rotator has been initialized with a
// log file. This must be performed early during application startup by calling
// InitLog.
var (
	// BackendLog is the logging backend used to create all subsystem loggers.
	BackendLog = NewBackend()

	adxrLog = BackendLog.Logger("ADXR")
	amgrLog = BackendLog.Logger("AMGR")
	cmgrLog = BackendLog.Logger("CMGR")
	ksdbLog = BackendLog.Logger("KSDB")
	kasdLog = BackendLog.Logger("KASD")
	bdagLog = BackendLog.Logger("BDAG")
	cnfgLog = BackendLog.Logger("CNFG")
	discLog = BackendLog.Logger("DISC")
	indxLog = BackendLog.Logger("INDX")
	minrLog = BackendLog.Logger("MINR")
	peerLog = BackendLog.Logger("PEER")
	rpcsLog = BackendLog.Logger("RPCS")
	rpccLog = BackendLog.Logger("RPCC")
	scrpLog = BackendLog.Logger("SCRP")
	srvrLog = BackendLog.Logger("SRVR")
	syncLog = BackendLog.Logger("SYNC")
	txmpLog = BackendLog.Logger("TXMP")
	utilLog = BackendLog.Logger("UTIL")
	profLog = BackendLog.Logger("PROF")
	protLog = BackendLog.Logger("PROT")
	muxxLog = BackendLog.Logger("MUXX")
	grpcLog = BackendLog.Logger("GRPC")
	p2psLog = BackendLog.Logger("P2PS")
	ntarLog = BackendLog.Logger("NTAR")
	dnssLog = BackendLog.Logger("DNSS")
	snvrLog = BackendLog.Logger("SNVR")
	wsvcLog = BackendLog.Logger("WSVC")
	reacLog = BackendLog.Logger("REAC")
	prnmLog = BackendLog.Logger("PRNM")
	blvlLog = BackendLog.Logger("BLVL")
)

// SubsystemTags is an enum of all sub system tags
var SubsystemTags = struct {
	ADXR,
	AMGR,
	CMGR,
	KSDB,
	KASD,
	BDAG,
	CNFG,
	DISC,
	INDX,
	MINR,
	PEER,
	RPCS,
	RPCC,
	SCRP,
	SRVR,
	SYNC,
	TXMP,
	UTIL,
	PROF,
	PROT,
	MUXX,
	GRPC,
	P2PS,
	NTAR,
	DNSS,
	SNVR,
	WSVC,
	REAC,
	PRNM,
	BLVL string
}{
	ADXR: "ADXR",
	AMGR: "AMGR",
	CMGR: "CMGR",
	KSDB: "KSDB",
	KASD: "KASD",
	BDAG: "BDAG",
	CNFG: "CNFG",
	DISC: "DISC",
	INDX: "INDX",
	MINR: "MINR",
	PEER: "PEER",
	RPCS: "RPCS",
	RPCC: "RPCC",
	SCRP: "SCRP",
	SRVR: "SRVR",
	SYNC: "SYNC",
	TXMP: "TXMP",
	UTIL: "UTIL",
	PROF: "PROF",
	PROT: "PROT",
	MUXX: "MUXX",
	GRPC: "GRPC",
	P2PS: "P2PS",
	NTAR: "NTAR",
	DNSS: "DNSS",
	SNVR: "SNVR",
	WSVC: "WSVC",
	REAC: "REAC",
	PRNM: "PRNM",
	BLVL: "BLVL",
}

// subsystemLoggers maps each subsystem identifier to its associated logger.
var subsystemLoggers = map[string]*Logger{
	SubsystemTags.ADXR: adxrLog,
	SubsystemTags.AMGR: amgrLog,
	SubsystemTags.CMGR: cmgrLog,
	SubsystemTags.KSDB: ksdbLog,
	SubsystemTags.KASD: kasdLog,
	SubsystemTags.BDAG: bdagLog,
	SubsystemTags.CNFG: cnfgLog,
	SubsystemTags.DISC: discLog,
	SubsystemTags.INDX: indxLog,
	SubsystemTags.MINR: minrLog,
	SubsystemTags.PEER: peerLog,
	SubsystemTags.RPCS: rpcsLog,
	SubsystemTags.RPCC: rpccLog,
	SubsystemTags.SCRP: scrpLog,
	SubsystemTags.SRVR: srvrLog,
	SubsystemTags.SYNC: syncLog,
	SubsystemTags.TXMP: txmpLog,
	SubsystemTags.UTIL: utilLog,
	SubsystemTags.PROF: profLog,
	SubsystemTags.PROT: protLog,
	SubsystemTags.MUXX: muxxLog,
	SubsystemTags.GRPC: grpcLog,
	SubsystemTags.P2PS: p2psLog,
	SubsystemTags.NTAR: ntarLog,
	SubsystemTags.DNSS: dnssLog,
	SubsystemTags.SNVR: snvrLog,
	SubsystemTags.WSVC: wsvcLog,
	SubsystemTags.REAC: reacLog,
	SubsystemTags.PRNM: prnmLog,
	SubsystemTags.BLVL: blvlLog,
}

// InitLog attaches log file and error log file to the backend log.
func InitLog(logFile, errLogFile string) {
	// 280 MB (MB=1000^2 bytes)
	err := BackendLog.AddLogFileWithCustomRotator(logFile, LevelTrace, 1000*280, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error adding log file %s as log rotator for level %s: %s", logFile, LevelTrace, err)
		os.Exit(1)
	}
	err = BackendLog.AddLogFile(errLogFile, LevelWarn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error adding log file %s as log rotator for level %s: %s", errLogFile, LevelWarn, err)
		os.Exit(1)
	}
}

// SetLogLevel sets the logging level for provided subsystem. Invalid
// subsystems are ignored. Uninitialized subsystems are dynamically created as
// needed.
func SetLogLevel(subsystemID string, logLevel string) {
	// Ignore invalid subsystems.
	logger, ok := subsystemLoggers[subsystemID]
	if !ok {
		return
	}

	// Defaults to info if the log level is invalid.
	level, _ := LevelFromString(logLevel)
	logger.SetLevel(level)
}

// SetLogLevels sets the log level for all subsystem loggers to the passed
// level. It also dynamically creates the subsystem loggers as needed, so it
// can be used to initialize the logging system.
func SetLogLevels(logLevel string) {
	// Configure all sub-systems with the new logging level. Dynamically
	// create loggers as needed.
	for subsystemID := range subsystemLoggers {
		SetLogLevel(subsystemID, logLevel)
	}
}

// DirectionString is a helper function that returns a string that represents
// the direction of a connection (inbound or outbound).
func DirectionString(inbound bool) string {
	if inbound {
		return "inbound"
	}
	return "outbound"
}

// PickNoun returns the singular or plural form of a noun depending
// on the count n.
func PickNoun(n uint64, singular, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
}

// SupportedSubsystems returns a sorted slice of the supported subsystems for
// logging purposes.
func SupportedSubsystems() []string {
	// Convert the subsystemLoggers map keys to a slice.
	subsystems := make([]string, 0, len(subsystemLoggers))
	for subsysID := range subsystemLoggers {
		subsystems = append(subsystems, subsysID)
	}

	// Sort the subsystems for stable display.
	sort.Strings(subsystems)
	return subsystems
}

// Get returns a logger of a specific sub system
func Get(tag string) (logger *Logger, ok bool) {
	logger, ok = subsystemLoggers[tag]
	return
}

// ParseAndSetDebugLevels attempts to parse the specified debug level and set
// the levels accordingly. An appropriate error is returned if anything is
// invalid.
func ParseAndSetDebugLevels(debugLevel string) error {
	// When the specified string doesn't have any delimters, treat it as
	// the log level for all subsystems.
	if !strings.Contains(debugLevel, ",") && !strings.Contains(debugLevel, "=") {
		// Validate debug log level.
		if !validLogLevel(debugLevel) {
			str := "The specified debug level [%s] is invalid"
			return errors.Errorf(str, debugLevel)
		}

		// Change the logging level for all subsystems.
		SetLogLevels(debugLevel)

		return nil
	}

	// Split the specified string into subsystem/level pairs while detecting
	// issues and update the log levels accordingly.
	for _, logLevelPair := range strings.Split(debugLevel, ",") {
		if !strings.Contains(logLevelPair, "=") {
			str := "The specified debug level contains an invalid " +
				"subsystem/level pair [%s]"
			return errors.Errorf(str, logLevelPair)
		}

		// Extract the specified subsystem and log level.
		fields := strings.Split(logLevelPair, "=")
		subsysID, logLevel := fields[0], fields[1]

		// Validate subsystem.
		if _, exists := Get(subsysID); !exists {
			str := "The specified subsystem [%s] is invalid -- " +
				"supported subsytems %s"
			return errors.Errorf(str, subsysID, strings.Join(SupportedSubsystems(), ", "))
		}

		// Validate log level.
		if !validLogLevel(logLevel) {
			str := "The specified debug level [%s] is invalid"
			return errors.Errorf(str, logLevel)
		}

		SetLogLevel(subsysID, logLevel)
	}

	return nil
}

// validLogLevel returns whether or not logLevel is a valid debug log level.
func validLogLevel(logLevel string) bool {
	switch logLevel {
	case "trace":
		fallthrough
	case "debug":
		fallthrough
	case "info":
		fallthrough
	case "warn":
		fallthrough
	case "error":
		fallthrough
	case "critical":
		return true
	}
	return false
}

// LogClosure is a closure that can be printed with %s to be used to
// generate expensive-to-create data for a detailed log level and avoid doing
// the work if the data isn't printed.
type LogClosure func() string

func (c LogClosure) String() string {
	return c()
}

// NewLogClosure casts a function to a LogClosure.
// See LogClosure for details.
func NewLogClosure(c func() string) LogClosure {
	return c
}
