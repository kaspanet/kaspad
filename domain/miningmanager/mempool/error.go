// Copyright (c) 2014-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package mempool

import (
	"fmt"

	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"

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

// RejectCode represents a numeric value by which a remote peer indicates
// why a message was rejected.
type RejectCode uint8

// These constants define the various supported reject codes.
const (
	RejectMalformed       RejectCode = 0x01
	RejectInvalid         RejectCode = 0x10
	RejectObsolete        RejectCode = 0x11
	RejectDuplicate       RejectCode = 0x12
	RejectNotRequested    RejectCode = 0x13
	RejectNonstandard     RejectCode = 0x40
	RejectDust            RejectCode = 0x41
	RejectInsufficientFee RejectCode = 0x42
	RejectFinality        RejectCode = 0x43
	RejectDifficulty      RejectCode = 0x44
)

// Map of reject codes back strings for pretty printing.
var rejectCodeStrings = map[RejectCode]string{
	RejectMalformed:       "REJECT_MALFORMED",
	RejectInvalid:         "REJECT_INVALID",
	RejectObsolete:        "REJECT_OBSOLETE",
	RejectDuplicate:       "REJECT_DUPLICATE",
	RejectNonstandard:     "REJECT_NONSTANDARD",
	RejectDust:            "REJECT_DUST",
	RejectInsufficientFee: "REJECT_INSUFFICIENTFEE",
	RejectFinality:        "REJECT_FINALITY",
	RejectDifficulty:      "REJECT_DIFFICULTY",
	RejectNotRequested:    "REJECT_NOTREQUESTED",
}

// String returns the RejectCode in human-readable form.
func (code RejectCode) String() string {
	if s, ok := rejectCodeStrings[code]; ok {
		return s
	}

	return fmt.Sprintf("Unknown RejectCode (%d)", uint8(code))
}

// TxRuleError identifies a rule violation. It is used to indicate that
// processing of a transaction failed due to one of the many validation
// rules. The caller can use type assertions to determine if a failure was
// specifically due to a rule violation and access the ErrorCode field to
// ascertain the specific reason for the rule violation.
type TxRuleError struct {
	RejectCode  RejectCode // The code to send with reject messages
	Description string     // Human readable description of the issue
}

// Error satisfies the error interface and prints human-readable errors.
func (e TxRuleError) Error() string {
	return e.Description
}

// txRuleError creates an underlying TxRuleError with the given a set of
// arguments and returns a RuleError that encapsulates it.
func txRuleError(c RejectCode, desc string) RuleError {
	return RuleError{
		Err: TxRuleError{RejectCode: c, Description: desc},
	}
}

// extractRejectCode attempts to return a relevant reject code for a given error
// by examining the error for known types. It will return true if a code
// was successfully extracted.
func extractRejectCode(err error) (RejectCode, bool) {
	// Pull the underlying error out of a RuleError.
	var ruleErr RuleError
	if ok := errors.As(err, &ruleErr); ok {
		err = ruleErr.Err
	}

	var dagRuleErr ruleerrors.RuleError
	if errors.As(err, &dagRuleErr) {
		// Convert the DAG error to a reject code.
		var code RejectCode
		switch dagRuleErr {
		// Rejected due to duplicate.
		case ruleerrors.ErrDuplicateBlock:
			code = RejectDuplicate

		// Rejected due to obsolete version.
		case ruleerrors.ErrBlockVersionTooOld:
			code = RejectObsolete

		// Rejected due to being earlier than the last finality point.
		case ruleerrors.ErrFinalityPointTimeTooOld:
			code = RejectFinality
		case ruleerrors.ErrDifficultyTooLow:
			code = RejectDifficulty

		// Everything else is due to the block or transaction being invalid.
		default:
			code = RejectInvalid
		}

		return code, true
	}

	var trErr TxRuleError
	if errors.As(err, &trErr) {
		return trErr.RejectCode, true
	}

	if err == nil {
		return RejectInvalid, false
	}

	return RejectInvalid, false
}
