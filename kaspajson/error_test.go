// Copyright (c) 2014 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package kaspajson_test

import (
	"testing"

	"github.com/kaspanet/kaspad/kaspajson"
)

// TestErrorCodeStringer tests the stringized output for the ErrorCode type.
func TestErrorCodeStringer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   kaspajson.ErrorCode
		want string
	}{
		{kaspajson.ErrDuplicateMethod, "ErrDuplicateMethod"},
		{kaspajson.ErrInvalidUsageFlags, "ErrInvalidUsageFlags"},
		{kaspajson.ErrInvalidType, "ErrInvalidType"},
		{kaspajson.ErrEmbeddedType, "ErrEmbeddedType"},
		{kaspajson.ErrUnexportedField, "ErrUnexportedField"},
		{kaspajson.ErrUnsupportedFieldType, "ErrUnsupportedFieldType"},
		{kaspajson.ErrNonOptionalField, "ErrNonOptionalField"},
		{kaspajson.ErrNonOptionalDefault, "ErrNonOptionalDefault"},
		{kaspajson.ErrMismatchedDefault, "ErrMismatchedDefault"},
		{kaspajson.ErrUnregisteredMethod, "ErrUnregisteredMethod"},
		{kaspajson.ErrNumParams, "ErrNumParams"},
		{kaspajson.ErrMissingDescription, "ErrMissingDescription"},
		{0xffff, "Unknown ErrorCode (65535)"},
	}

	// Detect additional error codes that don't have the stringer added.
	if len(tests)-1 != int(kaspajson.TstNumErrorCodes) {
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
		in   kaspajson.Error
		want string
	}{
		{
			kaspajson.Error{Description: "some error"},
			"some error",
		},
		{
			kaspajson.Error{Description: "human-readable error"},
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
