// Copyright (c) 2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package dagconfig

import (
	"testing"

	"github.com/daglabs/btcd/dagconfig/daghash"
)

func TestNewHashFromStr(t *testing.T) {
	tests := []struct {
		hexStr        string
		expectedHash  *daghash.Hash
		expectedPanic bool
	}{
		{"banana", nil, true},
		{"0000000000000000000000000000000000000000000000000000000000000000",
			&daghash.Hash{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, false},
		{"0101010101010101010101010101010101010101010101010101010101010101",
			&daghash.Hash{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}, false},
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

			if result.Cmp(test.expectedHash) != 0 {
				t.Errorf("%s: Expected hash: %s, but got %s", test.hexStr, test.expectedHash, result)
			}
		}()
	}
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
	mustRegister(&MainNetParams)
}

func TestDNSSeedToString(t *testing.T) {
	host := "test.dns.seed.com"
	seed := DNSSeed{HasFiltering: false, Host: host}

	result := seed.String()
	if result != host {
		t.Errorf("TestDNSSeedToString: Expected: %s, but got: %s", host, result)
	}
}
