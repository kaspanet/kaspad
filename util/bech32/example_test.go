// Copyright (c) 2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package bech32_test

import (
	"encoding/hex"
	"fmt"
	"github.com/daglabs/btcd/util/bech32"
)

// This example demonstrates how to decode a bech32 encoded string.
func ExampleDecode() {
	encoded := "customprefix!:::::q:ppzxzarpyp6x7grzv5sx2mnrdajx2epqd9h8gmeqgfjkx6pnxgc3swlew4"
	prefix, decoded, version, err := bech32.Decode(encoded)
	if err != nil {
		fmt.Println("Error:", err)
	}

	// Show the decoded data.
	fmt.Println("Decoded prefix:", prefix)
	fmt.Println("Decoded version:", version)
	fmt.Println("Decoded Data:", hex.EncodeToString(decoded))

	// Output:
	// Decoded prefix: customprefix!:::::q
	// Decoded version: 8
	// Decoded Data: 4461746120746f20626520656e636f64656420696e746f20426563683332
}

// This example demonstrates how to encode data into a bech32 string.
func ExampleEncode() {
	data := []byte("Data to be encoded into Bech32")
	encoded := bech32.Encode("customprefix!:::::q", data, 8)

	// Show the encoded data.
	fmt.Println("Encoded Data:", encoded)

	// Output:
	// Encoded Data: customprefix!:::::q:ppzxzarpyp6x7grzv5sx2mnrdajx2epqd9h8gmeqgfjkx6pnxgc3swlew4
}
