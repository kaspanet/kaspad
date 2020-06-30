// Copyright (c) 2014-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"testing"
)

// TestErrorCodeStringer tests the stringized output for the ErrorCode type.
func TestErrorCodeStringer(t *testing.T) {
	tests := []struct {
		in   ErrorCode
		want string
	}{
		{ErrDuplicateBlock, "ErrDuplicateBlock"},
		{ErrBlockMassTooHigh, "ErrBlockMassTooHigh"},
		{ErrBlockVersionTooOld, "ErrBlockVersionTooOld"},
		{ErrTimeTooOld, "ErrTimeTooOld"},
		{ErrTimeTooNew, "ErrTimeTooNew"},
		{ErrNoParents, "ErrNoParents"},
		{ErrWrongParentsOrder, "ErrWrongParentsOrder"},
		{ErrDifficultyTooLow, "ErrDifficultyTooLow"},
		{ErrUnexpectedDifficulty, "ErrUnexpectedDifficulty"},
		{ErrHighHash, "ErrHighHash"},
		{ErrBadMerkleRoot, "ErrBadMerkleRoot"},
		{ErrFinalityPointTimeTooOld, "ErrFinalityPointTimeTooOld"},
		{ErrNoTransactions, "ErrNoTransactions"},
		{ErrNoTxInputs, "ErrNoTxInputs"},
		{ErrTxMassTooHigh, "ErrTxMassTooHigh"},
		{ErrBadTxOutValue, "ErrBadTxOutValue"},
		{ErrDuplicateTxInputs, "ErrDuplicateTxInputs"},
		{ErrBadTxInput, "ErrBadTxInput"},
		{ErrMissingTxOut, "ErrMissingTxOut"},
		{ErrUnfinalizedTx, "ErrUnfinalizedTx"},
		{ErrDuplicateTx, "ErrDuplicateTx"},
		{ErrOverwriteTx, "ErrOverwriteTx"},
		{ErrImmatureSpend, "ErrImmatureSpend"},
		{ErrSpendTooHigh, "ErrSpendTooHigh"},
		{ErrBadFees, "ErrBadFees"},
		{ErrTooManySigOps, "ErrTooManySigOps"},
		{ErrFirstTxNotCoinbase, "ErrFirstTxNotCoinbase"},
		{ErrMultipleCoinbases, "ErrMultipleCoinbases"},
		{ErrBadCoinbasePayloadLen, "ErrBadCoinbasePayloadLen"},
		{ErrBadCoinbaseTransaction, "ErrBadCoinbaseTransaction"},
		{ErrScriptMalformed, "ErrScriptMalformed"},
		{ErrScriptValidation, "ErrScriptValidation"},
		{ErrParentBlockUnknown, "ErrParentBlockUnknown"},
		{ErrInvalidAncestorBlock, "ErrInvalidAncestorBlock"},
		{ErrParentBlockNotCurrentTips, "ErrParentBlockNotCurrentTips"},
		{ErrWithDiff, "ErrWithDiff"},
		{ErrFinality, "ErrFinality"},
		{ErrTransactionsNotSorted, "ErrTransactionsNotSorted"},
		{ErrInvalidGas, "ErrInvalidGas"},
		{ErrInvalidPayload, "ErrInvalidPayload"},
		{ErrInvalidPayloadHash, "ErrInvalidPayloadHash"},
		{ErrInvalidParentsRelation, "ErrInvalidParentsRelation"},
		{ErrDelayedBlockIsNotAllowed, "ErrDelayedBlockIsNotAllowed"},
		{0xffff, "Unknown ErrorCode (65535)"},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		result := test.in.String()
		if result != test.want {
			t.Errorf("String #%d\n got: %s want: %s", i, result,
				test.want)
			continue
		}
	}
}

// TestRuleError tests the error output for the RuleError type.
func TestRuleError(t *testing.T) {
	tests := []struct {
		in   RuleError
		want string
	}{
		{
			RuleError{Description: "duplicate block"},
			"duplicate block",
		},
		{
			RuleError{Description: "human-readable error"},
			"human-readable error",
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		result := test.in.Error()
		if result != test.want {
			t.Errorf("Error #%d\n got: %s want: %s", i, result,
				test.want)
			continue
		}
	}
}
