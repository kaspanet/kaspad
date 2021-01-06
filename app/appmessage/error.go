// Copyright (c) 2013-2015 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package appmessage

import (
	"fmt"
)

// MessageError describes an issue with a message.
// An example of some potential issues are messages from the wrong kaspa
// network, invalid commands, mismatched checksums, and exceeding max payloads.
//
// This provides a mechanism for the caller to type assert the error to
// differentiate between general io errors such as io.EOF and issues that
// resulted from malformed messages.
type MessageError struct {
	Func        string // Function name
	Description string // Human readable description of the issue
}

// Error satisfies the error interface and prints human-readable errors.
func (e *MessageError) Error() string {
	if e.Func != "" {
		return fmt.Sprintf("%s: %s", e.Func, e.Description)
	}
	return e.Description
}

// messageError creates an error for the given function and description.
func messageError(f string, desc string) *MessageError {
	return &MessageError{Func: f, Description: desc}
}

// RPCError represents an error arriving from the RPC
type RPCError struct {
	Message string
}

// RPCErrorf formats according to a format specifier and returns the string
// as an RPCError.
func RPCErrorf(format string, args ...interface{}) *RPCError {
	return &RPCError{
		Message: fmt.Sprintf(format, args...),
	}
}
