indexers
========

[![ISC License](http://img.shields.io/badge/license-ISC-blue.svg)](https://choosealicense.com/licenses/isc/)
[![GoDoc](https://godoc.org/github.com/kaspanet/kaspad/blockdag/indexers?status.png)](http://godoc.org/github.com/kaspanet/kaspad/blockdag/indexers)

Package indexers implements optional block chain indexes.

These indexes are typically used to enhance the amount of information available
via an RPC interface.

## Supported Indexers

- Transaction-by-hash (txindex) Index
  - Creates a mapping from the hash of each transaction to the block that
    contains it along with its offset and length within the serialized block
- Transaction-by-address (addrindex) Index
  - Creates a mapping from every address to all transactions which either credit
    or debit the address
  - Requires the transaction-by-hash index
- AcceptanceData-by-block Index
  - Creates a mapping from the hash of each block to the list of transaction this block
    accepts from it's .Blues

