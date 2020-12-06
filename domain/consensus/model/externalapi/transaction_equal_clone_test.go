package externalapi

import (
	"reflect"
	"testing"
)

type TransactionToCompare struct {
	tx             *DomainTransaction
	expectedResult bool
}

type TestDomainTransactionStruct struct {
	baseTx                 *DomainTransaction
	transactionToCompareTo []*TransactionToCompare
}

type TransactionInputToCompare struct {
	tx             *DomainTransactionInput
	expectedResult bool
}

type TestDomainTransactionInputStruct struct {
	baseTx                      *DomainTransactionInput
	transactionInputToCompareTo []*TransactionInputToCompare
}

type TransactionOutputToCompare struct {
	tx             *DomainTransactionOutput
	expectedResult bool
}

type TestDomainTransactionOutputStruct struct {
	baseTx                       *DomainTransactionOutput
	transactionOutputToCompareTo []*TransactionOutputToCompare
}

func initTestBaseTransaction() *DomainTransaction {

	testTx := &DomainTransaction{
		Version:      1,
		Inputs:       []*DomainTransactionInput{},
		Outputs:      []*DomainTransactionOutput{},
		LockTime:     1,
		SubnetworkID: DomainSubnetworkID{0x01},
		Gas:          1,
		PayloadHash: DomainHash{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		Payload: []byte{0x01},
		Fee:     0,
		Mass:    1,
		ID: &DomainTransactionID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02},
	}
	return testTx
}

func initTestTransactionToCompare() []*TransactionToCompare {

	testTx := []*TransactionToCompare{{
		tx: &DomainTransaction{
			Version:      1,
			Inputs:       []*DomainTransactionInput{},
			Outputs:      []*DomainTransactionOutput{},
			LockTime:     1,
			SubnetworkID: DomainSubnetworkID{0x01},
			Gas:          1,
			PayloadHash: DomainHash{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
			Payload: []byte{0x01},
			Fee:     0,
			Mass:    1,
			ID: &DomainTransactionID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
		},
		expectedResult: false,
	}, {
		tx: &DomainTransaction{
			Version: 1,
			Inputs:  []*DomainTransactionInput{},
			Outputs: []*DomainTransactionOutput{{uint64(0xFFFF),
				[]byte{1, 2}}, {}},
			LockTime:     1,
			SubnetworkID: DomainSubnetworkID{0x01},
			Gas:          1,
			PayloadHash: DomainHash{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			Payload: []byte{0x01},
			Fee:     0,
			Mass:    1,
			ID: &DomainTransactionID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02},
		},
		expectedResult: false,
	},
		{
			tx: &DomainTransaction{
				Version:      1,
				Inputs:       []*DomainTransactionInput{},
				Outputs:      []*DomainTransactionOutput{},
				LockTime:     1,
				SubnetworkID: DomainSubnetworkID{0x01, 0x02},
				Gas:          1,
				PayloadHash: DomainHash{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
				Payload: []byte{0x01},
				Fee:     0,
				Mass:    1,
				ID: &DomainTransactionID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02},
			},
			expectedResult: false,
		}, {
			tx: &DomainTransaction{
				Version:      1,
				Inputs:       []*DomainTransactionInput{},
				Outputs:      []*DomainTransactionOutput{},
				LockTime:     1,
				SubnetworkID: DomainSubnetworkID{0x01, 0x02},
				Gas:          1,
				PayloadHash: DomainHash{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
				Payload: []byte{0x01},
				Fee:     0,
				Mass:    1,
				ID: &DomainTransactionID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02},
			},
			expectedResult: false,
		}, {
			tx: &DomainTransaction{
				Version:      1,
				Inputs:       []*DomainTransactionInput{},
				Outputs:      []*DomainTransactionOutput{},
				LockTime:     1,
				SubnetworkID: DomainSubnetworkID{0x01, 0x02},
				Gas:          1,
				PayloadHash: DomainHash{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
				Payload: []byte{0x01, 0x02},
				Fee:     0,
				Mass:    1,
				ID: &DomainTransactionID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02},
			},
			expectedResult: false,
		}, {
			tx: &DomainTransaction{
				Version:      1,
				Inputs:       []*DomainTransactionInput{},
				Outputs:      []*DomainTransactionOutput{},
				LockTime:     1,
				SubnetworkID: DomainSubnetworkID{0x01},
				Gas:          1,
				PayloadHash: DomainHash{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
				Payload: []byte{0x01},
				Fee:     0,
				Mass:    1,
				ID: &DomainTransactionID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02},
			},
			expectedResult: true,
		},
	}

	return testTx

}

func initTestDomainTransactionForClone() []*DomainTransaction {

	tests := []*DomainTransaction{
		{
			Version:      1,
			Inputs:       []*DomainTransactionInput{},
			Outputs:      []*DomainTransactionOutput{},
			LockTime:     1,
			SubnetworkID: DomainSubnetworkID{0x01},
			Gas:          1,
			PayloadHash: DomainHash{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			Payload: []byte{0x01},
			Fee:     0,
			Mass:    1,
			ID: &DomainTransactionID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02},
		}, {
			Version:      1,
			Inputs:       []*DomainTransactionInput{},
			Outputs:      []*DomainTransactionOutput{},
			LockTime:     1,
			SubnetworkID: DomainSubnetworkID{0x01},
			Gas:          1,
			PayloadHash: DomainHash{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
			Payload: []byte{0x01},
			Fee:     0,
			Mass:    1,
			ID:      &DomainTransactionID{},
		},
	}
	return tests
}

func initTestDomainTransactionForEqual() []TestDomainTransactionStruct {

	tests := []TestDomainTransactionStruct{
		{
			baseTx:                 initTestBaseTransaction(),
			transactionToCompareTo: initTestTransactionToCompare(),
		},
	}
	return tests
}

func initTestBaseDomainTransactionInput() *DomainTransactionInput {
	basetxInput := &DomainTransactionInput{
		DomainOutpoint{DomainTransactionID{0x01}, 0xFFFF},
		[]byte{1, 2, 3},
		uint64(0xFFFFFFFF),
		&UTXOEntry{1,
			[]byte{0, 1, 2, 3},
			2,
			true},
	}
	return basetxInput
}

func initTestDomainTxInputToCompare() []*TransactionInputToCompare {
	txInput := []*TransactionInputToCompare{{
		tx: &DomainTransactionInput{
			DomainOutpoint{DomainTransactionID{0x01}, 0xFFFF},
			[]byte{1, 2, 3},
			uint64(0xFFFFFFFF),
			&UTXOEntry{1,
				[]byte{0, 1, 2, 3},
				2,
				true},
		},
		expectedResult: true,
	}, {
		tx: &DomainTransactionInput{
			DomainOutpoint{DomainTransactionID{0x01}, 0xFFFF},
			[]byte{1, 2, 3},
			uint64(0xFFFFFFFF),
			&UTXOEntry{1,
				[]byte{0, 1, 2, 3},
				2,
				false},
		},
		expectedResult: false,
	}, {
		tx: &DomainTransactionInput{
			DomainOutpoint{DomainTransactionID{0x01}, 0xFFFF},
			[]byte{1, 2, 3},
			uint64(0xFFFFFFF0),
			&UTXOEntry{1,
				[]byte{0, 1, 2, 3},
				2,
				true},
		},
		expectedResult: false,
	}}
	return txInput

}

func initTestDomainTransactionInputForClone() []*DomainTransactionInput {
	txInput := []*DomainTransactionInput{
		{
			DomainOutpoint{DomainTransactionID{0x01}, 0xFFFF},
			[]byte{1, 2, 3},
			uint64(0xFFFFFFFF),
			&UTXOEntry{1,
				[]byte{0, 1, 2, 3},
				2,
				true},
		}, {

			DomainOutpoint{DomainTransactionID{0x01}, 0xFFFF},
			[]byte{1, 2, 3},
			uint64(0xFFFFFFFF),
			&UTXOEntry{1,
				[]byte{0, 1, 2, 3},
				2,
				false},
		}, {

			DomainOutpoint{DomainTransactionID{0x01}, 0xFFFF},
			[]byte{1, 2, 3},
			uint64(0xFFFFFFF0),
			&UTXOEntry{1,
				[]byte{0, 1, 2, 3},
				2,
				true},
		}}
	return txInput
}

func initTestBaseDomainTransactionOutput() *DomainTransactionOutput {
	basetxOutput := &DomainTransactionOutput{
		0xFFFFFFFF,
		[]byte{0xFF, 0xFF},
	}
	return basetxOutput
}

func initTestDomainTxOutputToCompare() []*TransactionOutputToCompare {
	txInput := []*TransactionOutputToCompare{{
		tx: &DomainTransactionOutput{
			0xFFFFFFFF,
			[]byte{0xFF, 0xFF}},
		expectedResult: true,
	}, {
		tx: &DomainTransactionOutput{
			0xFFFFFFFF,
			[]byte{0xF0, 0xFF},
		},
		expectedResult: false,
	}, {
		tx: &DomainTransactionOutput{
			0xFFFFFFF0,
			[]byte{0xFF, 0xFF},
		},
		expectedResult: false,
	}}
	return txInput

}
func initTestDomainTransactionOutputForClone() []*DomainTransactionOutput {
	txInput := []*DomainTransactionOutput{
		{
			0xFFFFFFFF,
			[]byte{0xF0, 0xFF},
		}, {
			0xFFFFFFF1,
			[]byte{0xFF, 0xFF},
		}}
	return txInput
}

func initTestDomainTransactionOutputForEqual() []TestDomainTransactionOutputStruct {

	tests := []TestDomainTransactionOutputStruct{
		{
			baseTx:                       initTestBaseDomainTransactionOutput(),
			transactionOutputToCompareTo: initTestDomainTxOutputToCompare(),
		},
	}
	return tests
}

func initTestDomainTransactionInputForEqual() []TestDomainTransactionInputStruct {

	tests := []TestDomainTransactionInputStruct{
		{
			baseTx:                      initTestBaseDomainTransactionInput(),
			transactionInputToCompareTo: initTestDomainTxInputToCompare(),
		},
	}
	return tests
}

func TestDomainTransaction_Equal(t *testing.T) {

	txTests := initTestDomainTransactionForEqual()
	for i, test := range txTests {
		for j, subTest := range test.transactionToCompareTo {
			result1 := test.baseTx.Equal(subTest.tx)
			if result1 != subTest.expectedResult {
				t.Fatalf("Test #%d:%d: Expected %t but got %t", i, j, subTest.expectedResult, result1)
			}
			result2 := subTest.tx.Equal(test.baseTx)
			if result2 != subTest.expectedResult {
				t.Fatalf("Test #%d:%d: Expected %t but got %t", i, j, subTest.expectedResult, result2)
			}
		}
	}
}

func TestDomainTransaction_Clone(t *testing.T) {

	txs := initTestDomainTransactionForClone()
	for i, tx := range txs {
		txClone := tx.Clone()
		if !txClone.Equal(tx) {
			t.Fatalf("Test #%d:[Equal] clone should be equal to the original", i)
		}
		if !reflect.DeepEqual(tx, txClone) {
			t.Fatalf("Test #%d:[DeepEqual] clone should be equal to the original", i)
		}
	}
}

func TestDomainTransactionInput_Equal(t *testing.T) {

	txTests := initTestDomainTransactionInputForEqual()
	for i, test := range txTests {
		for j, subTest := range test.transactionInputToCompareTo {
			result1 := test.baseTx.Equal(subTest.tx)
			if result1 != subTest.expectedResult {
				t.Fatalf("Test #%d:%d: Expected %t but got %t", i, j, subTest.expectedResult, result1)
			}
			result2 := subTest.tx.Equal(test.baseTx)
			if result2 != subTest.expectedResult {
				t.Fatalf("Test #%d:%d: Expected %t but got %t", i, j, subTest.expectedResult, result2)
			}
		}
	}
}

func TestDomainTransactionInput_Clone(t *testing.T) {

	txInputs := initTestDomainTransactionInputForClone()
	for i, txInput := range txInputs {
		txInputClone := txInput.Clone()
		if !txInputClone.Equal(txInput) {
			t.Fatalf("Test #%d:[Equal] clone should be equal to the original", i)
		}
		if !reflect.DeepEqual(txInput, txInputClone) {
			t.Fatalf("Test #%d:[DeepEqual] clone should be equal to the original", i)
		}
	}
}

func TestDomainTransactionOutput_Equal(t *testing.T) {

	txTests := initTestDomainTransactionOutputForEqual()
	for i, test := range txTests {
		for j, subTest := range test.transactionOutputToCompareTo {
			result1 := test.baseTx.Equal(subTest.tx)
			if result1 != subTest.expectedResult {
				t.Fatalf("Test #%d:%d: Expected %t but got %t", i, j, subTest.expectedResult, result1)
			}
			result2 := subTest.tx.Equal(test.baseTx)
			if result2 != subTest.expectedResult {
				t.Fatalf("Test #%d:%d: Expected %t but got %t", i, j, subTest.expectedResult, result2)
			}
		}
	}
}

func TestDomainTransactionOutput_Clone(t *testing.T) {

	txInputs := initTestDomainTransactionOutputForClone()
	for i, txOutput := range txInputs {
		txOutputClone := txOutput.Clone()
		if !txOutputClone.Equal(txOutputClone) {
			t.Fatalf("Test #%d:[Equal] clone should be equal to the original", i)
		}
		if !reflect.DeepEqual(txOutput, txOutputClone) {
			t.Fatalf("Test #%d:[DeepEqual] clone should be equal to the original", i)
		}
	}
}
