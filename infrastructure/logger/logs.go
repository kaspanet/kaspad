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
	"os"
	"runtime"
	"sync/atomic"
)

// Logger is a subsystem logger for a Backend.
type Logger struct {
	lvl       Level // atomic
	tag       string
	b         *Backend
	writeChan chan<- logEntry
}

type logEntry struct {
	log   []byte
	level Level
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
		l.print(logLevel, l.tag, args...)
	}
}

// Writef formats message according to format specifier, prepends the prefix
// as necessary, and writes to log with the given logLevel.
func (l *Logger) Writef(logLevel Level, format string, args ...interface{}) {
	lvl := l.Level()
	if lvl <= logLevel {
		l.printf(logLevel, l.tag, format, args...)
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

// printf outputs a log message to the writer associated with the backend after
// creating a prefix for the given level and tag according to the formatHeader
// function and formatting the provided arguments according to the given format
// specifier.
func (l *Logger) printf(lvl Level, tag string, format string, args ...interface{}) {
	t := mstime.Now() // get as early as possible

	var file string
	var line int
	if l.b.flag&(LogFlagShortFile|LogFlagLongFile) != 0 {
		file, line = callsite(l.b.flag)
	}

	buf := make([]byte, 0, normalLogSize)

	formatHeader(&buf, t, lvl.String(), tag, file, line)
	bytesBuf := bytes.NewBuffer(buf)
	_, _ = fmt.Fprintf(bytesBuf, format, args...)
	bytesBuf.WriteByte('\n')

	if !l.b.IsRunning() {
		_, _ = fmt.Fprintf(os.Stderr, bytesBuf.String())
		panic("Writing to the logger when it's not running")
	}
	l.writeChan <- logEntry{bytesBuf.Bytes(), lvl}
}

// print outputs a log message to the writer associated with the backend after
// creating a prefix for the given level and tag according to the formatHeader
// function and formatting the provided arguments using the default formatting
// rules.
func (l *Logger) print(lvl Level, tag string, args ...interface{}) {
	if atomic.LoadUint32(&l.b.isRunning) == 0 {
		panic("printing log without initializing")
	}
	t := mstime.Now() // get as early as possible

	var file string
	var line int
	if l.b.flag&(LogFlagShortFile|LogFlagLongFile) != 0 {
		file, line = callsite(l.b.flag)
	}

	buf := make([]byte, 0, normalLogSize)
	formatHeader(&buf, t, lvl.String(), tag, file, line)
	bytesBuf := bytes.NewBuffer(buf)
	_, _ = fmt.Fprintln(bytesBuf, args...)

	if !l.b.IsRunning() {
		panic("Writing to the logger when it's not running")
	}
	l.writeChan <- logEntry{bytesBuf.Bytes(), lvl}
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
// If either of the LogFlagShortFile or LogFlagLongFile flags are specified, the file named
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
const calldepth = 4

// callsite returns the file name and line number of the callsite to the
// subsystem logger.
func callsite(flag uint32) (string, int) {
	_, file, line, ok := runtime.Caller(calldepth)
	if !ok {
		return "???", 0
	}
	if flag&LogFlagShortFile != 0 {
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
