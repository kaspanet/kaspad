// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package util_test

import (
	"bytes"
	"fmt"
	"golang.org/x/crypto/blake2b"
	"reflect"
	"strings"
	"testing"

	"github.com/kaspanet/kaspad/util"
)

func TestAddresses(t *testing.T) {
	tests := []struct {
		name           string
		addr           string
		encoded        string
		valid          bool
		result         util.Address
		f              func() (util.Address, error)
		passedPrefix   util.Bech32Prefix
		expectedPrefix util.Bech32Prefix
	}{
		// Positive P2PKH tests.
		{
			name:    "mainnet p2pkh",
			addr:    "kaspa:qr35ennsep3hxfe7lnz5ee7j5jgmkjswsn35ennsep3hxfe7ln35cdv0dy335",
			encoded: "kaspa:qr35ennsep3hxfe7lnz5ee7j5jgmkjswsn35ennsep3hxfe7ln35cdv0dy335",
			valid:   true,
			result: util.TstAddressPubKeyHash(
				util.Bech32PrefixKaspa,
				[blake2b.Size256]byte{
					0xe3, 0x4c, 0xce, 0x70, 0xc8, 0x63, 0x73, 0x27, 0x3e, 0xfc,
					0xc5, 0x4c, 0xe7, 0xd2, 0xa4, 0x91, 0xbb, 0x4a, 0x0e, 0x84,
					0xe3, 0x4c, 0xce, 0x70, 0xc8, 0x63, 0x73, 0x27, 0x3e, 0xfc,
					0xe3, 0x4c,
				}),
			f: func() (util.Address, error) {
				pkHash := []byte{
					0xe3, 0x4c, 0xce, 0x70, 0xc8, 0x63, 0x73, 0x27, 0x3e, 0xfc,
					0xc5, 0x4c, 0xe7, 0xd2, 0xa4, 0x91, 0xbb, 0x4a, 0x0e, 0x84,
					0xe3, 0x4c, 0xce, 0x70, 0xc8, 0x63, 0x73, 0x27, 0x3e, 0xfc,
					0xe3, 0x4c}
				return util.NewAddressPubKeyHash(pkHash, util.Bech32PrefixKaspa)
			},
			passedPrefix:   util.Bech32PrefixUnknown,
			expectedPrefix: util.Bech32PrefixKaspa,
		},
		{
			name:    "mainnet p2pkh 2",
			addr:    "kaspa:qq80qvqs0lfxuzmt7sz3909ze6camq9d4t35ennsep3hxfe7ln35cvfqgz3z8",
			encoded: "kaspa:qq80qvqs0lfxuzmt7sz3909ze6camq9d4t35ennsep3hxfe7ln35cvfqgz3z8",
			valid:   true,
			result: util.TstAddressPubKeyHash(
				util.Bech32PrefixKaspa,
				[blake2b.Size256]byte{
					0x0e, 0xf0, 0x30, 0x10, 0x7f, 0xd2, 0x6e, 0x0b, 0x6b, 0xf4,
					0x05, 0x12, 0xbc, 0xa2, 0xce, 0xb1, 0xdd, 0x80, 0xad, 0xaa,
					0xe3, 0x4c, 0xce, 0x70, 0xc8, 0x63, 0x73, 0x27, 0x3e, 0xfc,
					0xe3, 0x4c,
				}),
			f: func() (util.Address, error) {
				pkHash := []byte{
					0x0e, 0xf0, 0x30, 0x10, 0x7f, 0xd2, 0x6e, 0x0b, 0x6b, 0xf4,
					0x05, 0x12, 0xbc, 0xa2, 0xce, 0xb1, 0xdd, 0x80, 0xad, 0xaa,
					0xe3, 0x4c, 0xce, 0x70, 0xc8, 0x63, 0x73, 0x27, 0x3e, 0xfc,
					0xe3, 0x4c,
				}
				return util.NewAddressPubKeyHash(pkHash, util.Bech32PrefixKaspa)
			},
			passedPrefix:   util.Bech32PrefixKaspa,
			expectedPrefix: util.Bech32PrefixKaspa,
		},
		{
			name:    "testnet p2pkh",
			addr:    "kaspatest:qputx94qseratdmjs0j395mq8u03er0x3l35ennsep3hxfe7ln35ckquw528z",
			encoded: "kaspatest:qputx94qseratdmjs0j395mq8u03er0x3l35ennsep3hxfe7ln35ckquw528z",
			valid:   true,
			result: util.TstAddressPubKeyHash(
				util.Bech32PrefixKaspaTest,
				[blake2b.Size256]byte{
					0x78, 0xb3, 0x16, 0xa0, 0x86, 0x47, 0xd5, 0xb7, 0x72, 0x83,
					0xe5, 0x12, 0xd3, 0x60, 0x3f, 0x1f, 0x1c, 0x8d, 0xe6, 0x8f,
					0xe3, 0x4c, 0xce, 0x70, 0xc8, 0x63, 0x73, 0x27, 0x3e, 0xfc,
					0xe3, 0x4c,
				}),
			f: func() (util.Address, error) {
				pkHash := []byte{
					0x78, 0xb3, 0x16, 0xa0, 0x86, 0x47, 0xd5, 0xb7, 0x72, 0x83,
					0xe5, 0x12, 0xd3, 0x60, 0x3f, 0x1f, 0x1c, 0x8d, 0xe6, 0x8f,
					0xe3, 0x4c, 0xce, 0x70, 0xc8, 0x63, 0x73, 0x27, 0x3e, 0xfc,
					0xe3, 0x4c,
				}
				return util.NewAddressPubKeyHash(pkHash, util.Bech32PrefixKaspaTest)
			},
			passedPrefix:   util.Bech32PrefixKaspaTest,
			expectedPrefix: util.Bech32PrefixKaspaTest,
		},

		// Negative P2PKH tests.
		{
			name:  "p2pkh wrong hash length",
			addr:  "",
			valid: false,
			f: func() (util.Address, error) {
				pkHash := []byte{
					0x00, 0x0e, 0xf0, 0x30, 0x10, 0x7f, 0xd2, 0x6e, 0x0b, 0x6b,
					0xf4, 0x05, 0x12, 0xbc, 0xa2, 0xce, 0xb1, 0xdd, 0x80, 0xad,
					0xaa}
				return util.NewAddressPubKeyHash(pkHash, util.Bech32PrefixKaspa)
			},
			passedPrefix:   util.Bech32PrefixKaspa,
			expectedPrefix: util.Bech32PrefixKaspa,
		},
		{
			name:           "p2pkh bad checksum",
			addr:           "kaspa:qr35ennsep3hxfe7lnz5ee7j5jgmkjswss74as46gx",
			valid:          false,
			passedPrefix:   util.Bech32PrefixKaspa,
			expectedPrefix: util.Bech32PrefixKaspa,
		},

		// Positive P2SH tests.
		{
			name:    "mainnet p2sh",
			addr:    "kaspa:prq20q4qd9ulr044cauyy9wtpeupqpjv67pn2vyc6acly7xqkrjdzmh8rj9f4",
			encoded: "kaspa:prq20q4qd9ulr044cauyy9wtpeupqpjv67pn2vyc6acly7xqkrjdzmh8rj9f4",
			valid:   true,
			result: util.TstAddressScriptHash(
				util.Bech32PrefixKaspa,
				[blake2b.Size256]byte{
					0xc0, 0xa7, 0x82, 0xa0, 0x69, 0x79, 0xf1, 0xbe,
					0xb5, 0xc7, 0x78, 0x42, 0x15, 0xcb, 0x0e, 0x78,
					0x10, 0x06, 0x4c, 0xd7, 0x83, 0x35, 0x30, 0x98,
					0xd7, 0x71, 0xf2, 0x78, 0xc0, 0xb0, 0xe4, 0xd1,
				}),
			f: func() (util.Address, error) {
				script := []byte{
					0x52, 0x41, 0x04, 0x91, 0xbb, 0xa2, 0x51, 0x09, 0x12, 0xa5,
					0xbd, 0x37, 0xda, 0x1f, 0xb5, 0xb1, 0x67, 0x30, 0x10, 0xe4,
					0x3d, 0x2c, 0x6d, 0x81, 0x2c, 0x51, 0x4e, 0x91, 0xbf, 0xa9,
					0xf2, 0xeb, 0x12, 0x9e, 0x1c, 0x18, 0x33, 0x29, 0xdb, 0x55,
					0xbd, 0x86, 0x8e, 0x20, 0x9a, 0xac, 0x2f, 0xbc, 0x02, 0xcb,
					0x33, 0xd9, 0x8f, 0xe7, 0x4b, 0xf2, 0x3f, 0x0c, 0x23, 0x5d,
					0x61, 0x26, 0xb1, 0xd8, 0x33, 0x4f, 0x86, 0x41, 0x04, 0x86,
					0x5c, 0x40, 0x29, 0x3a, 0x68, 0x0c, 0xb9, 0xc0, 0x20, 0xe7,
					0xb1, 0xe1, 0x06, 0xd8, 0xc1, 0x91, 0x6d, 0x3c, 0xef, 0x99,
					0xaa, 0x43, 0x1a, 0x56, 0xd2, 0x53, 0xe6, 0x92, 0x56, 0xda,
					0xc0, 0x9e, 0xf1, 0x22, 0xb1, 0xa9, 0x86, 0x81, 0x8a, 0x7c,
					0xb6, 0x24, 0x53, 0x2f, 0x06, 0x2c, 0x1d, 0x1f, 0x87, 0x22,
					0x08, 0x48, 0x61, 0xc5, 0xc3, 0x29, 0x1c, 0xcf, 0xfe, 0xf4,
					0xec, 0x68, 0x74, 0x41, 0x04, 0x8d, 0x24, 0x55, 0xd2, 0x40,
					0x3e, 0x08, 0x70, 0x8f, 0xc1, 0xf5, 0x56, 0x00, 0x2f, 0x1b,
					0x6c, 0xd8, 0x3f, 0x99, 0x2d, 0x08, 0x50, 0x97, 0xf9, 0x97,
					0x4a, 0xb0, 0x8a, 0x28, 0x83, 0x8f, 0x07, 0x89, 0x6f, 0xba,
					0xb0, 0x8f, 0x39, 0x49, 0x5e, 0x15, 0xfa, 0x6f, 0xad, 0x6e,
					0xdb, 0xfb, 0x1e, 0x75, 0x4e, 0x35, 0xfa, 0x1c, 0x78, 0x44,
					0xc4, 0x1f, 0x32, 0x2a, 0x18, 0x63, 0xd4, 0x62, 0x13, 0x53,
					0xae}
				return util.NewAddressScriptHash(script, util.Bech32PrefixKaspa)
			},
			passedPrefix:   util.Bech32PrefixKaspa,
			expectedPrefix: util.Bech32PrefixKaspa,
		},
		{
			// Taken from transactions:
			// output: b0539a45de13b3e0403909b8bd1a555b8cbe45fd4e3f3fda76f3a5f52835c29d
			// input: (not yet redeemed at time test was written)
			name:    "mainnet p2sh 2",
			addr:    "kaspa:pr5vxqxg0xrwl2zvxlq9rxffqx00sm44ksqqqqqqqqqqqqqqqqqqq33flv3je",
			encoded: "kaspa:pr5vxqxg0xrwl2zvxlq9rxffqx00sm44ksqqqqqqqqqqqqqqqqqqq33flv3je",
			valid:   true,
			result: util.TstAddressScriptHash(
				util.Bech32PrefixKaspa,
				[blake2b.Size256]byte{
					0xe8, 0xc3, 0x00, 0xc8, 0x79, 0x86, 0xef, 0xa8, 0x4c, 0x37,
					0xc0, 0x51, 0x99, 0x29, 0x01, 0x9e, 0xf8, 0x6e, 0xb5, 0xb4,
					0xe8, 0xc3, 0x00, 0xc8, 0x79, 0x86, 0xef, 0xa8, 0x4c, 0x37,
					0xe8, 0xc3,
				}),
			f: func() (util.Address, error) {
				hash := []byte{
					0xe8, 0xc3, 0x00, 0xc8, 0x79, 0x86, 0xef, 0xa8, 0x4c, 0x37,
					0xc0, 0x51, 0x99, 0x29, 0x01, 0x9e, 0xf8, 0x6e, 0xb5, 0xb4,
					0xe8, 0xc3, 0x00, 0xc8, 0x79, 0x86, 0xef, 0xa8, 0x4c, 0x37,
					0xe8, 0xc3,
				}
				return util.NewAddressScriptHashFromHash(hash, util.Bech32PrefixKaspa)
			},
			passedPrefix:   util.Bech32PrefixKaspa,
			expectedPrefix: util.Bech32PrefixKaspa,
		},
		{
			name:    "testnet p2sh",
			addr:    "kaspatest:przhjdpv93xfygpqtckdc2zkzuzqeyj2pgqqqqqqqqqqqqqqqqqqqyjpt4duk",
			encoded: "kaspatest:przhjdpv93xfygpqtckdc2zkzuzqeyj2pgqqqqqqqqqqqqqqqqqqqyjpt4duk",
			valid:   true,
			result: util.TstAddressScriptHash(
				util.Bech32PrefixKaspaTest,
				[blake2b.Size256]byte{
					0xc5, 0x79, 0x34, 0x2c, 0x2c, 0x4c, 0x92, 0x20, 0x20, 0x5e,
					0x2c, 0xdc, 0x28, 0x56, 0x17, 0x04, 0x0c, 0x92, 0x4a, 0x0a,
					0xe8, 0xc3, 0x00, 0xc8, 0x79, 0x86, 0xef, 0xa8, 0x4c, 0x37,
					0xe8, 0xc3,
				}),
			f: func() (util.Address, error) {
				hash := []byte{
					0xc5, 0x79, 0x34, 0x2c, 0x2c, 0x4c, 0x92, 0x20, 0x20, 0x5e,
					0x2c, 0xdc, 0x28, 0x56, 0x17, 0x04, 0x0c, 0x92, 0x4a, 0x0a,
					0xe8, 0xc3, 0x00, 0xc8, 0x79, 0x86, 0xef, 0xa8, 0x4c, 0x37,
					0xe8, 0xc3,
				}
				return util.NewAddressScriptHashFromHash(hash, util.Bech32PrefixKaspaTest)
			},
			passedPrefix:   util.Bech32PrefixKaspaTest,
			expectedPrefix: util.Bech32PrefixKaspaTest,
		},

		// Negative P2SH tests.
		{
			name:  "p2sh wrong hash length",
			addr:  "",
			valid: false,
			f: func() (util.Address, error) {
				hash := []byte{
					0x00, 0xf8, 0x15, 0xb0, 0x36, 0xd9, 0xbb, 0xbc, 0xe5, 0xe9,
					0xf2, 0xa0, 0x0a, 0xbd, 0x1b, 0xf3, 0xdc, 0x91, 0xe9, 0x55,
					0x10}
				return util.NewAddressScriptHashFromHash(hash, util.Bech32PrefixKaspa)
			},
			passedPrefix:   util.Bech32PrefixKaspa,
			expectedPrefix: util.Bech32PrefixKaspa,
		},
	}

	for _, test := range tests {
		// Decode addr and compare error against valid.
		decoded, err := util.DecodeAddress(test.addr, test.passedPrefix)
		if (err == nil) != test.valid {
			t.Errorf("%v: decoding test failed: %v", test.name, err)
			return
		}

		if err == nil {
			// Ensure the stringer returns the same address as the
			// original.
			if decodedStringer, ok := decoded.(fmt.Stringer); ok {
				addr := test.addr

				if addr != decodedStringer.String() {
					t.Errorf("%v: String on decoded value does not match expected value: %v != %v",
						test.name, test.addr, decodedStringer.String())
					return
				}
			}

			// Encode again and compare against the original.
			encoded := decoded.EncodeAddress()
			if test.encoded != encoded {
				t.Errorf("%v: decoding and encoding produced different addressess: %v != %v",
					test.name, test.encoded, encoded)
				return
			}

			// Perform type-specific calculations.
			var saddr []byte
			switch decoded.(type) {
			case *util.AddressPubKeyHash:
				saddr = util.TstAddressSAddr(encoded)

			case *util.AddressScriptHash:
				saddr = util.TstAddressSAddr(encoded)
			}

			// Check script address, as well as the HashBlake2b method for P2PKH and
			// P2SH addresses.
			if !bytes.Equal(saddr, decoded.ScriptAddress()) {
				t.Errorf("%v: script addresses do not match:\n%x != \n%x",
					test.name, saddr, decoded.ScriptAddress())
				return
			}
			switch a := decoded.(type) {
			case *util.AddressPubKeyHash:
				if h := a.HashBlake2b()[:]; !bytes.Equal(saddr, h) {
					t.Errorf("%v: hashes do not match:\n%x != \n%x",
						test.name, saddr, h)
					return
				}

			case *util.AddressScriptHash:
				if h := a.HashBlake2b()[:]; !bytes.Equal(saddr, h) {
					t.Errorf("%v: hashes do not match:\n%x != \n%x",
						test.name, saddr, h)
					return
				}
			}

			// Ensure the address is for the expected network.
			if !decoded.IsForPrefix(test.expectedPrefix) {
				t.Errorf("%v: calculated network does not match expected",
					test.name)
				return
			}
		}

		if !test.valid {
			// If address is invalid, but a creation function exists,
			// verify that it returns a nil addr and non-nil error.
			if test.f != nil {
				_, err := test.f()
				if err == nil {
					t.Errorf("%v: address is invalid but creating new address succeeded",
						test.name)
					return
				}
			}
			continue
		}

		// Valid test, compare address created with f against expected result.
		addr, err := test.f()
		if err != nil {
			t.Errorf("%v: address is valid but creating new address failed with error %v",
				test.name, err)
			return
		}

		if !reflect.DeepEqual(addr, test.result) {
			t.Errorf("%v: created address does not match expected result",
				test.name)
			return
		}
	}
}

func TestDecodeAddressErrorConditions(t *testing.T) {
	tests := []struct {
		address      string
		prefix       util.Bech32Prefix
		errorMessage string
	}{
		{
			"bitcoincash:qpzry9x8gf2tvdw0s3jn54khce6mua7lcw20ayyn",
			util.Bech32PrefixUnknown,
			"decoded address's prefix could not be parsed",
		},
		{
			"kaspasim:raskzctpv9skzctpv9skzctpv9skzctpvy37ct7zafpv9skzctpvymmnd3gh8",
			util.Bech32PrefixKaspaSim,
			"unknown address type",
		},
		{
			"kaspasim:raskzcg58mth0an",
			util.Bech32PrefixKaspaSim,
			"decoded address is of unknown size",
		},
		{
			"kaspatest:qqq65mvpxcmajeq44n2n8vfn6u9f8l4zsy0xez0tzw",
			util.Bech32PrefixKaspa,
			"decoded address is of wrong network",
		},
	}

	for _, test := range tests {
		_, err := util.DecodeAddress(test.address, test.prefix)
		if err == nil {
			t.Errorf("decodeAddress unexpectedly succeeded")
		} else if !strings.Contains(err.Error(), test.errorMessage) {
			t.Errorf("received mismatched error. Expected '%s' but got '%s'",
				test.errorMessage, err)
		}
	}
}

func TestParsePrefix(t *testing.T) {
	tests := []struct {
		prefixStr      string
		expectedPrefix util.Bech32Prefix
		expectedError  bool
	}{
		{"kaspa", util.Bech32PrefixKaspa, false},
		{"kaspatest", util.Bech32PrefixKaspaTest, false},
		{"kaspasim", util.Bech32PrefixKaspaSim, false},
		{"blabla", util.Bech32PrefixUnknown, true},
		{"unknown", util.Bech32PrefixUnknown, true},
		{"", util.Bech32PrefixUnknown, true},
	}

	for _, test := range tests {
		result, err := util.ParsePrefix(test.prefixStr)
		if (err != nil) != test.expectedError {
			t.Errorf("TestParsePrefix: %s: expected error status: %t, but got %t",
				test.prefixStr, test.expectedError, err != nil)
		}

		if result != test.expectedPrefix {
			t.Errorf("TestParsePrefix: %s: expected prefix: %d, but got %d",
				test.prefixStr, test.expectedPrefix, result)
		}
	}
}

func TestPrefixToString(t *testing.T) {
	tests := []struct {
		prefix            util.Bech32Prefix
		expectedPrefixStr string
	}{
		{util.Bech32PrefixKaspa, "kaspa"},
		{util.Bech32PrefixKaspaTest, "kaspatest"},
		{util.Bech32PrefixKaspaSim, "kaspasim"},
		{util.Bech32PrefixUnknown, ""},
	}

	for _, test := range tests {
		result := test.prefix.String()

		if result != test.expectedPrefixStr {
			t.Errorf("TestPrefixToString: %s: expected string: %s, but got %s",
				test.prefix, test.expectedPrefixStr, result)
		}
	}
}
