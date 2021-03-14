package logger

import (
	"fmt"
	"github.com/jrick/logrotate/rotator"
	"github.com/pkg/errors"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
)

const normalLogSize = 512

// defaultFlags specifies changes to the default logger behavior. It is set
// during package init and configured using the LOGFLAGS environment variable.
// New logger backends can override these default flags using WithFlags.
// We're using this instead of `init()` function because variables are initialized before init functions,
// and this variable is used inside other variable intializations, so runs before them
var defaultFlags = getDefaultFlags()

// Flags to modify Backend's behavior.
const (
	// LogFlagLongFile modifies the logger output to include full path and line number
	// of the logging callsite, e.g. /a/b/c/main.go:123.
	LogFlagLongFile uint32 = 1 << iota

	// LogFlagShortFile modifies the logger output to include filename and line number
	// of the logging callsite, e.g. main.go:123. takes precedence over LogFlagLongFile.
	LogFlagShortFile
)

// Read logger flags from the LOGFLAGS environment variable. Multiple flags can
// be set at once, separated by commas.
func getDefaultFlags() (flags uint32) {
	for _, f := range strings.Split(os.Getenv("LOGFLAGS"), ",") {
		switch f {
		case "longfile":
			flags |= LogFlagLongFile
		case "shortfile":
			flags |= LogFlagShortFile
		}
	}
	return
}

const logsBuffer = 0

// Backend is a logging backend. Subsystems created from the backend write to
// the backend's Writer. Backend provides atomic writes to the Writer from all
// subsystems.
type Backend struct {
	flag      uint32
	isRunning uint32
	writers   []logWriter
	writeChan chan logEntry
	syncClose sync.Mutex // used to sync that the logger finished writing everything
}

// NewBackendWithFlags configures a Backend to use the specified flags rather than using
// the package's defaults as determined through the LOGFLAGS environment
// variable.
func NewBackendWithFlags(flags uint32) *Backend {
	return &Backend{flag: flags, writeChan: make(chan logEntry, logsBuffer)}
}

// NewBackend creates a new logger backend.
func NewBackend() *Backend {
	return NewBackendWithFlags(defaultFlags)
}

const (
	defaultThresholdKB = 100 * 1000 // 100 MB logs by default.
	defaultMaxRolls    = 8          // keep 8 last logs by default.
)

type logWriter interface {
	io.WriteCloser
	LogLevel() Level
}

type logWriterWrap struct {
	io.WriteCloser
	logLevel Level
}

func (lw logWriterWrap) LogLevel() Level {
	return lw.logLevel
}

// AddLogFile adds a file which the log will write into on a certain
// log level with the default log rotation settings. It'll create the file if it doesn't exist.
func (b *Backend) AddLogFile(logFile string, logLevel Level) error {
	return b.AddLogFileWithCustomRotator(logFile, logLevel, defaultThresholdKB, defaultMaxRolls)
}

// AddLogWriter adds a type implementing io.WriteCloser which the log will write into on a certain
// log level with the default log rotation settings. It'll create the file if it doesn't exist.
func (b *Backend) AddLogWriter(logWriter io.WriteCloser, logLevel Level) error {
	if b.IsRunning() {
		return errors.New("The logger is already running")
	}
	b.writers = append(b.writers, logWriterWrap{
		WriteCloser: logWriter,
		logLevel:    logLevel,
	})
	return nil
}

// AddLogFileWithCustomRotator adds a file which the log will write into on a certain
// log level, with the specified log rotation settings.
// It'll create the file if it doesn't exist.
func (b *Backend) AddLogFileWithCustomRotator(logFile string, logLevel Level, thresholdKB int64, maxRolls int) error {
	if b.IsRunning() {
		return errors.New("The logger is already running")
	}
	logDir, _ := filepath.Split(logFile)
	// if the logDir is empty then `logFile` is in the cwd and there's no need to create any directory.
	if logDir != "" {
		err := os.MkdirAll(logDir, 0700)
		if err != nil {
			return errors.Errorf("failed to create log directory: %+v", err)
		}
	}
	r, err := rotator.New(logFile, thresholdKB, false, maxRolls)
	if err != nil {
		return errors.Errorf("failed to create file rotator: %s", err)
	}
	b.writers = append(b.writers, logWriterWrap{
		WriteCloser: r,
		logLevel:    logLevel,
	})
	return nil
}

// Run launches the logger backend in a separate go-routine. should only be called once.
func (b *Backend) Run() error {
	if !atomic.CompareAndSwapUint32(&b.isRunning, 0, 1) {
		return errors.New("The logger is already running")
	}
	go func() {
		defer func() {
			if err := recover(); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Fatal error in logger.Backend goroutine: %+v\n", err)
				_, _ = fmt.Fprintf(os.Stderr, "Goroutine stacktrace: %s\n", debug.Stack())
			}
		}()
		b.runBlocking()
	}()
	return nil
}

func (b *Backend) runBlocking() {
	defer atomic.StoreUint32(&b.isRunning, 0)
	b.syncClose.Lock()
	defer b.syncClose.Unlock()

	for log := range b.writeChan {
		for _, writer := range b.writers {
			if log.level >= writer.LogLevel() {
				_, _ = writer.Write(log.log)
			}
		}
	}
}

// IsRunning returns true if backend.Run() has been called and false if it hasn't.
func (b *Backend) IsRunning() bool {
	return atomic.LoadUint32(&b.isRunning) != 0
}

// Close finalizes all log rotators for this backend
func (b *Backend) Close() {
	close(b.writeChan)
	// Wait for it to finish writing using the syncClose mutex.
	b.syncClose.Lock()
	defer b.syncClose.Unlock()
	for _, writer := range b.writers {
		_ = writer.Close()
	}
}

// Logger returns a new logger for a particular subsystem that writes to the
// Backend b. A tag describes the subsystem and is included in all log
// messages. The logger uses the info verbosity level by default.
func (b *Backend) Logger(subsystemTag string) *Logger {
	return &Logger{LevelOff, subsystemTag, b, b.writeChan}
}
