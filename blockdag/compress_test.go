// Copyright (c) 2015-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"bytes"
	"encoding/hex"
	"testing"
)

// hexToBytes converts the passed hex string into bytes and will panic if there
// is an error. This is only provided for the hard-coded constants so errors in
// the source code can be detected. It will only (and must only) be called with
// hard-coded values.
func hexToBytes(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic("invalid hex in source file: " + s)
	}
	return b
}

// TestScriptCompression ensures the domain-specific script compression and
// decompression works as expected.
func TestScriptCompression(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		uncompressed []byte
		compressed   []byte
	}{
		{
			name:         "nil",
			uncompressed: nil,
			compressed:   hexToBytes("06"),
		},
		{
			name:         "pay-to-pubkey-hash 1",
			uncompressed: hexToBytes("76a9141018853670f9f3b0582c5b9ee8ce93764ac32b9388ac"),
			compressed:   hexToBytes("001018853670f9f3b0582c5b9ee8ce93764ac32b93"),
		},
		{
			name:         "pay-to-pubkey-hash 2",
			uncompressed: hexToBytes("76a914e34cce70c86373273efcc54ce7d2a491bb4a0e8488ac"),
			compressed:   hexToBytes("00e34cce70c86373273efcc54ce7d2a491bb4a0e84"),
		},
		{
			name:         "pay-to-script-hash 1",
			uncompressed: hexToBytes("a914da1745e9b549bd0bfa1a569971c77eba30cd5a4b87"),
			compressed:   hexToBytes("01da1745e9b549bd0bfa1a569971c77eba30cd5a4b"),
		},
		{
			name:         "pay-to-script-hash 2",
			uncompressed: hexToBytes("a914f815b036d9bbbce5e9f2a00abd1bf3dc91e9551087"),
			compressed:   hexToBytes("01f815b036d9bbbce5e9f2a00abd1bf3dc91e95510"),
		},
		{
			name:         "pay-to-pubkey compressed 0x02",
			uncompressed: hexToBytes("2102192d74d0cb94344c9569c2e77901573d8d7903c3ebec3a957724895dca52c6b4ac"),
			compressed:   hexToBytes("02192d74d0cb94344c9569c2e77901573d8d7903c3ebec3a957724895dca52c6b4"),
		},
		{
			name:         "pay-to-pubkey compressed 0x03",
			uncompressed: hexToBytes("2103b0bd634234abbb1ba1e986e884185c61cf43e001f9137f23c2c409273eb16e65ac"),
			compressed:   hexToBytes("03b0bd634234abbb1ba1e986e884185c61cf43e001f9137f23c2c409273eb16e65"),
		},
		{
			name:         "pay-to-pubkey uncompressed 0x04 even",
			uncompressed: hexToBytes("4104192d74d0cb94344c9569c2e77901573d8d7903c3ebec3a957724895dca52c6b40d45264838c0bd96852662ce6a847b197376830160c6d2eb5e6a4c44d33f453eac"),
			compressed:   hexToBytes("04192d74d0cb94344c9569c2e77901573d8d7903c3ebec3a957724895dca52c6b4"),
		},
		{
			name:         "pay-to-pubkey uncompressed 0x04 odd",
			uncompressed: hexToBytes("410411db93e1dcdb8a016b49840f8c53bc1eb68a382e97b1482ecad7b148a6909a5cb2e0eaddfb84ccf9744464f82e160bfa9b8b64f9d4c03f999b8643f656b412a3ac"),
			compressed:   hexToBytes("0511db93e1dcdb8a016b49840f8c53bc1eb68a382e97b1482ecad7b148a6909a5c"),
		},
		{
			name:         "pay-to-pubkey invalid pubkey",
			uncompressed: hexToBytes("3302aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaac"),
			compressed:   hexToBytes("293302aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaac"),
		},
		{
			name:         "requires 2 size bytes - data push 200 bytes",
			uncompressed: append(hexToBytes("4cc8"), bytes.Repeat([]byte{0x00}, 200)...),
			// [0x80, 0x50] = 208 as a variable length quantity
			// [0x4c, 0xc8] = OP_PUSHDATA1 200
			compressed: append(hexToBytes("d04cc8"), bytes.Repeat([]byte{0x00}, 200)...),
		},
	}

	for _, test := range tests {
		// Ensure the script compresses to the expected bytes.
		w := &bytes.Buffer{}
		err := putCompressedScript(w,
			test.uncompressed)
		if err != nil {
			t.Fatalf("putCompressedScript: %s", err)
		}

		gotCompressed := w.Bytes()
		if !bytes.Equal(gotCompressed, test.compressed) {
			t.Errorf("putCompressedScript (%s): did not get "+
				"expected bytes - got %x, want %x", test.name,
				gotCompressed, test.compressed)
			continue
		}

		// Ensure the script decompresses to the expected bytes.
		gotDecompressed, err := decompressScript(bytes.NewReader(test.compressed))
		if err != nil {
			t.Errorf("decompressScript (%s) "+
				"unexpected error: %s", test.name, err)
			continue
		}

		if !bytes.Equal(gotDecompressed, test.uncompressed) {
			t.Errorf("decompressScript (%s): did not get expected "+
				"bytes - got %x, want %x", test.name,
				gotDecompressed, test.uncompressed)
			continue
		}
	}
}

// TestScriptCompressionErrors ensures calling various functions related to
// script compression with incorrect data returns the expected results.
func TestScriptCompressionErrors(t *testing.T) {
	t.Parallel()

	// A nil script must result in a nil decompressed script.
	if _, err := decompressScript(bytes.NewReader(nil)); err == nil {
		t.Fatalf("decompressScript expects an error for an empty reader")
	}

	// A compressed script for a pay-to-pubkey (uncompressed) that results
	// in an invalid pubkey must result in a nil decompressed script.
	compressedScript := hexToBytes("04012d74d0cb94344c9569c2e77901573d8d" +
		"7903c3ebec3a957724895dca52c6b4")
	if _, err := decompressScript(bytes.NewReader(compressedScript)); err == nil {
		t.Fatalf("decompressScript with compressed pay-to-" +
			"uncompressed-pubkey that is invalid did not return " +
			"an error")
	}
}

// TestAmountCompression ensures the domain-specific transaction output amount
// compression and decompression works as expected.
func TestAmountCompression(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		uncompressed uint64
		compressed   uint64
	}{
		{
			name:         "0 KAS",
			uncompressed: 0,
			compressed:   0,
		},
		{
			name:         "546 Sompi (current network dust value)",
			uncompressed: 546,
			compressed:   4911,
		},
		{
			name:         "0.00001 KAS (typical transaction fee)",
			uncompressed: 1000,
			compressed:   4,
		},
		{
			name:         "0.0001 KAS (typical transaction fee)",
			uncompressed: 10000,
			compressed:   5,
		},
		{
			name:         "0.12345678 KAS",
			uncompressed: 12345678,
			compressed:   111111101,
		},
		{
			name:         "0.5 KAS",
			uncompressed: 50000000,
			compressed:   48,
		},
		{
			name:         "1 KAS",
			uncompressed: 100000000,
			compressed:   9,
		},
		{
			name:         "5 KAS",
			uncompressed: 500000000,
			compressed:   49,
		},
		{
			name:         "21000000 KAS (max minted coins)",
			uncompressed: 2100000000000000,
			compressed:   21000000,
		},
	}

	for _, test := range tests {
		// Ensure the amount compresses to the expected value.
		gotCompressed := compressTxOutAmount(test.uncompressed)
		if gotCompressed != test.compressed {
			t.Errorf("compressTxOutAmount (%s): did not get "+
				"expected value - got %d, want %d", test.name,
				gotCompressed, test.compressed)
			continue
		}

		// Ensure the value decompresses to the expected value.
		gotDecompressed := decompressTxOutAmount(test.compressed)
		if gotDecompressed != test.uncompressed {
			t.Errorf("decompressTxOutAmount (%s): did not get "+
				"expected value - got %d, want %d", test.name,
				gotDecompressed, test.uncompressed)
			continue
		}
	}
}

// TestCompressedTxOut ensures the transaction output serialization and
// deserialization works as expected.
func TestCompressedTxOut(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		amount       uint64
		scriptPubKey []byte
		compressed   []byte
	}{
		{
			name:         "pay-to-pubkey-hash dust",
			amount:       546,
			scriptPubKey: hexToBytes("76a9141018853670f9f3b0582c5b9ee8ce93764ac32b9388ac"),
			compressed:   hexToBytes("fd2f13001018853670f9f3b0582c5b9ee8ce93764ac32b93"),
		},
		{
			name:         "pay-to-pubkey uncompressed 1 KAS",
			amount:       100000000,
			scriptPubKey: hexToBytes("4104192d74d0cb94344c9569c2e77901573d8d7903c3ebec3a957724895dca52c6b40d45264838c0bd96852662ce6a847b197376830160c6d2eb5e6a4c44d33f453eac"),
			compressed:   hexToBytes("0904192d74d0cb94344c9569c2e77901573d8d7903c3ebec3a957724895dca52c6b4"),
		},
	}

	for _, test := range tests {
		// Ensure the txout compresses to the expected value.
		w := &bytes.Buffer{}
		err := putCompressedTxOut(w, test.amount, test.scriptPubKey)
		if err != nil {
			t.Fatalf("putCompressedTxOut: %s", err)
		}

		gotCompressed := w.Bytes()
		if !bytes.Equal(gotCompressed, test.compressed) {
			t.Errorf("compressTxOut (%s): did not get expected "+
				"bytes - got %x, want %x", test.name,
				gotCompressed, test.compressed)
			continue
		}

		// Ensure the serialized bytes are decoded back to the expected
		// uncompressed values.
		gotAmount, gotScript, err := decodeCompressedTxOut(
			bytes.NewReader(test.compressed))
		if err != nil {
			t.Errorf("decodeCompressedTxOut (%s): unexpected "+
				"error: %v", test.name, err)
			continue
		}
		if gotAmount != test.amount {
			t.Errorf("decodeCompressedTxOut (%s): did not get "+
				"expected amount - got %d, want %d",
				test.name, gotAmount, test.amount)
			continue
		}
		if !bytes.Equal(gotScript, test.scriptPubKey) {
			t.Errorf("decodeCompressedTxOut (%s): did not get "+
				"expected script - got %x, want %x",
				test.name, gotScript, test.scriptPubKey)
			continue
		}
	}
}

// TestTxOutCompressionErrors ensures calling various functions related to
// txout compression with incorrect data returns the expected results.
func TestTxOutCompressionErrors(t *testing.T) {
	t.Parallel()

	// A compressed txout with missing compressed script must error.
	compressedTxOut := hexToBytes("00")
	_, _, err := decodeCompressedTxOut(bytes.NewReader(compressedTxOut))
	if err == nil {
		t.Fatalf("decodeCompressedTxOut with missing compressed script " +
			"did not return an error")
	}

	// A compressed txout with short compressed script must error.
	compressedTxOut = hexToBytes("0010")
	_, _, err = decodeCompressedTxOut(bytes.NewReader(compressedTxOut))
	if err == nil {
		t.Fatalf("decodeCompressedTxOut with short compressed script " +
			"did not return an error")
	}
}
