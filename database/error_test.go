// Copyright (c) 2015-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package database

import (
	"errors"
	"testing"
)

// TestErrorCodeStringer tests the stringized output for the ErrorCode type.
func TestErrorCodeStringer(t *testing.T) {
	tests := []struct {
		in   ErrorCode
		want string
	}{
		{ErrDbTypeRegistered, "ErrDbTypeRegistered"},
		{ErrDbUnknownType, "ErrDbUnknownType"},
		{ErrDbDoesNotExist, "ErrDbDoesNotExist"},
		{ErrDbExists, "ErrDbExists"},
		{ErrDbNotOpen, "ErrDbNotOpen"},
		{ErrDbAlreadyOpen, "ErrDbAlreadyOpen"},
		{ErrInvalid, "ErrInvalid"},
		{ErrCorruption, "ErrCorruption"},
		{ErrTxClosed, "ErrTxClosed"},
		{ErrTxNotWritable, "ErrTxNotWritable"},
		{ErrBucketNotFound, "ErrBucketNotFound"},
		{ErrBucketExists, "ErrBucketExists"},
		{ErrBucketNameRequired, "ErrBucketNameRequired"},
		{ErrKeyRequired, "ErrKeyRequired"},
		{ErrKeyTooLarge, "ErrKeyTooLarge"},
		{ErrValueTooLarge, "ErrValueTooLarge"},
		{ErrIncompatibleValue, "ErrIncompatibleValue"},
		{ErrBlockNotFound, "ErrBlockNotFound"},
		{ErrBlockExists, "ErrBlockExists"},
		{ErrBlockRegionInvalid, "ErrBlockRegionInvalid"},
		{ErrDriverSpecific, "ErrDriverSpecific"},

		{0xffff, "Unknown ErrorCode (65535)"},
	}

	// Detect additional error codes that don't have the stringer added.
	if len(tests)-1 != int(TstNumErrorCodes) {
		t.Errorf("It appears an error code was added without adding " +
			"an associated stringer test")
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		result := test.in.String()
		if result != test.want {
			t.Errorf("String #%d\ngot: %s\nwant: %s", i, result,
				test.want)
			continue
		}
	}
}

// TestError tests the error output for the Error type.
func TestError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   Error
		want string
	}{
		{
			Error{Description: "some error"},
			"some error",
		},
		{
			Error{Description: "human-readable error"},
			"human-readable error",
		},
		{
			Error{
				ErrorCode:   ErrDriverSpecific,
				Description: "some error",
				Err:         errors.New("driver-specific error"),
			},
			"some error: driver-specific error",
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

func TestIsErrorCode(t *testing.T) {
	dummyError := errors.New("")

	tests := []struct {
		err            error
		code           ErrorCode
		expectedResult bool
	}{
		{makeError(ErrBucketExists, "", dummyError), ErrBucketExists, true},
		{makeError(ErrBucketExists, "", dummyError), ErrBlockExists, false},
		{dummyError, ErrBlockExists, false},
		{nil, ErrBlockExists, false},
	}

	for i, test := range tests {
		actualResult := IsErrorCode(test.err, test.code)
		if test.expectedResult != actualResult {
			t.Errorf("TestIsErrorCode: %d: Expected: %t, but got: %t",
				i, test.expectedResult, actualResult)
		}
	}
}
