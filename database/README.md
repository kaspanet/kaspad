database
========

[![ISC License](http://img.shields.io/badge/license-ISC-blue.svg)](http://copyfree.org)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/kaspanet/kaspad/database)

Package database provides a block and metadata storage database.

Please note that this package is intended to enable kaspad to support different
database backends and is not something that a client can directly access as only
one entity can have the database open at a time (for most database backends),
and that entity will be kaspad.

When a client wants programmatic access to the data provided by kaspad, they'll
likely want to use the [rpcclient](https://github.com/kaspanet/kaspad/tree/master/rpcclient)
package which makes use of the [JSON-RPC API](https://github.com/kaspanet/kaspad/tree/master/docs/json_rpc_api.md).

The default backend, ffldb, has a strong focus on speed, efficiency, and
robustness.  It makes use of leveldb for the metadata, flat files for block
storage, and strict checksums in key areas to ensure data integrity.

## Feature Overview

- Key/value metadata store
- Kaspa block storage
- Efficient retrieval of block headers and regions (transactions, scripts, etc)
- Read-only and read-write transactions with both manual and managed modes
- Nested buckets
- Iteration support including cursors with seek capability
- Supports registration of backend databases
- Comprehensive test coverage

