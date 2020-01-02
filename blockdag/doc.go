/*
Package blockdag implements kaspa block handling and DAG selection rules.

The kaspa block handling and DAG selection rules are an integral, and quite
likely the most important, part of kaspa. At its core, kaspa is a distributed
consensus of which blocks are valid and which ones will comprise the DAG
(public ledger) that ultimately determines accepted transactions, so it is
extremely important that fully validating nodes agree on all rules.

At a high level, this package provides support for inserting new blocks into
the block DAG according to the aforementioned rules. It includes functionality
such as rejecting duplicate blocks, ensuring blocks and transactions follow all
rules, orphan handling, and DAG order along with reorganization.

Since this package does not deal with other kaspa specifics such as network
communication, it provides a notification system which gives the caller a high
level of flexibility in how they want to react to certain events such as orphan
blocks which need their parents requested and newly connected DAG blocks.

Kaspa DAG Processing Overview

Before a block is allowed into the block DAG, it must go through an intensive
series of validation rules. The following list serves as a general outline of
those rules to provide some intuition into what is going on under the hood, but
is by no means exhaustive:

 - Reject duplicate blocks
 - Perform a series of sanity checks on the block and its transactions such as
   verifying proof of work, timestamps, number and character of transactions,
   transaction amounts, script complexity, and merkle root calculations
 - Save the most recent orphan blocks for a limited time in case their parent
   blocks become available
 - Stop processing if the block is an orphan as the rest of the processing
   depends on the block's position within the block DAG
 - Perform a series of more thorough checks that depend on the block's position
   within the block DAG such as verifying block difficulties adhere to
   difficulty retarget rules, timestamps are after the median of the last
   several blocks, all transactions are finalized, and
   block versions are in line with the previous blocks
 - When a block is being connected to the DAG, perform further checks on the
   block's transactions such as verifying transaction duplicates, script
   complexity for the combination of connected scripts, coinbase maturity,
   double spends, and connected transaction values
 - Run the transaction scripts to verify the spender is allowed to spend the
   coins
 - Insert the block into the block database

Errors

Errors returned by this package are either the raw errors provided by underlying
calls or of type blockdag.RuleError. This allows the caller to differentiate
between unexpected errors, such as database errors, versus errors due to rule
violations through type assertions. In addition, callers can programmatically
determine the specific rule violation by examining the ErrorCode field of the
type asserted blockdag.RuleError.
*/
package blockdag
