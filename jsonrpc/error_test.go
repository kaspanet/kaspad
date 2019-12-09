// Copyright (c) 2014 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package jsonrpc_test

import (
	"testing"

	"github.com/kaspanet/kaspad/jsonrpc"
)

// TestErrorCodeStringer tests the stringized output for the ErrorCode type.
func TestErrorCodeStringer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   jsonrpc.ErrorCode
		want string
	}{
		{jsonrpc.ErrDuplicateMethod, "ErrDuplicateMethod"},
		{jsonrpc.ErrInvalidUsageFlags, "ErrInvalidUsageFlags"},
		{jsonrpc.ErrInvalidType, "ErrInvalidType"},
		{jsonrpc.ErrEmbeddedType, "ErrEmbeddedType"},
		{jsonrpc.ErrUnexportedField, "ErrUnexportedField"},
		{jsonrpc.ErrUnsupportedFieldType, "ErrUnsupportedFieldType"},
		{jsonrpc.ErrNonOptionalField, "ErrNonOptionalField"},
		{jsonrpc.ErrNonOptionalDefault, "ErrNonOptionalDefault"},
		{jsonrpc.ErrMismatchedDefault, "ErrMismatchedDefault"},
		{jsonrpc.ErrUnregisteredMethod, "ErrUnregisteredMethod"},
		{jsonrpc.ErrNumParams, "ErrNumParams"},
		{jsonrpc.ErrMissingDescription, "ErrMissingDescription"},
		{0xffff, "Unknown ErrorCode (65535)"},
	}

	// Detect additional error codes that don't have the stringer added.
	if len(tests)-1 != int(jsonrpc.TstNumErrorCodes) {
		t.Errorf("It appears an error code was added without adding an " +
			"associated stringer test")
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		result := test.in.String()
		if result != test.want {
			t.Errorf("String #%d\n got: %s want: %s", i, result,
				test.want)
			continue
		}
	}
}

// TestError tests the error output for the Error type.
func TestError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   jsonrpc.Error
		want string
	}{
		{
			jsonrpc.Error{Description: "some error"},
			"some error",
		},
		{
			jsonrpc.Error{Description: "human-readable error"},
			"human-readable error",
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		result := test.in.Error()
		if result != test.want {
			t.Errorf("Error #%d\n got: %s want: %s", i, result,
				test.want)
			continue
		}
	}
}
