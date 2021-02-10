package logger

import (
	"bytes"
	"fmt"
	"github.com/jrick/logrotate/rotator"
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
)

const normalLogSize = 512

// NewBackendWithFlags configures a Backend to use the specified flags rather than using
// the package's defaults as determined through the LOGFLAGS environment
// variable.
func NewBackendWithFlags(flags uint32) *Backend {
	return &Backend{flag: flags, stdoutLevel: LevelInfo}
}

// NewBackend creates a new logger backend.
func NewBackend() *Backend {
	return NewBackendWithFlags(defaultFlags)
}

// Backend is a logging backend. Subsystems created from the backend write to
// the backend's Writer. Backend provides atomic writes to the Writer from all
// subsystems.
type Backend struct {
	rotators    []*backendLogRotator
	mu          sync.Mutex // ensures atomic writes
	flag        uint32
	stdoutLevel Level
}

// From stdlib log package.
// Cheap integer to fixed-width decimal ASCII. Give a negative width to avoid zero-padding.
func itoa(buf *[]byte, i int, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

// Appends a header in the default format 'YYYY-MM-DD hh:mm:ss.sss [LVL] TAG: '.
// If either of the Lshortfile or Llongfile flags are specified, the file named
// and line number are included after the tag and before the final colon.
func formatHeader(buf *[]byte, t mstime.Time, lvl, tag string, file string, line int) {
	year, month, day := t.Date()
	hour, min, sec := t.Clock()
	ms := t.Millisecond()

	itoa(buf, year, 4)
	*buf = append(*buf, '-')
	itoa(buf, int(month), 2)
	*buf = append(*buf, '-')
	itoa(buf, day, 2)
	*buf = append(*buf, ' ')
	itoa(buf, hour, 2)
	*buf = append(*buf, ':')
	itoa(buf, min, 2)
	*buf = append(*buf, ':')
	itoa(buf, sec, 2)
	*buf = append(*buf, '.')
	itoa(buf, ms, 3)
	*buf = append(*buf, " ["...)
	*buf = append(*buf, lvl...)
	*buf = append(*buf, "] "...)
	*buf = append(*buf, tag...)
	if file != "" {
		*buf = append(*buf, ' ')
		*buf = append(*buf, file...)
		*buf = append(*buf, ':')
		itoa(buf, line, -1)
	}
	*buf = append(*buf, ": "...)
}

// calldepth is the call depth of the callsite function relative to the
// caller of the subsystem logger. It is used to recover the filename and line
// number of the logging call if either the short or long file flags are
// specified.
const calldepth = 3

// callsite returns the file name and line number of the callsite to the
// subsystem logger.
func callsite(flag uint32) (string, int) {
	_, file, line, ok := runtime.Caller(calldepth)
	if !ok {
		return "???", 0
	}
	if flag&Lshortfile != 0 {
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if os.IsPathSeparator(file[i]) {
				short = file[i+1:]
				break
			}
		}
		file = short
	}
	return file, line
}

const (
	defaultThresholdKB = 100 * 1000 // 100 MB logs by default.
	defaultMaxRolls    = 8          // keep 8 last logs by default.
)

// AddLogFile adds a file which the log will write into on a certain
// log level with the default log rotation settings. It'll create the file if it doesn't exist.
func (b *Backend) AddLogFile(logFile string, logLevel Level) error {
	return b.AddLogFileWithCustomRotator(logFile, logLevel, defaultThresholdKB, defaultMaxRolls)
}

// AddLogFileWithCustomRotator adds a file which the log will write into on a certain
// log level, with the specified log rotation settings.
// It'll create the file if it doesn't exist.
func (b *Backend) AddLogFileWithCustomRotator(logFile string, logLevel Level, thresholdKB int64, maxRolls int) error {
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
	b.rotators = append(b.rotators, &backendLogRotator{
		Rotator:  r,
		logLevel: logLevel,
	})
	return nil
}

// print outputs a log message to the writer associated with the backend after
// creating a prefix for the given level and tag according to the formatHeader
// function and formatting the provided arguments using the default formatting
// rules.
func (b *Backend) print(lvl Level, tag string, args ...interface{}) {
	t := mstime.Now() // get as early as possible

	var file string
	var line int
	if b.flag&(Lshortfile|Llongfile) != 0 {
		file, line = callsite(b.flag)
	}

	buf := make([]byte, 0, normalLogSize)
	formatHeader(&buf, t, lvl.String(), tag, file, line)
	bytesBuf := bytes.NewBuffer(buf)
	_, _ = fmt.Fprintln(bytesBuf, args...)

	b.write(lvl, bytesBuf.Bytes())
}

// printf outputs a log message to the writer associated with the backend after
// creating a prefix for the given level and tag according to the formatHeader
// function and formatting the provided arguments according to the given format
// specifier.
func (b *Backend) printf(lvl Level, tag string, format string, args ...interface{}) {
	t := mstime.Now() // get as early as possible

	var file string
	var line int
	if b.flag&(Lshortfile|Llongfile) != 0 {
		file, line = callsite(b.flag)
	}

	buf := make([]byte, 0, normalLogSize)

	formatHeader(&buf, t, lvl.String(), tag, file, line)
	bytesBuf := bytes.NewBuffer(buf)
	_, _ = fmt.Fprintf(bytesBuf, format, args...)
	bytesBuf.WriteByte('\n')

	b.write(lvl, bytesBuf.Bytes())
}

func (b *Backend) write(lvl Level, bytesToWrite []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if lvl >= b.StdoutLevel() {
		_, _ = os.Stdout.Write(bytesToWrite)
	}

	for _, r := range b.rotators {
		if lvl >= r.logLevel {
			_, _ = r.Write(bytesToWrite)
		}
	}
}

// StdoutLevel returns the current stdout logging level
func (b *Backend) StdoutLevel() Level {
	return Level(atomic.LoadUint32((*uint32)(&b.stdoutLevel)))
}

// SetStdoutLevel changes the logging level to the passed level.
func (b *Backend) SetStdoutLevel(level Level) {
	atomic.StoreUint32((*uint32)(&b.stdoutLevel), uint32(level))
}

// Close finalizes all log rotators for this backend
func (b *Backend) Close() {
	for _, r := range b.rotators {
		_ = r.Close()
	}
}

// Logger returns a new logger for a particular subsystem that writes to the
// Backend b. A tag describes the subsystem and is included in all log
// messages. The logger uses the info verbosity level by default.
func (b *Backend) Logger(subsystemTag string) *Logger {
	return &Logger{LevelInfo, subsystemTag, b}
}
