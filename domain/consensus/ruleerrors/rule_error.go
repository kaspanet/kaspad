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

	// ErrBlockVersionTooOld indicates the block version is too old and is
	// no longer accepted since the majority of the network has upgraded
	// to a newer version.
	ErrBlockVersionTooOld = newRuleError("ErrBlockVersionTooOld")

	// ErrTimeTooOld indicates the time is either before the median time of
	// the last several blocks per the DAG consensus rules.
	ErrTimeTooOld = newRuleError("ErrTimeTooOld")

	//ErrTimeTooMuchInTheFuture indicates that the block timestamp is too much in the future.
	ErrTimeTooMuchInTheFuture = newRuleError("ErrTimeTooMuchInTheFuture")

	// ErrNoParents indicates that the block is missing parents
	ErrNoParents = newRuleError("ErrNoParents")

	// ErrUnexpectedDifficulty indicates specified bits do not align with
	// the expected value either because it doesn't match the calculated
	// valued based on difficulty regarted rules.
	ErrUnexpectedDifficulty = newRuleError("ErrUnexpectedDifficulty")

	// ErrUnexpectedDAAScore indicates specified DAA score does not align with
	// the expected value.
	ErrUnexpectedDAAScore = newRuleError("ErrUnexpectedDAAScore")

	// ErrUnexpectedBlueWork indicates specified blue work does not align with
	// the expected value.
	ErrUnexpectedBlueWork = newRuleError("ErrUnexpectedBlueWork")

	// ErrUnexpectedFinalityPoint indicates specified finality point does not align with
	// the expected value.
	ErrUnexpectedFinalityPoint = newRuleError("ErrUnexpectedFinalityPoint")

	// ErrUnexpectedBlueScore indicates specified blue score does not align with
	// the expected value.
	ErrUnexpectedBlueScore = newRuleError("ErrUnexpectedBlueScore")

	// ErrTargetTooHigh indicates specified bits do not align with
	// the expected value either because it is above the valid
	// range.
	ErrTargetTooHigh = newRuleError("ErrTargetTooHigh")

	// ErrUnexpectedDifficulty indicates specified bits do not align with
	// the expected value either because it is negative.
	ErrNegativeTarget = newRuleError("ErrNegativeTarget")

	// ErrInvalidPoW indicates that the block proof-of-work is invalid.
	ErrInvalidPoW = newRuleError("ErrInvalidPoW")

	// ErrBadMerkleRoot indicates the calculated merkle root does not match
	// the expected value.
	ErrBadMerkleRoot = newRuleError("ErrBadMerkleRoot")

	// ErrBadUTXOCommitment indicates the calculated UTXO commitment does not match
	// the expected value.
	ErrBadUTXOCommitment = newRuleError("ErrBadUTXOCommitment")

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

	// ErrBadTxOutValue indicates an output value for a transaction is
	// invalid in some way such as being out of range.
	ErrBadTxOutValue = newRuleError("ErrBadTxOutValue")

	// ErrDuplicateTxInputs indicates a transaction references the same
	// input more than once.
	ErrDuplicateTxInputs = newRuleError("ErrDuplicateTxInputs")

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

	// ErrImmatureSpend indicates a transaction is attempting to spend a
	// coinbase that has not yet reached the required maturity.
	ErrImmatureSpend = newRuleError("ErrImmatureSpend")

	// ErrSpendTooHigh indicates a transaction is attempting to spend more
	// value than the sum of all of its inputs.
	ErrSpendTooHigh = newRuleError("ErrSpendTooHigh")

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

	// ErrInvalidAncestorBlock indicates that an ancestor of this block has
	// already failed validation.
	ErrInvalidAncestorBlock = newRuleError("ErrInvalidAncestorBlock")

	// ErrTransactionsNotSorted indicates that transactions in block are not
	// sorted by subnetwork
	ErrTransactionsNotSorted = newRuleError("ErrTransactionsNotSorted")

	// ErrInvalidGas transaction wants to use more GAS than allowed
	// by subnetwork
	ErrInvalidGas = newRuleError("ErrInvalidGas")

	// ErrInvalidPayload transaction includes a payload in a subnetwork that doesn't allow
	// a Payload
	ErrInvalidPayload = newRuleError("ErrInvalidPayload")

	// ErrWrongSigOpCount transaction input specifies an incorrect SigOpCount
	ErrWrongSigOpCount = newRuleError("ErrWrongSigOpCount")

	// ErrSubnetwork indicates that a block doesn't adhere to the subnetwork
	// registry rules
	ErrSubnetworkRegistry = newRuleError("ErrSubnetworkRegistry")

	// ErrInvalidParentsRelation indicates that one of the parents of a block
	// is also an ancestor of another parent
	ErrInvalidParentsRelation = newRuleError("ErrInvalidParentsRelation")

	// ErrTooManyParents indicates that a block points to more then `MaxNumParentBlocks` parents
	ErrTooManyParents = newRuleError("ErrTooManyParents")

	// ErrViolatingBoundedMergeDepth indicates that a block is violating finality from
	// its own point of view
	ErrViolatingBoundedMergeDepth = newRuleError("ErrViolatingBoundedMergeDepth")

	// ErrViolatingMergeLimit indicates that a block merges more than mergeLimit blocks
	ErrViolatingMergeLimit = newRuleError("ErrViolatingMergeLimit")

	// ErrChainedTransactions indicates that a block contains a transaction that spends an output of a transaction
	// In the same block
	ErrChainedTransactions = newRuleError("ErrChainedTransactions")

	// ErrBlockMassTooHigh indicates the mass of a block exceeds the maximum
	// allowed limits.
	ErrBlockMassTooHigh = newRuleError("ErrBlockMassTooHigh")

	ErrKnownInvalid = newRuleError("ErrKnownInvalid")

	ErrSubnetworksDisabled    = newRuleError("ErrSubnetworksDisabled")
	ErrBadPruningPointUTXOSet = newRuleError("ErrBadPruningPointUTXOSet")

	ErrMalformedUTXO = newRuleError("ErrMalformedUTXO")

	ErrWrongPruningPointHash = newRuleError("ErrWrongPruningPointHash")

	//ErrPruningPointViolation indicates that the pruning point isn't in the block past.
	ErrPruningPointViolation = newRuleError("ErrPruningPointViolation")

	ErrUnexpectedPruningPoint = newRuleError("ErrUnexpectedPruningPoint")

	ErrSuggestedPruningViolatesFinality = newRuleError("ErrSuggestedPruningViolatesFinality")

	//ErrBlockVersionIsUnknown indicates that the block version is unknown.
	ErrBlockVersionIsUnknown = newRuleError("ErrBlockVersionIsUnknown")

	//ErrTransactionVersionIsUnknown indicates that the transaction version is unknown.
	ErrTransactionVersionIsUnknown = newRuleError("ErrTransactionVersionIsUnknown")

	// ErrPrunedBlock indicates that the block currently being validated had already been pruned.
	ErrPrunedBlock = newRuleError("ErrPrunedBlock")

	ErrGetVirtualUTXOsWrongVirtualParents = newRuleError("ErrGetVirtualUTXOsWrongVirtualParents")

	ErrVirtualGenesisParent = newRuleError("ErrVirtualGenesisParent")

	ErrGenesisOnInitializedConsensus = newRuleError("ErrGenesisOnInitializedConsensus")

	ErrPruningPointSelectedChildDisqualifiedFromChain = newRuleError("ErrPruningPointSelectedChildDisqualifiedFromChain")
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
	Error       error
}

func (invalid InvalidTransaction) String() string {
	return fmt.Sprintf("(%v: %s)", consensushashing.TransactionID(invalid.Transaction), invalid.Error)
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
