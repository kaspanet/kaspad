blockchain
==========

[![ISC License](http://img.shields.io/badge/license-ISC-blue.svg)](https://choosealicense.com/licenses/isc/)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/kaspanet/kaspad/blockchain)

Package blockdag implements Kaspa block handling, organization of the blockDAG, 
block sorting and UTXO-set maintenance.
The test coverage is currently only around 75%, but will be increasing over
time.

## Kaspad BlockDAG Processing Overview

Before a block is allowed into the block DAG, it must go through an intensive
series of validation rules. The following list serves as a general outline of
those rules to provide some intuition into what is going on under the hood, but
is by no means exhaustive:

 - Reject duplicate blocks
 - Perform a series of sanity checks on the block and its transactions such as
   verifying proof of work, timestamps, number and character of transactions,
   transaction amounts, script complexity, and merkle root calculations
 - Save the most recent orphan blocks for a limited time in case their parent
   blocks become available.
 - Save blocks from the future for delayed processing
 - Stop processing if the block is an orphan or delayed as the rest of the 
   processing depends on the block's position within the block chain
 - Make sure the block does not violate finality rules
 - Perform a series of more thorough checks that depend on the block's position
   within the blockDAG such as verifying block difficulties adhere to
   difficulty retarget rules, timestamps are after the median of the last
   several blocks, all transactions are finalized, checkpoint blocks match, and
   block versions are in line with the previous blocks
 - Determine how the block fits into the DAG and perform different actions
   accordingly 
 - Run the transaction scripts to verify the spender is allowed to spend the
   coins
 - Run GhostDAG to fit the block in a canonical sorting
 - Build the block's UTXO Set, as well as update the global UTXO Set accordingly
 - Insert the block into the block database

