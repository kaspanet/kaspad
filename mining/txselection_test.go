package mining

import (
	"bou.ke/monkey"
	"fmt"
	"github.com/daglabs/btcd/blockdag"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/txscript"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/util/subnetworkid"
	"github.com/daglabs/btcd/wire"
	"math"
	"strings"
	"testing"
)

type testTxDescDefinition struct {
	fee  uint64
	mass uint64
	gas  uint64

	isExpectedToBeSelected bool
}

func TestSelectTxs(t *testing.T) {
	params := dagconfig.SimNetParams
	params.BlockCoinbaseMaturity = 0

	dag, teardownFunc, err := blockdag.DAGSetup("TestSelectTxs", blockdag.Config{
		DAGParams: &params,
	})
	if err != nil {
		t.Fatalf("Failed to setup DAG instance: %v", err)
	}
	defer teardownFunc()

	txSource := &fakeTxSource{
		txDescs: []*TxDesc{},
	}

	blockTemplateGenerator := NewBlkTmplGenerator(&Policy{BlockMaxMass: 50000},
		&params, txSource, dag, blockdag.NewMedianTime(), txscript.NewSigCache(100000))

	OpTrueAddr, err := OpTrueAddress(params.Prefix)
	if err != nil {
		t.Fatalf("OpTrueAddress: %s", err)
	}
	template, err := blockTemplateGenerator.NewBlockTemplate(OpTrueAddr)
	if err != nil {
		t.Fatalf("NewBlockTemplate: %v", err)
	}
	isOrphan, delay, err := dag.ProcessBlock(util.NewBlock(template.Block), blockdag.BFNoPoWCheck)
	if err != nil {
		t.Fatalf("ProcessBlock: %v", err)
	}
	if delay != 0 {
		t.Fatalf("ProcessBlock: template " +
			"is too far in the future")
	}
	if isOrphan {
		t.Fatalf("ProcessBlock: template got unexpectedly orphan")
	}

	fakeSubnetworkID := subnetworkid.SubnetworkID{250}
	signatureScript, err := txscript.PayToScriptHashSignatureScript(blockdag.OpTrueScript, nil)
	if err != nil {
		t.Fatalf("Error creating signature script: %s", err)
	}
	pkScript, err := txscript.NewScriptBuilder().AddOp(txscript.OpTrue).Script()
	if err != nil {
		t.Fatalf("Failed to create pkScript: %v", err)
	}

	tests := []struct {
		name      string
		massLimit uint64
		gasLimit  uint64
		sourceTxs []testTxDescDefinition
	}{
		{
			name:      "no source txs",
			massLimit: 10,
			gasLimit:  10,
			sourceTxs: []testTxDescDefinition{},
		},
		{
			name:      "zero fee",
			massLimit: 10,
			gasLimit:  10,
			sourceTxs: []testTxDescDefinition{
				{
					mass:                   0,
					gas:                    0,
					fee:                    0,
					isExpectedToBeSelected: false,
				},
			},
		},
		{
			name:      "single transaction",
			massLimit: 100,
			gasLimit:  100,
			sourceTxs: []testTxDescDefinition{
				{
					mass:                   10,
					gas:                    10,
					fee:                    10,
					isExpectedToBeSelected: true,
				},
			},
		},
		{
			name:      "none fit, limited gas and mass",
			massLimit: 2,
			gasLimit:  2,
			sourceTxs: []testTxDescDefinition{
				{
					mass:                   10,
					gas:                    10,
					fee:                    100,
					isExpectedToBeSelected: false,
				},
				{
					mass:                   5,
					gas:                    5,
					fee:                    50,
					isExpectedToBeSelected: false,
				},
			},
		},
		{
			name:      "only one fits, limited gas and mass",
			massLimit: 2,
			gasLimit:  2,
			sourceTxs: []testTxDescDefinition{
				{
					mass:                   0,
					gas:                    0,
					fee:                    1,
					isExpectedToBeSelected: true,
				},
				{
					mass:                   10,
					gas:                    10,
					fee:                    100,
					isExpectedToBeSelected: false,
				},
				{
					mass:                   5,
					gas:                    5,
					fee:                    50,
					isExpectedToBeSelected: false,
				},
			},
		},
		{
			name:      "all fit, limited gas",
			massLimit: wire.MaxMassPerBlock,
			gasLimit:  10,
			sourceTxs: []testTxDescDefinition{
				{
					mass:                   100,
					gas:                    1,
					fee:                    100,
					isExpectedToBeSelected: true,
				},
				{
					mass:                   0,
					gas:                    1,
					fee:                    1,
					isExpectedToBeSelected: true,
				},
				{
					mass:                   2,
					gas:                    1,
					fee:                    100,
					isExpectedToBeSelected: true,
				},
				{
					mass:                   3,
					gas:                    1,
					fee:                    100,
					isExpectedToBeSelected: true,
				},
				{
					mass:                   4,
					gas:                    1,
					fee:                    100,
					isExpectedToBeSelected: true,
				},
			},
		},
		{
			name:      "all fit, limited mass",
			massLimit: 10,
			gasLimit:  math.MaxUint64,
			sourceTxs: []testTxDescDefinition{
				{
					mass:                   1,
					gas:                    100,
					fee:                    100,
					isExpectedToBeSelected: true,
				},
				{
					mass:                   1,
					gas:                    0,
					fee:                    1,
					isExpectedToBeSelected: true,
				},
				{
					mass:                   1,
					gas:                    2,
					fee:                    100,
					isExpectedToBeSelected: true,
				},
				{
					mass:                   1,
					gas:                    3,
					fee:                    100,
					isExpectedToBeSelected: true,
				},
				{
					mass:                   1,
					gas:                    4,
					fee:                    100,
					isExpectedToBeSelected: true,
				},
			},
		},
	}

	for _, test := range tests {
		func() {
			// Force the mass limit to always be test.massLimit
			blockTemplateGenerator.policy.BlockMaxMass = test.massLimit

			// Force the mass to be as defined in the definition.
			// We use the first payload byte to resolve which definition to use.
			massPatch := monkey.Patch(blockdag.CalcTxMass, func(tx *util.Tx, _ blockdag.UTXOSet) (uint64, error) {
				if tx.IsCoinBase() {
					return 1, nil
				}
				index := tx.MsgTx().Payload[0]
				definition := test.sourceTxs[index]
				return definition.mass, nil
			})
			defer massPatch.Unpatch()

			// Force the gas limit to always be test.gasLimit
			gasLimitPatch := monkey.Patch((*blockdag.SubnetworkStore).GasLimit, func(_ *blockdag.SubnetworkStore, subnetworkID *subnetworkid.SubnetworkID) (uint64, error) {
				return test.gasLimit, nil
			})
			defer gasLimitPatch.Unpatch()

			// Force the fee to be as defined in the definition.
			// We use the first payload byte to resolve which definition to use.
			feePatch := monkey.Patch(blockdag.CheckTransactionInputsAndCalulateFee, func(tx *util.Tx, _ uint64, _ blockdag.UTXOSet, _ *dagconfig.Params, _ bool) (txFeeInSatoshi uint64, err error) {
				if tx.IsCoinBase() {
					return 1, nil
				}
				index := tx.MsgTx().Payload[0]
				definition := test.sourceTxs[index]
				return definition.fee, nil
			})
			defer feePatch.Unpatch()

			// Load the txSource with transactions as defined in test.sourceTxs.
			// Note that we're saving the definition index in the msgTx payload
			// so that we may use it in massPatch and feePatch.
			// We're also saving for later the util.txs that we expect to be selected
			txSource.txDescs = make([]*TxDesc, len(test.sourceTxs))
			expectedSelectedTxs := make([]*util.Tx, 0, len(test.sourceTxs))
			for i, definition := range test.sourceTxs {
				txIn := &wire.TxIn{
					PreviousOutpoint: wire.Outpoint{
						TxID:  *template.Block.Transactions[util.CoinbaseTransactionIndex].TxID(),
						Index: 0,
					},
					Sequence:        wire.MaxTxInSequenceNum,
					SignatureScript: signatureScript,
				}
				txOut := &wire.TxOut{
					PkScript: pkScript,
					Value:    1,
				}
				msgTx := wire.NewSubnetworkMsgTx(
					wire.TxVersion, []*wire.TxIn{txIn}, []*wire.TxOut{txOut},
					&fakeSubnetworkID, definition.gas, []byte{byte(i)})
				tx := util.NewTx(msgTx)
				txDesc := TxDesc{
					Fee: definition.fee,
					Tx:  tx,
				}
				txSource.txDescs[i] = &txDesc

				if definition.isExpectedToBeSelected {
					expectedSelectedTxs = append(expectedSelectedTxs, tx)
				}
			}

			result, err := blockTemplateGenerator.selectTxs(OpTrueAddr)
			if err != nil {
				t.Errorf("selectTxs unexpectedly failed in test '%s': %s",
					test.name, err)
				return
			}

			// Ignore the first transactions because it's the coinbase.
			selectedTxs := result.selectedTxs[1:]

			// Check whether expectedSelectedTxs and selectedTxs contain
			// the same txs.
			areLengthsEqual := len(expectedSelectedTxs) == len(selectedTxs)
			areEqual := areLengthsEqual
			if areLengthsEqual {
				for _, expectedTx := range expectedSelectedTxs {
					wasFound := false
					for _, selectedTx := range selectedTxs {
						if expectedTx == selectedTx {
							wasFound = true
							break
						}
					}
					if !wasFound {
						areEqual = false
						break
					}
				}
			}

			if !areEqual {
				t.Errorf("unexpected selected txs in test '%s'. Want: [%s], got: [%s] ",
					test.name, formatTxs(expectedSelectedTxs), formatTxs(selectedTxs))
			}
		}()
	}
}

func formatTxs(txs []*util.Tx) string {
	strs := make([]string, len(txs))
	for i, tx := range txs {
		mass, _ := blockdag.CalcTxMass(tx, nil)
		fee, _ := blockdag.CheckTransactionInputsAndCalulateFee(tx, 0, nil, nil, false)
		strs[i] = fmt.Sprintf("[mass: %d, gas: %d, fee: %d]",
			mass, tx.MsgTx().Gas, fee)
	}
	return strings.Join(strs, ", ")
}
