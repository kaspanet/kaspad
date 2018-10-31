// Copyright (c) 2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package bech32_test

import (
	"github.com/daglabs/btcd/util/bech32"
	"testing"
)

var checkEncodingStringTests = []struct {
	prefix  string
	version byte
	in      string
	out     string
}{
	{"a", 0, "", "a:qqeq69uvrh"},
	{"a", 8, "", "a:pq99546ray"},
	{"a", 120, "", "a:0qf6jrhtdq"},
	{"b", 8, " ", "b:pqsqzsjd64fv"},
	{"b", 8, "-", "b:pqksmhczf8ud"},
	{"b", 8, "0", "b:pqcq53eqrk0e"},
	{"b", 8, "1", "b:pqcshg75y0vf"},
	{"b", 8, "-1", "b:pqknzl4e9y0zy"},
	{"b", 8, "11", "b:pqcnzt888ytdg"},
	{"b", 8, "abc", "b:ppskycc8txxxn2w"},
	{"b", 8, "1234598760", "b:pqcnyve5x5unsdekxqeusxeyu2"},
	{"b", 8, "abcdefghijklmnopqrstuvwxyz", "b:ppskycmyv4nxw6rfdf4kcmtwdac8zunnw36hvamc09aqtpppz8lk"},
	{"b", 8, "000000000000000000000000000000000000000000", "b:pqcrqvpsxqcrqvpsxqcrqvpsxqcrqvpsxqcrqvpsxqcrqvpsxqcrqvpsxqcrqvpsxqcrq7ag684l3"},
}

func TestBech32(t *testing.T) {
	for x, test := range checkEncodingStringTests {
		// test encoding
		encoded := bech32.Encode(test.prefix, []byte(test.in), test.version)
		if encoded != test.out {
			t.Errorf("Encode test #%d failed: got %s, want: %s", x, encoded, test.out)
		}

		// test decoding
		prefix, decoded, version, err := bech32.Decode(test.out)
		if err != nil {
			t.Errorf("Decode test #%d failed with err: %v", x, err)
		} else if prefix != test.prefix {
			t.Errorf("Decode test #%d failed: got prefix: %s want: %s", x, prefix, test.prefix)
		} else if version != test.version {
			t.Errorf("Decode test #%d failed: got version: %d want: %d", x, version, test.version)
		} else if string(decoded) != test.in {
			t.Errorf("Decode test #%d failed: got: %s want: %s", x, decoded, test.in)
		}
	}
}

func TestDecodeError(t *testing.T) {
	_, _, _, err := bech32.Decode("™")
	if err == nil {
		t.Errorf("decode unexpectedly succeeded")
	}
}
