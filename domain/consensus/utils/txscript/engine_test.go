// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package txscript

import (
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

	"github.com/kaspanet/kaspad/app/appmessage"
)

// TestBadPC sets the pc to a deliberately bad result then confirms that Step()
// and Disasm fail correctly.
func TestBadPC(t *testing.T) {
	t.Parallel()

	tests := []struct {
		script, off int
	}{
		{script: 2, off: 0},
		{script: 0, off: 2},
	}

	// tx with almost empty scripts.
	txIns := []*appmessage.TxIn{
		{
			PreviousOutpoint: appmessage.Outpoint{
				TxID: externalapi.DomainTransactionID([32]byte{
					0xc9, 0x97, 0xa5, 0xe5,
					0x6e, 0x10, 0x41, 0x02,
					0xfa, 0x20, 0x9c, 0x6a,
					0x85, 0x2d, 0xd9, 0x06,
					0x60, 0xa2, 0x0b, 0x2d,
					0x9c, 0x35, 0x24, 0x23,
					0xed, 0xce, 0x25, 0x85,
					0x7f, 0xcd, 0x37, 0x04,
				}),
				Index: 0,
			},
			SignatureScript: mustParseShortForm(""),
			Sequence:        4294967295,
		},
	}
	txOuts := []*appmessage.TxOut{{
		Value:        1000000000,
		ScriptPubKey: nil,
	}}
	tx := appmessage.MsgTxToDomainTransaction(appmessage.NewNativeMsgTx(1, txIns, txOuts))
	scriptPubKey := mustParseShortForm("NOP")

	for _, test := range tests {
		vm, err := NewEngine(scriptPubKey, tx, 0, 0, nil)
		if err != nil {
			t.Errorf("Failed to create script: %v", err)
		}

		// set to after all scripts
		vm.scriptIdx = test.script
		vm.scriptOff = test.off

		_, err = vm.Step()
		if err == nil {
			t.Errorf("Step with invalid pc (%v) succeeds!", test)
			continue
		}

		_, err = vm.DisasmPC()
		if err == nil {
			t.Errorf("DisasmPC with invalid pc (%v) succeeds!",
				test)
		}
	}
}

func TestCheckErrorCondition(t *testing.T) {
	tests := []struct {
		script      string
		finalScript bool
		stepCount   int
		expectedErr error
	}{
		{"OP_1", true, 1, nil},
		{"NOP", true, 0, scriptError(ErrScriptUnfinished, "")},
		{"NOP", true, 1, scriptError(ErrEmptyStack, "")},
		{"OP_1 OP_1", true, 2, scriptError(ErrCleanStack, "")},
		{"OP_0", true, 1, scriptError(ErrEvalFalse, "")},
	}

	for i, test := range tests {
		func() {
			txIns := []*appmessage.TxIn{{
				PreviousOutpoint: appmessage.Outpoint{
					TxID: externalapi.DomainTransactionID([32]byte{
						0xc9, 0x97, 0xa5, 0xe5,
						0x6e, 0x10, 0x41, 0x02,
						0xfa, 0x20, 0x9c, 0x6a,
						0x85, 0x2d, 0xd9, 0x06,
						0x60, 0xa2, 0x0b, 0x2d,
						0x9c, 0x35, 0x24, 0x23,
						0xed, 0xce, 0x25, 0x85,
						0x7f, 0xcd, 0x37, 0x04,
					}),
					Index: 0,
				},
				SignatureScript: nil,
				Sequence:        4294967295,
			}}
			txOuts := []*appmessage.TxOut{{
				Value:        1000000000,
				ScriptPubKey: nil,
			}}
			tx := appmessage.MsgTxToDomainTransaction(appmessage.NewNativeMsgTx(1, txIns, txOuts))

			scriptPubKey := mustParseShortForm(test.script)

			vm, err := NewEngine(scriptPubKey, tx, 0, 0, nil)
			if err != nil {
				t.Errorf("TestCheckErrorCondition: %d: failed to create script: %v", i, err)
			}

			for j := 0; j < test.stepCount; j++ {
				_, err = vm.Step()
				if err != nil {
					t.Errorf("TestCheckErrorCondition: %d: failed to execute step No. %d: %v", i, j+1, err)
					return
				}

				if j != test.stepCount-1 {
					err = vm.CheckErrorCondition(false)
					if !IsErrorCode(err, ErrScriptUnfinished) {
						t.Fatalf("TestCheckErrorCondition: %d: got unexepected error %v on %dth iteration",
							i, err, j)
						return
					}
				}
			}

			err = vm.CheckErrorCondition(test.finalScript)
			if e := checkScriptError(err, test.expectedErr); e != nil {
				t.Errorf("TestCheckErrorCondition: %d: %s", i, e)
			}
		}()
	}
}

// TestCheckPubKeyEncoding ensures the internal checkPubKeyEncoding function
// works as expected.
func TestCheckPubKeyEncoding(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		key     []byte
		isValid bool
	}{
		{
			name: "uncompressed ok",
			key: hexToBytes("0411db93e1dcdb8a016b49840f8c53bc1eb68" +
				"a382e97b1482ecad7b148a6909a5cb2e0eaddfb84ccf" +
				"9744464f82e160bfa9b8b64f9d4c03f999b8643f656b" +
				"412a3"),
			isValid: true,
		},
		{
			name: "compressed ok",
			key: hexToBytes("02ce0b14fb842b1ba549fdd675c98075f12e9" +
				"c510f8ef52bd021a9a1f4809d3b4d"),
			isValid: true,
		},
		{
			name: "compressed ok",
			key: hexToBytes("032689c7c2dab13309fb143e0e8fe39634252" +
				"1887e976690b6b47f5b2a4b7d448e"),
			isValid: true,
		},
		{
			name: "hybrid",
			key: hexToBytes("0679be667ef9dcbbac55a06295ce870b07029" +
				"bfcdb2dce28d959f2815b16f81798483ada7726a3c46" +
				"55da4fbfc0e1108a8fd17b448a68554199c47d08ffb1" +
				"0d4b8"),
			isValid: false,
		},
		{
			name:    "empty",
			key:     nil,
			isValid: false,
		},
	}

	vm := Engine{}
	for _, test := range tests {
		err := vm.checkPubKeyEncoding(test.key)
		if err != nil && test.isValid {
			t.Errorf("checkSignatureLength test '%s' failed "+
				"when it should have succeeded: %v", test.name,
				err)
		} else if err == nil && !test.isValid {
			t.Errorf("checkSignatureEncooding test '%s' succeeded "+
				"when it should have failed", test.name)
		}
	}

}

func TestDisasmPC(t *testing.T) {
	t.Parallel()

	// tx with almost empty scripts.
	txIns := []*appmessage.TxIn{{
		PreviousOutpoint: appmessage.Outpoint{
			TxID: externalapi.DomainTransactionID([32]byte{
				0xc9, 0x97, 0xa5, 0xe5,
				0x6e, 0x10, 0x41, 0x02,
				0xfa, 0x20, 0x9c, 0x6a,
				0x85, 0x2d, 0xd9, 0x06,
				0x60, 0xa2, 0x0b, 0x2d,
				0x9c, 0x35, 0x24, 0x23,
				0xed, 0xce, 0x25, 0x85,
				0x7f, 0xcd, 0x37, 0x04,
			}),
			Index: 0,
		},
		SignatureScript: mustParseShortForm("OP_2"),
		Sequence:        4294967295,
	}}
	txOuts := []*appmessage.TxOut{{
		Value:        1000000000,
		ScriptPubKey: nil,
	}}
	tx := appmessage.MsgTxToDomainTransaction(appmessage.NewNativeMsgTx(1, txIns, txOuts))

	scriptPubKey := mustParseShortForm("OP_DROP NOP TRUE")

	vm, err := NewEngine(scriptPubKey, tx, 0, 0, nil)
	if err != nil {
		t.Fatalf("failed to create script: %v", err)
	}

	tests := []struct {
		expected    string
		expectedErr error
	}{
		{"00:0000: OP_2", nil},
		{"01:0000: OP_DROP", nil},
		{"01:0001: OP_NOP", nil},
		{"01:0002: OP_1", nil},
		{"", scriptError(ErrInvalidProgramCounter, "")},
	}

	for i, test := range tests {
		actual, err := vm.DisasmPC()
		if e := checkScriptError(err, test.expectedErr); e != nil {
			t.Errorf("TestDisasmPC: %d: %s", i, e)
		}

		if actual != test.expected {
			t.Errorf("TestDisasmPC: %d: expected: '%s'. Got: '%s'", i, test.expected, actual)
		}

		// ignore results from vm.Step() to keep going even when no opcodes left, to hit error case
		_, _ = vm.Step()
	}
}

func TestDisasmScript(t *testing.T) {
	t.Parallel()

	// tx with almost empty scripts.
	txIns := []*appmessage.TxIn{{
		PreviousOutpoint: appmessage.Outpoint{
			TxID: externalapi.DomainTransactionID([32]byte{
				0xc9, 0x97, 0xa5, 0xe5,
				0x6e, 0x10, 0x41, 0x02,
				0xfa, 0x20, 0x9c, 0x6a,
				0x85, 0x2d, 0xd9, 0x06,
				0x60, 0xa2, 0x0b, 0x2d,
				0x9c, 0x35, 0x24, 0x23,
				0xed, 0xce, 0x25, 0x85,
				0x7f, 0xcd, 0x37, 0x04,
			}),
			Index: 0,
		},
		SignatureScript: mustParseShortForm("OP_2"),
		Sequence:        4294967295,
	}}
	txOuts := []*appmessage.TxOut{{
		Value:        1000000000,
		ScriptPubKey: nil,
	}}
	tx := appmessage.MsgTxToDomainTransaction(appmessage.NewNativeMsgTx(1, txIns, txOuts))
	scriptPubKey := mustParseShortForm("OP_DROP NOP TRUE")

	vm, err := NewEngine(scriptPubKey, tx, 0, 0, nil)
	if err != nil {
		t.Fatalf("failed to create script: %v", err)
	}

	tests := []struct {
		index       int
		expected    string
		expectedErr error
	}{
		{-1, "", scriptError(ErrInvalidIndex, "")},
		{0, "00:0000: OP_2\n", nil},
		{1, "01:0000: OP_DROP\n01:0001: OP_NOP\n01:0002: OP_1\n", nil},
		{2, "", scriptError(ErrInvalidIndex, "")},
	}

	for _, test := range tests {
		actual, err := vm.DisasmScript(test.index)
		if e := checkScriptError(err, test.expectedErr); e != nil {
			t.Errorf("TestDisasmScript: %d: %s", test.index, e)
		}

		if actual != test.expected {
			t.Errorf("TestDisasmScript: %d: expected: '%s'. Got: '%s'", test.index, test.expected, actual)
		}
	}
}
