// Copyright (c) 2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

/*
Package bech32 provides a Go implementation of the bech32 format.

Bech32 strings consist of a prefix, followed by the
separator :, then a checksummed data part encoded using the 32 characters
"qpzry9x8gf2tvdw0s3jn54khce6mua7l".

More info: https://github.com/daglabs/spec/blob/master/dagcoin.pdf
*/
package bech32