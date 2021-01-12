// Copyright (c) 2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
//
// Copyright (c) 2009 The Go Authors. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//    * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//    * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//    * Neither the name of Google Inc. nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package logger

import (
	"bytes"
	"fmt"
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/jrick/logrotate/rotator"
)

// defaultFlags specifies changes to the default logger behavior. It is set
// during package init and configured using the LOGFLAGS environment variable.
// New logger backends can override these default flags using WithFlags.
var defaultFlags uint32

// Flags to modify Backend's behavior.
const (
	// Llongfile modifies the logger output to include full path and line number
	// of the logging callsite, e.g. /a/b/c/main.go:123.
	Llongfile uint32 = 1 << iota

	// Lshortfile modifies the logger output to include filename and line number
	// of the logging callsite, e.g. main.go:123. Overrides Llongfile.
	Lshortfile
)

// Read logger flags from the LOGFLAGS environment variable. Multiple flags can
// be set at once, separated by commas.
func init() {
	for _, f := range strings.Split(os.Getenv("LOGFLAGS"), ",") {
		switch f {
		case "longfile":
			defaultFlags |= Llongfile
		case "shortfile":
			defaultFlags |= Lshortfile
		}
	}
}

// Level is the level at which a logger is configured. All messages sent
// to a level which is below the current level are filtered.
type Level uint32

// Level constants.
const (
	LevelTrace Level = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelCritical
	LevelOff
)

// levelStrs defines the human-readable names for each logging level.
var levelStrs = [...]string{"TRC", "DBG", "INF", "WRN", "ERR", "CRT", "OFF"}

// LevelFromString returns a level based on the input string s. If the input
// can't be interpreted as a valid log level, the info level and false is
// returned.
func LevelFromString(s string) (l Level, ok bool) {
	switch strings.ToLower(s) {
	case "trace", "trc":
		return LevelTrace, true
	case "debug", "dbg":
		return LevelDebug, true
	case "info", "inf":
		return LevelInfo, true
	case "warn", "wrn":
		return LevelWarn, true
	case "error", "err":
		return LevelError, true
	case "critical", "crt":
		return LevelCritical, true
	case "off":
		return LevelOff, true
	default:
		return LevelInfo, false
	}
}

// String returns the tag of the logger used in log messages, or "OFF" if
// the level will not produce any log output.
func (l Level) String() string {
	if l >= LevelOff {
		return "OFF"
	}
	return levelStrs[l]
}

// NewBackend creates a new logger backend.
func NewBackend(opts ...BackendOption) *Backend {
	b := &Backend{flag: defaultFlags, toStdout: true}
	for _, o := range opts {
		o(b)
	}
	return b
}

type backendLogRotator struct {
	*rotator.Rotator
	logLevel Level
}

// Backend is a logging backend. Subsystems created from the backend write to
// the backend's Writer. Backend provides atomic writes to the Writer from all
// subsystems.
type Backend struct {
	rotators []*backendLogRotator
	mu       sync.Mutex // ensures atomic writes
	flag     uint32
	toStdout bool
}

// BackendOption is a function used to modify the behavior of a Backend.
type BackendOption func(b *Backend)

// WithFlags configures a Backend to use the specified flags rather than using
// the package's defaults as determined through the LOGFLAGS environment
// variable.
func WithFlags(flags uint32) BackendOption {
	return func(b *Backend) {
		b.flag = flags
	}
}

// bufferPool defines a concurrent safe free list of byte slices used to provide
// temporary buffers for formatting log messages prior to outputting them.
var bufferPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, 0, 120)
		return &b // pointer to slice to avoid boxing alloc
	},
}

// buffer returns a byte slice from the free list. A new buffer is allocated if
// there are not any available on the free list. The returned byte slice should
// be returned to the fee list by using the recycleBuffer function when the
// caller is done with it.
func buffer() *[]byte {
	return bufferPool.Get().(*[]byte)
}

// recycleBuffer puts the provided byte slice, which should have been obtain via
// the buffer function, back on the free list.
func recycleBuffer(b *[]byte) {
	*b = (*b)[:0]
	bufferPool.Put(b)
}

// From stdlib log package.
// Cheap integer to fixed-width decimal ASCII. Give a negative width to avoid
// zero-padding.
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
	err := os.MkdirAll(logDir, 0700)
	if err != nil {
		return errors.Errorf("failed to create log directory: %s", err)
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

	bytebuf := buffer()

	var file string
	var line int
	if b.flag&(Lshortfile|Llongfile) != 0 {
		file, line = callsite(b.flag)
	}

	formatHeader(bytebuf, t, lvl.String(), tag, file, line)
	buf := bytes.NewBuffer(*bytebuf)
	fmt.Fprintln(buf, args...)
	*bytebuf = buf.Bytes()

	b.write(lvl, *bytebuf)

	recycleBuffer(bytebuf)
}

// printf outputs a log message to the writer associated with the backend after
// creating a prefix for the given level and tag according to the formatHeader
// function and formatting the provided arguments according to the given format
// specifier.
func (b *Backend) printf(lvl Level, tag string, format string, args ...interface{}) {
	t := mstime.Now() // get as early as possible

	bytebuf := buffer()

	var file string
	var line int
	if b.flag&(Lshortfile|Llongfile) != 0 {
		file, line = callsite(b.flag)
	}

	formatHeader(bytebuf, t, lvl.String(), tag, file, line)
	buf := bytes.NewBuffer(*bytebuf)
	fmt.Fprintf(buf, format, args...)
	*bytebuf = append(buf.Bytes(), '\n')
	b.write(lvl, *bytebuf)

	recycleBuffer(bytebuf)
}

func (b *Backend) write(lvl Level, bytesToWrite []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.toStdout {
		os.Stdout.Write(bytesToWrite)
	}

	for _, r := range b.rotators {
		if lvl >= r.logLevel {
			r.Write(bytesToWrite)
		}
	}
}

// Close finalizes all log rotators for this backend
func (b *Backend) Close() {
	for _, r := range b.rotators {
		r.Close()
	}
}

// WriteToStdout sets if the backend will print to stdout or not (default: true)
func (b *Backend) WriteToStdout(stdout bool) {
	b.toStdout = stdout
}

// Logger returns a new logger for a particular subsystem that writes to the
// Backend b. A tag describes the subsystem and is included in all log
// messages. The logger uses the info verbosity level by default.
func (b *Backend) Logger(subsystemTag string) *Logger {
	return &Logger{LevelInfo, subsystemTag, b}
}

// Logger is a subsystem logger for a Backend.
type Logger struct {
	lvl Level // atomic
	tag string
	b   *Backend
}

// Trace formats message using the default formats for its operands, prepends
// the prefix as necessary, and writes to log with LevelTrace.
func (l *Logger) Trace(args ...interface{}) {
	l.Write(LevelTrace, args...)
}

// Tracef formats message according to format specifier, prepends the prefix as
// necessary, and writes to log with LevelTrace.
func (l *Logger) Tracef(format string, args ...interface{}) {
	l.Writef(LevelTrace, format, args...)
}

// Debug formats message using the default formats for its operands, prepends
// the prefix as necessary, and writes to log with LevelDebug.
func (l *Logger) Debug(args ...interface{}) {
	l.Write(LevelDebug, args...)
}

// Debugf formats message according to format specifier, prepends the prefix as
// necessary, and writes to log with LevelDebug.
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Writef(LevelDebug, format, args...)
}

// Info formats message using the default formats for its operands, prepends
// the prefix as necessary, and writes to log with LevelInfo.
func (l *Logger) Info(args ...interface{}) {
	l.Write(LevelInfo, args...)
}

// Infof formats message according to format specifier, prepends the prefix as
// necessary, and writes to log with LevelInfo.
func (l *Logger) Infof(format string, args ...interface{}) {
	l.Writef(LevelInfo, format, args...)
}

// Warn formats message using the default formats for its operands, prepends
// the prefix as necessary, and writes to log with LevelWarn.
func (l *Logger) Warn(args ...interface{}) {
	l.Write(LevelWarn, args...)
}

// Warnf formats message according to format specifier, prepends the prefix as
// necessary, and writes to log with LevelWarn.
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Writef(LevelWarn, format, args...)
}

// Error formats message using the default formats for its operands, prepends
// the prefix as necessary, and writes to log with LevelError.
func (l *Logger) Error(args ...interface{}) {
	l.Write(LevelError, args...)
}

// Errorf formats message according to format specifier, prepends the prefix as
// necessary, and writes to log with LevelError.
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Writef(LevelError, format, args...)
}

// Critical formats message using the default formats for its operands, prepends
// the prefix as necessary, and writes to log with LevelCritical.
func (l *Logger) Critical(args ...interface{}) {
	l.Write(LevelCritical, args...)
}

// Criticalf formats message according to format specifier, prepends the prefix
// as necessary, and writes to log with LevelCritical.
func (l *Logger) Criticalf(format string, args ...interface{}) {
	l.Writef(LevelCritical, format, args...)
}

// Write formats message using the default formats for its operands, prepends
// the prefix as necessary, and writes to log with the given logLevel.
func (l *Logger) Write(logLevel Level, args ...interface{}) {
	lvl := l.Level()
	if lvl <= logLevel {
		l.b.print(logLevel, l.tag, args...)
	}
}

// Writef formats message according to format specifier, prepends the prefix
// as necessary, and writes to log with the given logLevel.
func (l *Logger) Writef(logLevel Level, format string, args ...interface{}) {
	lvl := l.Level()
	if lvl <= logLevel {
		l.b.printf(logLevel, l.tag, format, args...)
	}
}

// Level returns the current logging level
func (l *Logger) Level() Level {
	return Level(atomic.LoadUint32((*uint32)(&l.lvl)))
}

// SetLevel changes the logging level to the passed level.
func (l *Logger) SetLevel(level Level) {
	atomic.StoreUint32((*uint32)(&l.lvl), uint32(level))
}

// Backend returns the log backend
func (l *Logger) Backend() *Backend {
	return l.b
}
