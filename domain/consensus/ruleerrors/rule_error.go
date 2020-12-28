package ruleerrors

import (
	"fmt"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/pkg/errors"
)

// These constants are used to identify a specific RuleError.
var (
	// ErrDuplicateBlock indicates a block with the same hash already
	// exists.
	ErrDuplicateBlock = newRuleError("ErrDuplicateBlock")

	// ErrBlockMassTooHigh indicates the mass of a block exceeds the maximum
	// allowed limits.
	ErrBlockMassTooHigh = newRuleError("ErrBlockMassTooHigh")

	// ErrBlockVersionTooOld indicates the block version is too old and is
	// no longer accepted since the majority of the network has upgraded
	// to a newer version.
	ErrBlockVersionTooOld = newRuleError("ErrBlockVersionTooOld")

	// ErrTimeTooOld indicates the time is either before the median time of
	// the last several blocks per the DAG consensus rules.
	ErrTimeTooOld = newRuleError("ErrTimeTooOld")

	// ErrTimeTooNew indicates the time is too far in the future as compared
	// the current time.
	ErrTimeTooNew = newRuleError("ErrTimeTooNew")

	// ErrNoParents indicates that the block is missing parents
	ErrNoParents = newRuleError("ErrNoParents")

	// ErrDifficultyTooLow indicates the difficulty for the block is lower
	// than the difficulty required.
	ErrDifficultyTooLow = newRuleError("ErrDifficultyTooLow")

	// ErrUnexpectedDifficulty indicates specified bits do not align with
	// the expected value either because it doesn't match the calculated
	// valued based on difficulty regarted rules or it is out of the valid
	// range.
	ErrUnexpectedDifficulty = newRuleError("ErrUnexpectedDifficulty")

	// ErrHighHash indicates the block does not hash to a value which is
	// lower than the required target difficultly.
	ErrHighHash = newRuleError("ErrHighHash")

	// ErrBadMerkleRoot indicates the calculated merkle root does not match
	// the expected value.
	ErrBadMerkleRoot = newRuleError("ErrBadMerkleRoot")

	// ErrBadUTXOCommitment indicates the calculated UTXO commitment does not match
	// the expected value.
	ErrBadUTXOCommitment = newRuleError("ErrBadUTXOCommitment")

	// ErrInvalidSubnetwork indicates the subnetwork is now allowed.
	ErrInvalidSubnetwork = newRuleError("ErrInvalidSubnetwork")

	// ErrFinalityPointTimeTooOld indicates a block has a timestamp before the
	// last finality point.
	ErrFinalityPointTimeTooOld = newRuleError("ErrFinalityPointTimeTooOld")

	// ErrNoTransactions indicates the block does not have a least one
	// transaction. A valid block must have at least the coinbase
	// transaction.
	ErrNoTransactions = newRuleError("ErrNoTransactions")

	// ErrNoTxInputs indicates a transaction does not have any inputs. A
	// valid transaction must have at least one input.
	ErrNoTxInputs = newRuleError("ErrNoTxInputs")

	// ErrTxMassTooHigh indicates the mass of a transaction exceeds the maximum
	// allowed limits.
	ErrTxMassTooHigh = newRuleError("ErrTxMassTooHigh")

	// ErrBadTxOutValue indicates an output value for a transaction is
	// invalid in some way such as being out of range.
	ErrBadTxOutValue = newRuleError("ErrBadTxOutValue")

	// ErrDuplicateTxInputs indicates a transaction references the same
	// input more than once.
	ErrDuplicateTxInputs = newRuleError("ErrDuplicateTxInputs")

	// ErrBadTxInput indicates a transaction input is invalid in some way
	// such as referencing a previous transaction outpoint which is out of
	// range or not referencing one at all.
	ErrBadTxInput = newRuleError("ErrBadTxInput")

	// ErrDoubleSpendInSameBlock indicates a transaction
	// that spends an output that was already spent by another
	// transaction in the same block.
	ErrDoubleSpendInSameBlock = newRuleError("ErrDoubleSpendInSameBlock")

	// ErrUnfinalizedTx indicates a transaction has not been finalized.
	// A valid block may only contain finalized transactions.
	ErrUnfinalizedTx = newRuleError("ErrUnfinalizedTx")

	// ErrDuplicateTx indicates a block contains an identical transaction
	// (or at least two transactions which hash to the same value). A
	// valid block may only contain unique transactions.
	ErrDuplicateTx = newRuleError("ErrDuplicateTx")

	// ErrOverwriteTx indicates a block contains a transaction that has
	// the same hash as a previous transaction which has not been fully
	// spent.
	ErrOverwriteTx = newRuleError("ErrOverwriteTx")

	// ErrImmatureSpend indicates a transaction is attempting to spend a
	// coinbase that has not yet reached the required maturity.
	ErrImmatureSpend = newRuleError("ErrImmatureSpend")

	// ErrSpendTooHigh indicates a transaction is attempting to spend more
	// value than the sum of all of its inputs.
	ErrSpendTooHigh = newRuleError("ErrSpendTooHigh")

	// ErrBadFees indicates the total fees for a block are invalid due to
	// exceeding the maximum possible value.
	ErrBadFees = newRuleError("ErrBadFees")

	// ErrTooManySigOps indicates the total number of signature operations
	// for a transaction or block exceed the maximum allowed limits.
	ErrTooManySigOps = newRuleError("ErrTooManySigOps")

	// ErrFirstTxNotCoinbase indicates the first transaction in a block
	// is not a coinbase transaction.
	ErrFirstTxNotCoinbase = newRuleError("ErrFirstTxNotCoinbase")

	// ErrMultipleCoinbases indicates a block contains more than one
	// coinbase transaction.
	ErrMultipleCoinbases = newRuleError("ErrMultipleCoinbases")

	// ErrBadCoinbasePayloadLen indicates the length of the payload
	// for a coinbase transaction is too high.
	ErrBadCoinbasePayloadLen = newRuleError("ErrBadCoinbasePayloadLen")

	// ErrBadCoinbaseTransaction indicates that the block's coinbase transaction is not build as expected
	ErrBadCoinbaseTransaction = newRuleError("ErrBadCoinbaseTransaction")

	// ErrScriptMalformed indicates a transaction script is malformed in
	// some way. For example, it might be longer than the maximum allowed
	// length or fail to parse.
	ErrScriptMalformed = newRuleError("ErrScriptMalformed")

	// ErrScriptValidation indicates the result of executing transaction
	// script failed. The error covers any failure when executing scripts
	// such signature verification failures and execution past the end of
	// the stack.
	ErrScriptValidation = newRuleError("ErrScriptValidation")

	// ErrParentBlockUnknown indicates that the parent block is not known.
	ErrParentBlockUnknown = newRuleError("ErrParentBlockUnknown")

	// ErrInvalidAncestorBlock indicates that an ancestor of this block has
	// already failed validation.
	ErrInvalidAncestorBlock = newRuleError("ErrInvalidAncestorBlock")

	// ErrParentBlockNotCurrentTips indicates that the block's parents are not the
	// current tips. This is not a block validation rule, but is required
	// for block proposals submitted via getblocktemplate RPC.
	ErrParentBlockNotCurrentTips = newRuleError("ErrParentBlockNotCurrentTips")

	// ErrWithDiff indicates that there was an error with UTXOSet.WithDiff
	ErrWithDiff = newRuleError("ErrWithDiff")

	// ErrFinality indicates that a block doesn't adhere to the finality rules
	ErrFinality = newRuleError("ErrFinality")

	// ErrTransactionsNotSorted indicates that transactions in block are not
	// sorted by subnetwork
	ErrTransactionsNotSorted = newRuleError("ErrTransactionsNotSorted")

	// ErrInvalidGas transaction wants to use more GAS than allowed
	// by subnetwork
	ErrInvalidGas = newRuleError("ErrInvalidGas")

	// ErrInvalidPayload transaction includes a payload in a subnetwork that doesn't allow
	// a Payload
	ErrInvalidPayload = newRuleError("ErrInvalidPayload")

	// ErrInvalidPayloadHash invalid hash of transaction's payload
	ErrInvalidPayloadHash = newRuleError("ErrInvalidPayloadHash")

	// ErrSubnetwork indicates that a block doesn't adhere to the subnetwork
	// registry rules
	ErrSubnetworkRegistry = newRuleError("ErrSubnetworkRegistry")

	// ErrInvalidParentsRelation indicates that one of the parents of a block
	// is also an ancestor of another parent
	ErrInvalidParentsRelation = newRuleError("ErrInvalidParentsRelation")

	// ErrTooManyParents indicates that a block points to more then `MaxNumParentBlocks` parents
	ErrTooManyParents = newRuleError("ErrTooManyParents")

	// ErrDelayedBlockIsNotAllowed indicates that a block with a delayed timestamp was
	// submitted with BFDisallowDelay flag raised.
	ErrDelayedBlockIsNotAllowed = newRuleError("ErrDelayedBlockIsNotAllowed")

	// ErrOrphanBlockIsNotAllowed indicates that an orphan block was submitted with
	// BFDisallowOrphans flag raised.
	ErrOrphanBlockIsNotAllowed = newRuleError("ErrOrphanBlockIsNotAllowed")

	// ErrViolatingBoundedMergeDepth indicates that a block is violating finality from
	// its own point of view
	ErrViolatingBoundedMergeDepth = newRuleError("ErrViolatingBoundedMergeDepth")

	// ErrViolatingMergeLimit indicates that a block merges more than mergeLimit blocks
	ErrViolatingMergeLimit = newRuleError("ErrViolatingMergeLimit")

	// ErrChainedTransactions indicates that a block contains a transaction that spends an output of a transaction
	// In the same block
	ErrChainedTransactions = newRuleError("ErrChainedTransactions")

	// ErrSelectedParentDisqualifiedFromChain indicates that a block's selectedParent has the status DisqualifiedFromChain
	ErrSelectedParentDisqualifiedFromChain = newRuleError("ErrSelectedParentDisqualifiedFromChain")

	// ErrBlockSizeTooHigh indicates the size of a block exceeds the maximum
	// allowed limits.
	ErrBlockSizeTooHigh = newRuleError("ErrBlockSizeTooHigh")

	// ErrBuiltInTransactionHasGas indicates that a transaction with built in subnetwork ID has a non zero gas.
	ErrBuiltInTransactionHasGas = newRuleError("ErrBuiltInTransactionHasGas")

	ErrKnownInvalid = newRuleError("ErrKnownInvalid")

	ErrSubnetworksDisabled    = newRuleError("ErrSubnetworksDisabled")
	ErrBadPruningPointUTXOSet = newRuleError("ErrBadPruningPointUTXOSet")

	ErrMalformedUTXO = newRuleError("ErrMalformedUTXO")

	ErrWrongPruningPointHash = newRuleError("ErrWrongPruningPointHash")

	//ErrPruningPointViolation indicates that the pruning point isn't in the block past.
	ErrPruningPointViolation = newRuleError("ErrPruningPointViolation")

	//ErrBlockIsTooMuchInTheFuture indicates that the block timestamp is too much in the future.
	ErrBlockIsTooMuchInTheFuture = newRuleError("ErrBlockIsTooMuchInTheFuture")

	//ErrBlockVersionIsUnknown indicates that the block version is unknown.
	ErrBlockVersionIsUnknown = newRuleError("ErrBlockVersionIsUnknown")
)

// RuleError identifies a rule violation. It is used to indicate that
// processing of a block or transaction failed due to one of the many validation
// rules. The caller can use type assertions to determine if a failure was
// specifically due to a rule violation.
type RuleError struct {
	message string
	inner   error
}

// Error satisfies the error interface and prints human-readable errors.
func (e RuleError) Error() string {
	if e.inner != nil {
		return e.message + ": " + e.inner.Error()
	}
	return e.message
}

// Unwrap satisfies the errors.Unwrap interface
func (e RuleError) Unwrap() error {
	return e.inner
}

// Cause satisfies the github.com/pkg/errors.Cause interface
func (e RuleError) Cause() error {
	return e.inner
}

func newRuleError(message string) RuleError {
	return RuleError{message: message, inner: nil}
}

// ErrMissingTxOut indicates a transaction output referenced by an input
// either does not exist or has already been spent.
type ErrMissingTxOut struct {
	MissingOutpoints []*externalapi.DomainOutpoint
}

func (e ErrMissingTxOut) Error() string {
	return fmt.Sprintf("missing the following outpoint: %v", e.MissingOutpoints)
}

// NewErrMissingTxOut Creates a new ErrMissingTxOut error wrapped in a RuleError
func NewErrMissingTxOut(missingOutpoints []*externalapi.DomainOutpoint) error {
	return errors.WithStack(RuleError{
		message: "ErrMissingTxOut",
		inner:   ErrMissingTxOut{missingOutpoints},
	})
}

// ErrMissingParents indicates a block points to unknown parent(s).
type ErrMissingParents struct {
	MissingParentHashes []*externalapi.DomainHash
}

func (e ErrMissingParents) Error() string {
	return fmt.Sprintf("missing the following parent hashes: %v", e.MissingParentHashes)
}

// NewErrMissingParents creates a new ErrMissingParents error wrapped in a RuleError
func NewErrMissingParents(missingParentHashes []*externalapi.DomainHash) error {
	return errors.WithStack(RuleError{
		message: "ErrMissingParents",
		inner:   ErrMissingParents{missingParentHashes},
	})
}

// InvalidTransaction is a struct containing an invalid transaction, and the error explaining why it's invalid.
type InvalidTransaction struct {
	Transaction *externalapi.DomainTransaction
	err         error
}

func (invalid InvalidTransaction) String() string {
	return fmt.Sprintf("(%v: %s)", consensushashing.TransactionID(invalid.Transaction), invalid.err)
}

// ErrInvalidTransactionsInNewBlock indicates that some transactions in a new block are invalid
type ErrInvalidTransactionsInNewBlock struct {
	InvalidTransactions []InvalidTransaction
}

func (e ErrInvalidTransactionsInNewBlock) Error() string {
	return fmt.Sprint(e.InvalidTransactions)
}

// NewErrInvalidTransactionsInNewBlock Creates a new ErrInvalidTransactionsInNewBlock error wrapped in a RuleError
func NewErrInvalidTransactionsInNewBlock(invalidTransactions []InvalidTransaction) error {
	return errors.WithStack(RuleError{
		message: "ErrInvalidTransactionsInNewBlock",
		inner:   ErrInvalidTransactionsInNewBlock{invalidTransactions},
	})
}
