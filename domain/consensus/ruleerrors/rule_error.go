package ruleerrors

import (
	"fmt"
	"github.com/pkg/errors"
)

// ErrorCode identifies a kind of error.
type ErrorCode int

// These constants are used to identify a specific RuleError.
const (
	// ErrDuplicateBlock indicates a block with the same hash already
	// exists.
	ErrDuplicateBlock ErrorCode = iota

	// ErrBlockMassTooHigh indicates the mass of a block exceeds the maximum
	// allowed limits.
	ErrBlockMassTooHigh

	// ErrBlockVersionTooOld indicates the block version is too old and is
	// no longer accepted since the majority of the network has upgraded
	// to a newer version.
	ErrBlockVersionTooOld

	// ErrTimeTooOld indicates the time is either before the median time of
	// the last several blocks per the DAG consensus rules.
	ErrTimeTooOld

	// ErrTimeTooNew indicates the time is too far in the future as compared
	// the current time.
	ErrTimeTooNew

	// ErrNoParents indicates that the block is missing parents
	ErrNoParents

	// ErrWrongParentsOrder indicates that the block's parents are not ordered by hash, as expected
	ErrWrongParentsOrder

	// ErrDifficultyTooLow indicates the difficulty for the block is lower
	// than the difficulty required.
	ErrDifficultyTooLow

	// ErrUnexpectedDifficulty indicates specified bits do not align with
	// the expected value either because it doesn't match the calculated
	// valued based on difficulty regarted rules or it is out of the valid
	// range.
	ErrUnexpectedDifficulty

	// ErrHighHash indicates the block does not hash to a value which is
	// lower than the required target difficultly.
	ErrHighHash

	// ErrBadMerkleRoot indicates the calculated merkle root does not match
	// the expected value.
	ErrBadMerkleRoot

	// ErrBadUTXOCommitment indicates the calculated UTXO commitment does not match
	// the expected value.
	ErrBadUTXOCommitment

	// ErrInvalidSubnetwork indicates the subnetwork is now allowed.
	ErrInvalidSubnetwork

	// ErrFinalityPointTimeTooOld indicates a block has a timestamp before the
	// last finality point.
	ErrFinalityPointTimeTooOld

	// ErrNoTransactions indicates the block does not have a least one
	// transaction. A valid block must have at least the coinbase
	// transaction.
	ErrNoTransactions

	// ErrNoTxInputs indicates a transaction does not have any inputs. A
	// valid transaction must have at least one input.
	ErrNoTxInputs

	// ErrTxMassTooHigh indicates the mass of a transaction exceeds the maximum
	// allowed limits.
	ErrTxMassTooHigh

	// ErrBadTxOutValue indicates an output value for a transaction is
	// invalid in some way such as being out of range.
	ErrBadTxOutValue

	// ErrDuplicateTxInputs indicates a transaction references the same
	// input more than once.
	ErrDuplicateTxInputs

	// ErrBadTxInput indicates a transaction input is invalid in some way
	// such as referencing a previous transaction outpoint which is out of
	// range or not referencing one at all.
	ErrBadTxInput

	// ErrMissingTxOut indicates a transaction output referenced by an input
	// either does not exist or has already been spent.
	ErrMissingTxOut

	// ErrDoubleSpendInSameBlock indicates a transaction
	// that spends an output that was already spent by another
	// transaction in the same block.
	ErrDoubleSpendInSameBlock

	// ErrUnfinalizedTx indicates a transaction has not been finalized.
	// A valid block may only contain finalized transactions.
	ErrUnfinalizedTx

	// ErrDuplicateTx indicates a block contains an identical transaction
	// (or at least two transactions which hash to the same value). A
	// valid block may only contain unique transactions.
	ErrDuplicateTx

	// ErrOverwriteTx indicates a block contains a transaction that has
	// the same hash as a previous transaction which has not been fully
	// spent.
	ErrOverwriteTx

	// ErrImmatureSpend indicates a transaction is attempting to spend a
	// coinbase that has not yet reached the required maturity.
	ErrImmatureSpend

	// ErrSpendTooHigh indicates a transaction is attempting to spend more
	// value than the sum of all of its inputs.
	ErrSpendTooHigh

	// ErrBadFees indicates the total fees for a block are invalid due to
	// exceeding the maximum possible value.
	ErrBadFees

	// ErrTooManySigOps indicates the total number of signature operations
	// for a transaction or block exceed the maximum allowed limits.
	ErrTooManySigOps

	// ErrFirstTxNotCoinbase indicates the first transaction in a block
	// is not a coinbase transaction.
	ErrFirstTxNotCoinbase

	// ErrMultipleCoinbases indicates a block contains more than one
	// coinbase transaction.
	ErrMultipleCoinbases

	// ErrBadCoinbasePayloadLen indicates the length of the payload
	// for a coinbase transaction is too high.
	ErrBadCoinbasePayloadLen

	// ErrBadCoinbaseTransaction indicates that the block's coinbase transaction is not build as expected
	ErrBadCoinbaseTransaction

	// ErrScriptMalformed indicates a transaction script is malformed in
	// some way. For example, it might be longer than the maximum allowed
	// length or fail to parse.
	ErrScriptMalformed

	// ErrScriptValidation indicates the result of executing transaction
	// script failed. The error covers any failure when executing scripts
	// such signature verification failures and execution past the end of
	// the stack.
	ErrScriptValidation

	// ErrParentBlockUnknown indicates that the parent block is not known.
	ErrParentBlockUnknown

	// ErrInvalidAncestorBlock indicates that an ancestor of this block has
	// already failed validation.
	ErrInvalidAncestorBlock

	// ErrParentBlockNotCurrentTips indicates that the block's parents are not the
	// current tips. This is not a block validation rule, but is required
	// for block proposals submitted via getblocktemplate RPC.
	ErrParentBlockNotCurrentTips

	// ErrWithDiff indicates that there was an error with UTXOSet.WithDiff
	ErrWithDiff

	// ErrFinality indicates that a block doesn't adhere to the finality rules
	ErrFinality

	// ErrTransactionsNotSorted indicates that transactions in block are not
	// sorted by subnetwork
	ErrTransactionsNotSorted

	// ErrInvalidGas transaction wants to use more GAS than allowed
	// by subnetwork
	ErrInvalidGas

	// ErrInvalidPayload transaction includes a payload in a subnetwork that doesn't allow
	// a Payload
	ErrInvalidPayload

	// ErrInvalidPayloadHash invalid hash of transaction's payload
	ErrInvalidPayloadHash

	// ErrSubnetwork indicates that a block doesn't adhere to the subnetwork
	// registry rules
	ErrSubnetworkRegistry

	// ErrInvalidParentsRelation indicates that one of the parents of a block
	// is also an ancestor of another parent
	ErrInvalidParentsRelation

	// ErrTooManyParents indicates that a block points to more then `MaxNumParentBlocks` parents
	ErrTooManyParents

	// ErrDelayedBlockIsNotAllowed indicates that a block with a delayed timestamp was
	// submitted with BFDisallowDelay flag raised.
	ErrDelayedBlockIsNotAllowed

	// ErrOrphanBlockIsNotAllowed indicates that an orphan block was submitted with
	// BFDisallowOrphans flag raised.
	ErrOrphanBlockIsNotAllowed

	// ErrViolatingBoundedMergeDepth indicates that a block is violating finality from
	// its own point of view
	ErrViolatingBoundedMergeDepth

	// ErrViolatingMergeLimit indicates that a block merges more than mergeLimit blocks
	ErrViolatingMergeLimit

	// ErrChainedTransactions indicates that a block contains a transaction that spends an output of a transaction
	// In the same block
	ErrChainedTransactions

	// ErrSelectedParentDisqualifiedFromChain indicates that a block's selectedParent has the status DisqualifiedFromChain
	ErrSelectedParentDisqualifiedFromChain

	// ErrBlockSizeTooHigh indicates the size of a block exceeds the maximum
	// allowed limits.
	ErrBlockSizeTooHigh
)

// Map of ErrorCode values back to their constant names for pretty printing.
var errorCodeStrings = map[ErrorCode]string{
	ErrDuplicateBlock:                      "ErrDuplicateBlock",
	ErrBlockMassTooHigh:                    "ErrBlockMassTooHigh",
	ErrBlockVersionTooOld:                  "ErrBlockVersionTooOld",
	ErrTimeTooOld:                          "ErrTimeTooOld",
	ErrTimeTooNew:                          "ErrTimeTooNew",
	ErrNoParents:                           "ErrNoParents",
	ErrWrongParentsOrder:                   "ErrWrongParentsOrder",
	ErrDifficultyTooLow:                    "ErrDifficultyTooLow",
	ErrUnexpectedDifficulty:                "ErrUnexpectedDifficulty",
	ErrHighHash:                            "ErrHighHash",
	ErrBadMerkleRoot:                       "ErrBadMerkleRoot",
	ErrFinalityPointTimeTooOld:             "ErrFinalityPointTimeTooOld",
	ErrNoTransactions:                      "ErrNoTransactions",
	ErrNoTxInputs:                          "ErrNoTxInputs",
	ErrTxMassTooHigh:                       "ErrTxMassTooHigh",
	ErrBadTxOutValue:                       "ErrBadTxOutValue",
	ErrDuplicateTxInputs:                   "ErrDuplicateTxInputs",
	ErrBadTxInput:                          "ErrBadTxInput",
	ErrMissingTxOut:                        "ErrMissingTxOut",
	ErrDoubleSpendInSameBlock:              "ErrDoubleSpendInSameBlock",
	ErrUnfinalizedTx:                       "ErrUnfinalizedTx",
	ErrDuplicateTx:                         "ErrDuplicateTx",
	ErrOverwriteTx:                         "ErrOverwriteTx",
	ErrImmatureSpend:                       "ErrImmatureSpend",
	ErrSpendTooHigh:                        "ErrSpendTooHigh",
	ErrBadFees:                             "ErrBadFees",
	ErrTooManySigOps:                       "ErrTooManySigOps",
	ErrFirstTxNotCoinbase:                  "ErrFirstTxNotCoinbase",
	ErrMultipleCoinbases:                   "ErrMultipleCoinbases",
	ErrBadCoinbasePayloadLen:               "ErrBadCoinbasePayloadLen",
	ErrBadCoinbaseTransaction:              "ErrBadCoinbaseTransaction",
	ErrScriptMalformed:                     "ErrScriptMalformed",
	ErrScriptValidation:                    "ErrScriptValidation",
	ErrParentBlockUnknown:                  "ErrParentBlockUnknown",
	ErrInvalidAncestorBlock:                "ErrInvalidAncestorBlock",
	ErrParentBlockNotCurrentTips:           "ErrParentBlockNotCurrentTips",
	ErrWithDiff:                            "ErrWithDiff",
	ErrFinality:                            "ErrFinality",
	ErrTransactionsNotSorted:               "ErrTransactionsNotSorted",
	ErrInvalidGas:                          "ErrInvalidGas",
	ErrInvalidPayload:                      "ErrInvalidPayload",
	ErrInvalidPayloadHash:                  "ErrInvalidPayloadHash",
	ErrSubnetworkRegistry:                  "ErrSubnetworkRegistry",
	ErrInvalidParentsRelation:              "ErrInvalidParentsRelation",
	ErrTooManyParents:                      "ErrTooManyParents",
	ErrDelayedBlockIsNotAllowed:            "ErrDelayedBlockIsNotAllowed",
	ErrOrphanBlockIsNotAllowed:             "ErrOrphanBlockIsNotAllowed",
	ErrViolatingBoundedMergeDepth:          "ErrViolatingBoundedMergeDepth",
	ErrSelectedParentDisqualifiedFromChain: "ErrSelectedParentDisqualifiedFromChain",
	ErrChainedTransactions:                 "ErrChainedTransactions",
	ErrBlockSizeTooHigh:                    "ErrBlockSizeTooHigh",
}

// String returns the ErrorCode as a human-readable name.
func (e ErrorCode) String() string {
	if s := errorCodeStrings[e]; s != "" {
		return s
	}
	return fmt.Sprintf("Unknown ErrorCode (%d)", int(e))
}

// RuleError identifies a rule violation. It is used to indicate that
// processing of a block or transaction failed due to one of the many validation
// rules. The caller can use type assertions to determine if a failure was
// specifically due to a rule violation and access the ErrorCode field to
// ascertain the specific reason for the rule violation.
type RuleError struct {
	ErrorCode   ErrorCode // Describes the kind of error
	Description string    // Human readable description of the issue
}

// Error satisfies the error interface and prints human-readable errors.
func (e RuleError) Error() string {
	return e.Description
}

// Errorf formats according to a format specifier and returns the string
// as a RuleError.
func Errorf(code ErrorCode, format string, args ...interface{}) error {
	return errors.WithStack(RuleError{
		ErrorCode:   code,
		Description: fmt.Sprintf(format, args...),
	})
}
