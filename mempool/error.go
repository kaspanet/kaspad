// Copyright (c) 2014-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package mempool

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

// RuleError identifies a rule violation. It is used to indicate that
// processing of a transaction failed due to one of the many validation
// rules. The caller can use type assertions to determine if a failure was
// specifically due to a rule violation and use the Err field to access the
// underlying error, which will be either a TxRuleError or a
// blockdag.RuleError.
type RuleError struct {
	Err error
}

// Error satisfies the error interface and prints human-readable errors.
func (e RuleError) Error() string {
	if e.Err == nil {
		return "<nil>"
	}
	return e.Err.Error()
}

// TxRuleError identifies a rule violation. It is used to indicate that
// processing of a transaction failed due to one of the many validation
// rules. The caller can use type assertions to determine if a failure was
// specifically due to a rule violation and access the ErrorCode field to
// ascertain the specific reason for the rule violation.
type TxRuleError struct {
	RejectCode  wire.RejectCode // The code to send with reject messages
	Description string          // Human readable description of the issue
}

// Error satisfies the error interface and prints human-readable errors.
func (e TxRuleError) Error() string {
	return e.Description
}

// txRuleError creates an underlying TxRuleError with the given a set of
// arguments and returns a RuleError that encapsulates it.
func txRuleError(c wire.RejectCode, desc string) RuleError {
	return RuleError{
		Err: TxRuleError{RejectCode: c, Description: desc},
	}
}

// dagRuleError returns a RuleError that encapsulates the given
// blockdag.RuleError.
func dagRuleError(dagErr blockdag.RuleError) RuleError {
	return RuleError{
		Err: dagErr,
	}
}

// extractRejectCode attempts to return a relevant reject code for a given error
// by examining the error for known types. It will return true if a code
// was successfully extracted.
func extractRejectCode(err error) (wire.RejectCode, bool) {
	// Pull the underlying error out of a RuleError.
	var ruleErr RuleError
	if ok := errors.As(err, &ruleErr); ok {
		err = ruleErr.Err
	}

	var dagRuleErr blockdag.RuleError
	if errors.As(err, &dagRuleErr) {
		// Convert the DAG error to a reject code.
		var code wire.RejectCode
		switch dagRuleErr.ErrorCode {
		// Rejected due to duplicate.
		case blockdag.ErrDuplicateBlock:
			code = wire.RejectDuplicate

		// Rejected due to obsolete version.
		case blockdag.ErrBlockVersionTooOld:
			code = wire.RejectObsolete

		// Rejected due to being earlier than the last finality point.
		case blockdag.ErrFinalityPointTimeTooOld:
			code = wire.RejectFinality
		case blockdag.ErrDifficultyTooLow:
			code = wire.RejectDifficulty

		// Everything else is due to the block or transaction being invalid.
		default:
			code = wire.RejectInvalid
		}

		return code, true
	}

	var trErr TxRuleError
	if errors.As(err, &trErr) {
		return trErr.RejectCode, true
	}

	if err == nil {
		return wire.RejectInvalid, false
	}

	return wire.RejectInvalid, false
}

// ErrToRejectErr examines the underlying type of the error and returns a reject
// code and string appropriate to be sent in a wire.MsgReject message.
func ErrToRejectErr(err error) (wire.RejectCode, string) {
	// Return the reject code along with the error text if it can be
	// extracted from the error.
	rejectCode, found := extractRejectCode(err)
	if found {
		return rejectCode, err.Error()
	}

	// Return a generic rejected string if there is no error. This really
	// should not happen unless the code elsewhere is not setting an error
	// as it should be, but it's best to be safe and simply return a generic
	// string rather than allowing the following code that dereferences the
	// err to panic.
	if err == nil {
		return wire.RejectInvalid, "rejected"
	}

	// When the underlying error is not one of the above cases, just return
	// wire.RejectInvalid with a generic rejected string plus the error
	// text.
	return wire.RejectInvalid, "rejected: " + err.Error()
}
