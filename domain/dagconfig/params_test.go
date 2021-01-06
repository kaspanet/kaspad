// Copyright (c) 2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package dagconfig

import (
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

func TestNewHashFromStr(t *testing.T) {
	tests := []struct {
		hexStr        string
		expectedHash  *externalapi.DomainHash
		expectedPanic bool
	}{
		{"banana", nil, true},
		{"0000000000000000000000000000000000000000000000000000000000000000",
			externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}),
			false},
		{"0101010101010101010101010101010101010101010101010101010101010101",
			externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}),
			false},
	}

	for _, test := range tests {
		func() {
			defer func() {
				err := recover()
				if (err != nil) != test.expectedPanic {
					t.Errorf("%s: Expected panic: %t for invalid hash, got %t", test.hexStr, test.expectedPanic, err != nil)
				}
			}()

			result := newHashFromStr(test.hexStr)

			if !result.Equal(test.expectedHash) {
				t.Errorf("%s: Expected hash: %s, but got %s", test.hexStr, test.expectedHash, result)
			}
		}()
	}
}

// newHashFromStr converts the passed big-endian hex string into a externalapi.DomainHash.
// It only differs from the one available in hashes package in that it panics on an error
// since it will only be called from tests.
func newHashFromStr(hexStr string) *externalapi.DomainHash {
	hash, err := externalapi.NewDomainHashFromString(hexStr)
	if err != nil {
		panic(err)
	}
	return hash
}

// TestMustRegisterPanic ensures the mustRegister function panics when used to
// register an invalid network.
func TestMustRegisterPanic(t *testing.T) {
	t.Parallel()

	// Setup a defer to catch the expected panic to ensure it actually
	// paniced.
	defer func() {
		if err := recover(); err == nil {
			t.Error("mustRegister did not panic as expected")
		}
	}()

	// Intentionally try to register duplicate params to force a panic.
	mustRegister(&MainnetParams)
}

// TestSkipProofOfWork ensures all of the hard coded network params don't set SkipProofOfWork as true.
func TestSkipProofOfWork(t *testing.T) {
	allParams := []Params{
		MainnetParams,
		TestnetParams,
		SimnetParams,
		DevnetParams,
	}

	for _, params := range allParams {
		if params.SkipProofOfWork {
			t.Errorf("SkipProofOfWork is enabled for %s. This option should be "+
				"used only for tests.", params.Name)
		}
	}
}
