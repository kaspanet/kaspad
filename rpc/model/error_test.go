// Copyright (c) 2014 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package model_test

import (
	"testing"

	"github.com/kaspanet/kaspad/rpc/model"
)

// TestErrorCodeStringer tests the stringized output for the ErrorCode type.
func TestErrorCodeStringer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   model.ErrorCode
		want string
	}{
		{model.ErrDuplicateMethod, "ErrDuplicateMethod"},
		{model.ErrInvalidUsageFlags, "ErrInvalidUsageFlags"},
		{model.ErrInvalidType, "ErrInvalidType"},
		{model.ErrEmbeddedType, "ErrEmbeddedType"},
		{model.ErrUnexportedField, "ErrUnexportedField"},
		{model.ErrUnsupportedFieldType, "ErrUnsupportedFieldType"},
		{model.ErrNonOptionalField, "ErrNonOptionalField"},
		{model.ErrNonOptionalDefault, "ErrNonOptionalDefault"},
		{model.ErrMismatchedDefault, "ErrMismatchedDefault"},
		{model.ErrUnregisteredMethod, "ErrUnregisteredMethod"},
		{model.ErrNumParams, "ErrNumParams"},
		{model.ErrMissingDescription, "ErrMissingDescription"},
		{0xffff, "Unknown ErrorCode (65535)"},
	}

	// Detect additional error codes that don't have the stringer added.
	if len(tests)-1 != int(model.TstNumErrorCodes) {
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
		in   model.Error
		want string
	}{
		{
			model.Error{Description: "some error"},
			"some error",
		},
		{
			model.Error{Description: "human-readable error"},
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
