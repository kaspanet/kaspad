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
	"sync"

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

	// subsystemLoggers maps each subsystem identifier to its associated logger.
	subsystemLoggers      = make(map[string]*Logger)
	subsystemLoggersMutex sync.Mutex
)

// RegisterSubSystem Registers a new subsystem logger, should be called in a global variable,
// panics if the subsystem is already registered
func RegisterSubSystem(subsystem string) *Logger {
	subsystemLoggersMutex.Lock()
	defer subsystemLoggersMutex.Unlock()
	logger, exists := subsystemLoggers[subsystem]
	if !exists {
		logger = BackendLog.Logger(subsystem)
		subsystemLoggers[subsystem] = logger
	}
	return logger
}

// InitLog attaches log file and error log file to the backend log.
func InitLog(logFile, errLogFile string) {
	// 280 MB (MB=1000^2 bytes)
	err := BackendLog.AddLogFileWithCustomRotator(logFile, LevelTrace, 1000*280, 64)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error adding log file %s as log rotator for level %s: %s", logFile, LevelTrace, err)
		os.Exit(1)
	}
	err = BackendLog.AddLogFile(errLogFile, LevelWarn)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error adding log file %s as log rotator for level %s: %s", errLogFile, LevelWarn, err)
		os.Exit(1)
	}
}

// SetLogLevel sets the logging level for provided subsystem. Invalid
// subsystems are ignored. Uninitialized subsystems are dynamically created as
// needed.
func SetLogLevel(subsystemID string, logLevel string) error {
	// Ignore invalid subsystems.
	logger, ok := subsystemLoggers[subsystemID]
	if !ok {
		return errors.Errorf("'%s' Isn't a valid subsystem", subsystemID)
	}
	// Defaults to info if the log level is invalid.
	level, ok := LevelFromString(logLevel)
	if !ok {
		return errors.Errorf("'%s' Isn't a valid log level", logLevel)
	}

	logger.SetLevel(level)
	return nil
}

// SetLogLevels sets the log level for all subsystem loggers to the passed
// level. It also dynamically creates the subsystem loggers as needed, so it
// can be used to initialize the logging system.
func SetLogLevels(logLevel string) error {
	subsystemLoggersMutex.Lock()
	defer subsystemLoggersMutex.Unlock()
	// Configure all sub-systems with the new logging level. Dynamically
	// create loggers as needed.
	level, ok := LevelFromString(logLevel)
	if !ok {
		return errors.Errorf("'%s' Isn't a valid log level", logLevel)
	}
	for _, logger := range subsystemLoggers {
		logger.SetLevel(level)
	}
	return nil
}

// SupportedSubsystems returns a sorted slice of the supported subsystems for
// logging purposes.
func SupportedSubsystems() []string {
	subsystemLoggersMutex.Lock()
	defer subsystemLoggersMutex.Unlock()
	// Convert the subsystemLoggers map keys to a slice.
	subsystems := make([]string, 0, len(subsystemLoggers))
	for subsysID := range subsystemLoggers {
		subsystems = append(subsystems, subsysID)
	}

	// Sort the subsystems for stable display.
	sort.Strings(subsystems)
	return subsystems
}

func getSubsystem(tag string) (logger *Logger, ok bool) {
	subsystemLoggersMutex.Lock()
	defer subsystemLoggersMutex.Unlock()
	logger, ok = subsystemLoggers[tag]
	return
}

// ParseAndSetLogLevels attempts to parse the specified debug level and set
// the levels accordingly. An appropriate error is returned if anything is
// invalid.
func ParseAndSetLogLevels(logLevel string) error {
	// When the specified string doesn't have any delimters, treat it as
	// the log level for all subsystems.
	if !strings.Contains(logLevel, ",") && !strings.Contains(logLevel, "=") {
		// Validate debug log level.
		if !validLogLevel(logLevel) {
			str := "The specified debug level [%s] is invalid"
			return errors.Errorf(str, logLevel)
		}

		// Change the logging level for all subsystems.
		return SetLogLevels(logLevel)
	}

	// Split the specified string into subsystem/level pairs while detecting
	// issues and update the log levels accordingly.
	for _, logLevelPair := range strings.Split(logLevel, ",") {
		if !strings.Contains(logLevelPair, "=") {
			str := "The specified debug level contains an invalid " +
				"subsystem/level pair [%s]"
			return errors.Errorf(str, logLevelPair)
		}

		// Extract the specified subsystem and log level.
		fields := strings.Split(logLevelPair, "=")
		subsysID, logLevel := fields[0], fields[1]

		// Validate subsystem.
		if _, exists := getSubsystem(subsysID); !exists {
			str := "The specified subsystem [%s] is invalid -- " +
				"supported subsytems %s"
			return errors.Errorf(str, subsysID, strings.Join(SupportedSubsystems(), ", "))
		}

		// Validate log level.
		if !validLogLevel(logLevel) {
			str := "The specified debug level [%s] is invalid"
			return errors.Errorf(str, logLevel)
		}

		err := SetLogLevel(subsysID, logLevel)
		if err != nil {
			return err
		}
	}
	return nil
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
