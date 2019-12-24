rpctest
=======

[![ISC License](http://img.shields.io/badge/license-ISC-blue.svg)](http://copyfree.org)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/kaspanet/kaspad/integration/rpctest)

Package rpctest provides a kaspad-specific RPC testing harness crafting and
executing integration tests by driving a `kaspad` instance via the `RPC`
interface. Each instance of an active harness comes equipped with a simple
in-memory HD wallet capable of properly syncing to the generated DAG,
creating new addresses, and crafting fully signed transactions paying to an
arbitrary set of outputs.

