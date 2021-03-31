// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

/*
This test file is part of the util package rather than than the
util_test package so it can bridge access to the internals to properly test
cases which are either not possible or can't reliably be tested via the public
interface. The functions are only exported while the tests are being run.
*/

package util

import (
	"github.com/kaspanet/kaspad/util/bech32"
	"golang.org/x/crypto/blake2b"
)

// TstAppDataDir makes the internal appDir function available to the test
// package.
func TstAppDataDir(goos, appName string, roaming bool) string {
	return appDir(goos, appName, roaming)
}

func TstAddressPubKey(prefix Bech32Prefix, hash [PublicKeySize]byte) *AddressPubKey {
	return &AddressPubKey{
		prefix: prefix,
		pubKey: hash,
	}
}

// TstAddressScriptHash makes an AddressScriptHash, setting the
// unexported fields with the parameters hash and netID.
func TstAddressScriptHash(prefix Bech32Prefix, hash [blake2b.Size256]byte) *AddressScriptHash {

	return &AddressScriptHash{
		prefix: prefix,
		hash:   hash,
	}
}

// TstAddressSAddr returns the expected script address bytes for
// P2PKH kaspa addresses.
func TstAddressSAddrP2PK(addr string) []byte {
	_, decoded, _, _ := bech32.Decode(addr)
	return decoded[:PublicKeySize]
}

// TstAddressSAddrP2SH returns the expected script address bytes for
// P2SH kaspa addresses.
func TstAddressSAddrP2SH(addr string) []byte {
	_, decoded, _, _ := bech32.Decode(addr)
	return decoded[:blake2b.Size256]
}
