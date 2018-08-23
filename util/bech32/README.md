bech32dag
==========

[![Build Status](http://img.shields.io/travis/btcsuite/btcutil.svg)](https://travis-ci.org/btcsuite/btcutil)
[![ISC License](http://img.shields.io/badge/license-ISC-blue.svg)](http://copyfree.org)
[![GoDoc](https://godoc.org/github.com/daglabs/btcutil/bech32dag?status.png)](http://godoc.org/github.com/daglabs/btcutil/bech32)

Package bech32 provides a Go implementation of the bech32 format specified in
[the spec](https://github.com/daglabs/spec/blob/master/dagcoin.pdf).

Test vectors from the spec are added to ensure compatibility with the BIP.

## Installation and Updating

```bash
$ go get -u github.com/daglabs/btcutil/bech32dag
```

## Examples

* [Bech32 decode Example](http://godoc.org/github.com/daglabs/btcutil/bech32dag#example-Bech32Decode)
  Demonstrates how to decode a bech32 encoded string.
* [Bech32 encode Example](http://godoc.org/github.com/daglabs/btcutil/bech32dag#example-BechEncode)
  Demonstrates how to encode data into a bech32 string.

## License

Package bech32 is licensed under the [copyfree](http://copyfree.org) ISC
License.
