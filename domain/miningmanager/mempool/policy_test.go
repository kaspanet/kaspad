// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package mempool

import (
	"bytes"
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"

	consensusexternalapi "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
)

// TestCalcMinRequiredTxRelayFee tests the calcMinRequiredTxRelayFee API.
func TestCalcMinRequiredTxRelayFee(t *testing.T) {
	tests := []struct {
		name     string      // test description.
		size     int64       // Transaction size in bytes.
		relayFee util.Amount // minimum relay transaction fee.
		want     int64       // Expected fee.
	}{
		{
			// Ensure combination of size and fee that are less than 1000
			// produce a non-zero fee.
			"250 bytes with relay fee of 3",
			250,
			3,
			3,
		},
		{
			"100 bytes with default minimum relay fee",
			100,
			DefaultMinRelayTxFee,
			100,
		},
		{
			"max standard tx size with default minimum relay fee",
			MaxStandardTxSize,
			DefaultMinRelayTxFee,
			100000,
		},
		{
			"max standard tx size with max sompi relay fee",
			MaxStandardTxSize,
			util.MaxSompi,
			util.MaxSompi,
		},
		{
			"1500 bytes with 5000 relay fee",
			1500,
			5000,
			7500,
		},
		{
			"1500 bytes with 3000 relay fee",
			1500,
			3000,
			4500,
		},
		{
			"782 bytes with 5000 relay fee",
			782,
			5000,
			3910,
		},
		{
			"782 bytes with 3000 relay fee",
			782,
			3000,
			2346,
		},
		{
			"782 bytes with 2550 relay fee",
			782,
			2550,
			1994,
		},
	}

	for _, test := range tests {
		got := calcMinRequiredTxRelayFee(test.size, test.relayFee)
		if got != test.want {
			t.Errorf("TestCalcMinRequiredTxRelayFee test '%s' "+
				"failed: got %v want %v", test.name, got,
				test.want)
			continue
		}
	}
}

// TestDust tests the isDust API.
func TestDust(t *testing.T) {
	scriptPublicKey := &consensusexternalapi.ScriptPublicKey{
		[]byte{0x76, 0xa9, 0x21, 0x03, 0x2f, 0x7e, 0x43,
			0x0a, 0xa4, 0xc9, 0xd1, 0x59, 0x43, 0x7e, 0x84, 0xb9,
			0x75, 0xdc, 0x76, 0xd9, 0x00, 0x3b, 0xf0, 0x92, 0x2c,
			0xf3, 0xaa, 0x45, 0x28, 0x46, 0x4b, 0xab, 0x78, 0x0d,
			0xba, 0x5e}, 0}

	tests := []struct {
		name     string // test description
		txOut    consensusexternalapi.DomainTransactionOutput
		relayFee util.Amount // minimum relay transaction fee.
		isDust   bool
	}{
		{
			// Any value is allowed with a zero relay fee.
			"zero value with zero relay fee",
			consensusexternalapi.DomainTransactionOutput{Value: 0, ScriptPublicKey: scriptPublicKey},
			0,
			false,
		},
		{
			// Zero value is dust with any relay fee"
			"zero value with very small tx fee",
			consensusexternalapi.DomainTransactionOutput{Value: 0, ScriptPublicKey: scriptPublicKey},
			1,
			true,
		},
		{
			"36 byte public key script with value 605",
			consensusexternalapi.DomainTransactionOutput{Value: 605, ScriptPublicKey: scriptPublicKey},
			1000,
			true,
		},
		{
			"36 byte public key script with value 606",
			consensusexternalapi.DomainTransactionOutput{Value: 606, ScriptPublicKey: scriptPublicKey},
			1000,
			false,
		},
		{
			// Maximum allowed value is never dust.
			"max sompi amount is never dust",
			consensusexternalapi.DomainTransactionOutput{Value: util.MaxSompi, ScriptPublicKey: scriptPublicKey},
			util.MaxSompi,
			false,
		},
		{
			// Maximum int64 value causes overflow.
			"maximum int64 value",
			consensusexternalapi.DomainTransactionOutput{Value: 1<<63 - 1, ScriptPublicKey: scriptPublicKey},
			1<<63 - 1,
			true,
		},
		{
			// Unspendable ScriptPublicKey due to an invalid public key
			// script.
			"unspendable ScriptPublicKey",
			consensusexternalapi.DomainTransactionOutput{Value: 5000, ScriptPublicKey: &consensusexternalapi.ScriptPublicKey{[]byte{0x01}, 0}},
			0, // no relay fee
			true,
		},
	}
	for _, test := range tests {
		res := isDust(&test.txOut, test.relayFee)
		if res != test.isDust {
			t.Fatalf("Dust test '%s' failed: want %v got %v",
				test.name, test.isDust, res)
			continue
		}
	}
}

// TestCheckTransactionStandard tests the checkTransactionStandard API.
func TestCheckTransactionStandard(t *testing.T) {
	// Create some dummy, but otherwise standard, data for transactions.
	prevOutTxID := &consensusexternalapi.DomainTransactionID{}
	dummyPrevOut := consensusexternalapi.DomainOutpoint{TransactionID: *prevOutTxID, Index: 1}
	dummySigScript := bytes.Repeat([]byte{0x00}, 65)
	dummyTxIn := consensusexternalapi.DomainTransactionInput{
		PreviousOutpoint: dummyPrevOut,
		SignatureScript:  dummySigScript,
		Sequence:         constants.MaxTxInSequenceNum,
	}
	addrHash := [32]byte{0x01}
	addr, err := util.NewAddressPubKeyHash(addrHash[:], util.Bech32PrefixKaspaTest)
	if err != nil {
		t.Fatalf("NewAddressPubKeyHash: unexpected error: %v", err)
	}
	dummyScriptPublicKey, err := txscript.PayToAddrScript(addr)
	if err != nil {
		t.Fatalf("PayToAddrScript: unexpected error: %v", err)
	}
	dummyTxOut := consensusexternalapi.DomainTransactionOutput{
		Value:           100000000, // 1 KAS
		ScriptPublicKey: dummyScriptPublicKey,
	}

	tests := []struct {
		name       string
		tx         consensusexternalapi.DomainTransaction
		height     uint64
		isStandard bool
		code       RejectCode
	}{
		{
			name:       "Typical pay-to-pubkey-hash transaction",
			tx:         consensusexternalapi.DomainTransaction{Version: 0, Inputs: []*consensusexternalapi.DomainTransactionInput{&dummyTxIn}, Outputs: []*consensusexternalapi.DomainTransactionOutput{&dummyTxOut}},
			height:     300000,
			isStandard: true,
		},
		{
			name:       "Transaction version too high",
			tx:         consensusexternalapi.DomainTransaction{Version: constants.MaxTransactionVersion + 1, Inputs: []*consensusexternalapi.DomainTransactionInput{&dummyTxIn}, Outputs: []*consensusexternalapi.DomainTransactionOutput{&dummyTxOut}},
			height:     300000,
			isStandard: false,
			code:       RejectNonstandard,
		},

		{
			name: "Transaction size is too large",
			tx: consensusexternalapi.DomainTransaction{Version: 0, Inputs: []*consensusexternalapi.DomainTransactionInput{&dummyTxIn}, Outputs: []*consensusexternalapi.DomainTransactionOutput{{
				Value:           0,
				ScriptPublicKey: &consensusexternalapi.ScriptPublicKey{bytes.Repeat([]byte{0x00}, MaxStandardTxSize+1), 0},
			}}},
			height:     300000,
			isStandard: false,
			code:       RejectNonstandard,
		},
		{
			name: "Signature script size is too large",
			tx: consensusexternalapi.DomainTransaction{Version: 0, Inputs: []*consensusexternalapi.DomainTransactionInput{{
				PreviousOutpoint: dummyPrevOut,
				SignatureScript: bytes.Repeat([]byte{0x00},
					maxStandardSigScriptSize+1),
				Sequence: constants.MaxTxInSequenceNum,
			}}, Outputs: []*consensusexternalapi.DomainTransactionOutput{&dummyTxOut}},
			height:     300000,
			isStandard: false,
			code:       RejectNonstandard,
		},
		{
			name: "Valid but non standard public key script",
			tx: consensusexternalapi.DomainTransaction{Version: 0, Inputs: []*consensusexternalapi.DomainTransactionInput{&dummyTxIn}, Outputs: []*consensusexternalapi.DomainTransactionOutput{{
				Value:           100000000,
				ScriptPublicKey: &consensusexternalapi.ScriptPublicKey{[]byte{txscript.OpTrue}, 0},
			}}},
			height:     300000,
			isStandard: false,
			code:       RejectNonstandard,
		},
		{ //Todo : check on ScriptPublicKey type.
			name: "Dust output",
			tx: consensusexternalapi.DomainTransaction{Version: 0, Inputs: []*consensusexternalapi.DomainTransactionInput{&dummyTxIn}, Outputs: []*consensusexternalapi.DomainTransactionOutput{{
				Value:           0,
				ScriptPublicKey: dummyScriptPublicKey,
			}}},
			height:     300000,
			isStandard: false,
			code:       RejectDust,
		},
		{
			name: "Nulldata transaction",
			tx: consensusexternalapi.DomainTransaction{Version: 0, Inputs: []*consensusexternalapi.DomainTransactionInput{&dummyTxIn}, Outputs: []*consensusexternalapi.DomainTransactionOutput{{
				Value:           0,
				ScriptPublicKey: &consensusexternalapi.ScriptPublicKey{[]byte{txscript.OpReturn}, 0},
			}}},
			height:     300000,
			isStandard: false,
			code:       RejectNonstandard,
		},
	}

	for _, test := range tests {
		// Ensure standardness is as expected.
		err := checkTransactionStandard(&test.tx, &policy{MinRelayTxFee: DefaultMinRelayTxFee, MaxTxVersion: 0})
		if err == nil && test.isStandard {
			// Test passes since function returned standard for a
			// transaction which is intended to be standard.
			continue
		}
		if err == nil && !test.isStandard {
			t.Errorf("checkTransactionStandard (%s): standard when "+
				"it should not be", test.name)
			continue
		}
		if err != nil && test.isStandard {
			t.Errorf("checkTransactionStandard (%s): nonstandard "+
				"when it should not be: %v", test.name, err)
			continue
		}

		// Ensure error type is a TxRuleError inside of a RuleError.
		var ruleErr RuleError
		if !errors.As(err, &ruleErr) {
			t.Errorf("checkTransactionStandard (%s): unexpected "+
				"error type - got %T", test.name, err)
			continue
		}
		txRuleErr, ok := ruleErr.Err.(TxRuleError)
		if !ok {
			t.Errorf("checkTransactionStandard (%s): unexpected "+
				"error type - got %T", test.name, ruleErr.Err)
			continue
		}

		// Ensure the reject code is the expected one.
		if txRuleErr.RejectCode != test.code {
			t.Errorf("checkTransactionStandard (%s): unexpected "+
				"error code - got %v, want %v", test.name,
				txRuleErr.RejectCode, test.code)
			continue
		}
	}
}
