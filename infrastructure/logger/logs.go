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
	"os"
	"strings"
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

type backendLogRotator struct {
	*rotator.Rotator
	logLevel Level
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
