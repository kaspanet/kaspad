// Copyright (c) 2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package dagconfig

import "testing"

// TestInvalidHashStr ensures the newShaHashFromStr function panics when used to
// with an invalid hash string.
func TestInvalidHashStr(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for invalid hash, got nil")
		}
	}()
	newHashFromStr("banana")
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

func TestParsePrefix(t *testing.T) {
	tests := []struct {
		prefixStr string
		prefix    Bech32Prefix
		isError   bool
	}{
		{"dagcoin", DagCoin, false},
		{"dagreg", DagReg, false},
		{"dagtest", DagTest, false},
		{"dagsim", DagSim, false},
		{"blabla", Unknown, true},
		{"unknown", Unknown, true},
		{"", Unknown, true},
	}

	for _, test := range tests {
		result, err := ParsePrefix(test.prefixStr)
		if (err != nil) != test.isError {
			t.Errorf("TestParsePrefix: %s: expected error status: %t, but got %t",
				test.prefixStr, test.isError, (err != nil))
		}

		if result != test.prefix {
			t.Errorf("TestParsePrefix: %s: expected prefix: %d, but got %d",
				test.prefixStr, test.prefix, result)
		}
	}
}

func TestPrefixToString(t *testing.T) {
	tests := []struct {
		prefix    Bech32Prefix
		prefixStr string
	}{
		{DagCoin, "dagcoin"},
		{DagReg, "dagreg"},
		{DagTest, "dagtest"},
		{DagSim, "dagsim"},
		{Unknown, ""},
	}

	for _, test := range tests {
		result := test.prefix.String()

		if result != test.prefixStr {
			t.Errorf("TestPrefixToString: %s: expected string: %s, but got %s",
				test.prefix, test.prefixStr, result)
		}
	}
}

func TestDNSSeedToString(t *testing.T) {
	host := "test.dns.seed.com"
	seed := DNSSeed{HasFiltering: false, Host: host}

	result := seed.String()
	if result != host {
		t.Errorf("TestDNSSeedToString: Expected: %s, but got: %s", host, result)
	}
}
