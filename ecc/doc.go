/*
Package ecc implements support for the elliptic curves needed for kaspa.

Kaspa uses elliptic curve cryptography using koblitz curves
(specifically secp256k1) for cryptographic functions. See
http://www.secg.org/collateral/sec2_final.pdf for details on the
standard.

This package provides the data structures and functions implementing the
crypto/elliptic Curve interface in order to permit using these curves
with the standard crypto/ecdsa package provided with go. Helper
functionality is provided to parse signatures and public keys from
standard formats. It was originally based on some initial work by
ThePiachu, but has significantly diverged since then.
*/
package ecc
