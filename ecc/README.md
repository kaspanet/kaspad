ecc
=====

[![ISC License](http://img.shields.io/badge/license-ISC-blue.svg)](https://choosealicense.com/licenses/isc/)
[![GoDoc](https://godoc.org/github.com/kaspanet/kaspad/ecc?status.png)](http://godoc.org/github.com/kaspanet/kaspad/ecc)

Package ecc implements elliptic curve cryptography needed for working with
Kaspa. It is designed so that it may be used with the standard crypto/ecdsa 
packages provided with go. A comprehensive suite of tests is provided to ensure 
proper functionality. Package ecc was originally based on work from ThePiachu 
which is licensed under the same terms as Go, but it has signficantly diverged 
since then. The kaspanet developers original is licensed under the liberal ISC 
license.

## Examples

* [Sign Message](http://godoc.org/github.com/kaspanet/kaspad/ecc#example-package--SignMessage) 
  Demonstrates signing a message with a secp256k1 private key that is first
  parsed form raw bytes and serializing the generated signature.

* [Verify Signature](http://godoc.org/github.com/kaspanet/kaspad/ecc#example-package--VerifySignature) 
  Demonstrates verifying a secp256k1 signature against a public key that is
  first parsed from raw bytes. The signature is also parsed from raw bytes.

