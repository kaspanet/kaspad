// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

/*
This test file is part of the util package rather than than the
btcutil_test package so it can bridge access to the internals to properly test
cases which are either not possible or can't reliably be tested via the public
interface. The functions are only exported while the tests are being run.
*/

package util

import (
	"github.com/daglabs/btcd/btcec"
	"github.com/daglabs/btcd/util/bech32"
	"golang.org/x/crypto/ripemd160"
)

// SetBlockBytes sets the internal serialized block byte buffer to the passed
// buffer.  It is used to inject errors and is only available to the test
// package.
func (b *Block) SetBlockBytes(buf []byte) {
	b.serializedBlock = buf
}

// TstAppDataDir makes the internal appDataDir function available to the test
// package.
func TstAppDataDir(goos, appName string, roaming bool) string {
	return appDataDir(goos, appName, roaming)
}

func TstAddressPubKeyHash(prefix Bech32Prefix, hash [ripemd160.Size]byte) *AddressPubKeyHash {
	return &AddressPubKeyHash{
		prefix: prefix,
		hash:   hash,
	}
}

// TstAddressScriptHash makes an AddressScriptHash, setting the
// unexported fields with the parameters hash and netID.
func TstAddressScriptHash(prefix Bech32Prefix, hash [ripemd160.Size]byte) *AddressScriptHash {

	return &AddressScriptHash{
		prefix: prefix,
		hash:   hash,
	}
}

// TstAddressPubKey makes an AddressPubKey, setting the unexported fields with
// the parameters.
func TstAddressPubKey(serializedPubKey []byte, pubKeyFormat PubKeyFormat,
	netID byte) *AddressPubKey {

	pubKey, _ := btcec.ParsePubKey(serializedPubKey, btcec.S256())
	return &AddressPubKey{
		pubKeyFormat: pubKeyFormat,
		pubKey:       (*btcec.PublicKey)(pubKey),
		pubKeyHashID: netID,
	}
}

// TstAddressSAddr returns the expected script address bytes for
// P2PKH and P2SH bitcoin addresses.
func TstAddressSAddr(addr string) []byte {
	_, decoded, _, _ := bech32.Decode(addr)
	return decoded[:ripemd160.Size]
}
