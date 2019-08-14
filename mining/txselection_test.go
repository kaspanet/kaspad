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
	"testing"
)

type testTxDescDefinition struct {
	fee  uint64
	mass uint64
	gas  uint64

	expectedMinSelectedTimes uint64
	expectedMaxSelectedTimes uint64

	tx *util.Tx
}

func (dd testTxDescDefinition) String() string {
	return fmt.Sprintf("[fee: %d, gas: %d, mass: %d]", dd.fee, dd.gas, dd.mass)
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
		name          string
		runTimes      int
		massLimit     uint64
		gasLimit      uint64
		txDefinitions []*testTxDescDefinition
	}{
		{
			name:          "no source txs",
			runTimes:      1,
			massLimit:     10,
			gasLimit:      10,
			txDefinitions: []*testTxDescDefinition{},
		},
		{
			name:      "zero fee",
			runTimes:  1,
			massLimit: 10,
			gasLimit:  10,
			txDefinitions: []*testTxDescDefinition{
				{
					mass: 0,
					gas:  0,
					fee:  0,

					// Expected probability: 0
					expectedMinSelectedTimes: 0,
					expectedMaxSelectedTimes: 0,
				},
			},
		},
		{
			name:      "single transaction",
			runTimes:  1,
			massLimit: 100,
			gasLimit:  100,
			txDefinitions: []*testTxDescDefinition{
				{
					mass: 10,
					gas:  10,
					fee:  10,

					// Expected probability: 1
					expectedMinSelectedTimes: 1,
					expectedMaxSelectedTimes: 1,
				},
			},
		},
		{
			name:      "none fit, limited gas and mass",
			runTimes:  1,
			massLimit: 2,
			gasLimit:  2,
			txDefinitions: []*testTxDescDefinition{
				{
					mass: 10,
					gas:  10,
					fee:  100,

					// Expected probability: 0
					expectedMinSelectedTimes: 0,
					expectedMaxSelectedTimes: 0,
				},
				{
					mass: 5,
					gas:  5,
					fee:  50,

					// Expected probability: 0
					expectedMinSelectedTimes: 0,
					expectedMaxSelectedTimes: 0,
				},
			},
		},
		{
			name:      "only one fits, limited gas and mass",
			runTimes:  1,
			massLimit: 2,
			gasLimit:  2,
			txDefinitions: []*testTxDescDefinition{
				{
					mass: 1,
					gas:  1,
					fee:  50,

					// Expected probability: 1
					expectedMinSelectedTimes: 1,
					expectedMaxSelectedTimes: 1,
				},
				{
					mass: 10,
					gas:  10,
					fee:  100,

					// Expected probability: 0
					expectedMinSelectedTimes: 0,
					expectedMaxSelectedTimes: 0,
				},
				{
					mass: 5,
					gas:  5,
					fee:  50,

					// Expected probability: 0
					expectedMinSelectedTimes: 0,
					expectedMaxSelectedTimes: 0,
				},
			},
		},
		{
			name:      "all fit, limited gas",
			runTimes:  1,
			massLimit: wire.MaxMassPerBlock,
			gasLimit:  10,
			txDefinitions: []*testTxDescDefinition{
				{
					mass: 100,
					gas:  1,
					fee:  100,

					// Expected probability: 1
					expectedMinSelectedTimes: 1,
					expectedMaxSelectedTimes: 1,
				},
				{
					mass: 0,
					gas:  1,
					fee:  1,

					// Expected probability: 1
					expectedMinSelectedTimes: 1,
					expectedMaxSelectedTimes: 1,
				},
				{
					mass: 2,
					gas:  1,
					fee:  100,

					// Expected probability: 1
					expectedMinSelectedTimes: 1,
					expectedMaxSelectedTimes: 1,
				},
				{
					mass: 3,
					gas:  1,
					fee:  100,

					// Expected probability: 1
					expectedMinSelectedTimes: 1,
					expectedMaxSelectedTimes: 1,
				},
				{
					mass: 4,
					gas:  1,
					fee:  100,

					// Expected probability: 1
					expectedMinSelectedTimes: 1,
					expectedMaxSelectedTimes: 1,
				},
			},
		},
		{
			name:      "all fit, limited mass",
			runTimes:  1,
			massLimit: 10,
			gasLimit:  math.MaxUint64,
			txDefinitions: []*testTxDescDefinition{
				{
					mass: 1,
					gas:  100,
					fee:  100,

					// Expected probability: 1
					expectedMinSelectedTimes: 1,
					expectedMaxSelectedTimes: 1,
				},
				{
					mass: 1,
					gas:  0,
					fee:  1,

					// Expected probability: 1
					expectedMinSelectedTimes: 1,
					expectedMaxSelectedTimes: 1,
				},
				{
					mass: 1,
					gas:  2,
					fee:  100,

					// Expected probability: 1
					expectedMinSelectedTimes: 1,
					expectedMaxSelectedTimes: 1,
				},
				{
					mass: 1,
					gas:  3,
					fee:  100,

					// Expected probability: 1
					expectedMinSelectedTimes: 1,
					expectedMaxSelectedTimes: 1,
				},
				{
					mass: 1,
					gas:  4,
					fee:  100,

					// Expected probability: 1
					expectedMinSelectedTimes: 1,
					expectedMaxSelectedTimes: 1,
				},
			},
		},
		{
			name:      "equal selection probability",
			runTimes:  1000,
			massLimit: 100,
			gasLimit:  100,
			txDefinitions: []*testTxDescDefinition{
				{
					mass: 75,
					gas:  75,
					fee:  100,

					// Expected probability: 0.25
					expectedMinSelectedTimes: 200,
					expectedMaxSelectedTimes: 300,
				},
				{
					mass: 75,
					gas:  75,
					fee:  100,

					// Expected probability: 0.25
					expectedMinSelectedTimes: 200,
					expectedMaxSelectedTimes: 300,
				},
				{
					mass: 75,
					gas:  75,
					fee:  100,

					// Expected probability: 0.25
					expectedMinSelectedTimes: 200,
					expectedMaxSelectedTimes: 300,
				},
				{
					mass: 75,
					gas:  75,
					fee:  100,

					// Expected probability: 0.25
					expectedMinSelectedTimes: 200,
					expectedMaxSelectedTimes: 300,
				},
			},
		},
		{
			name:      "unequal selection probability",
			runTimes:  1000,
			massLimit: 100,
			gasLimit:  100,
			txDefinitions: []*testTxDescDefinition{
				{
					mass: 50,
					gas:  50,
					fee:  100,

					// Expected probability: 0.33
					expectedMinSelectedTimes: 280,
					expectedMaxSelectedTimes: 380,
				},
				{
					mass: 100,
					gas:  0,
					fee:  100,

					// Expected probability: 0.50
					expectedMinSelectedTimes: 450,
					expectedMaxSelectedTimes: 550,
				},
				{
					mass: 0,
					gas:  100,
					fee:  100,

					// Expected probability: 0.50
					expectedMinSelectedTimes: 450,
					expectedMaxSelectedTimes: 550,
				},
			},
		},
		{
			name:      "distributed selection probability",
			runTimes:  100,
			massLimit: 32,
			gasLimit:  32,
			txDefinitions: []*testTxDescDefinition{
				{
					mass: 1,
					gas:  1,
					fee:  100,

					// Expected probability: 1
					expectedMinSelectedTimes: 95,
					expectedMaxSelectedTimes: 100,
				},
				{
					mass: 2,
					gas:  2,
					fee:  100,

					// Expected probability: 1
					expectedMinSelectedTimes: 95,
					expectedMaxSelectedTimes: 100,
				},
				{
					mass: 4,
					gas:  4,
					fee:  100,

					// Expected probability: 1
					expectedMinSelectedTimes: 95,
					expectedMaxSelectedTimes: 100,
				},
				{
					mass: 8,
					gas:  8,
					fee:  100,

					// Expected probability: 0.98
					expectedMinSelectedTimes: 90,
					expectedMaxSelectedTimes: 100,
				},
				{
					mass: 16,
					gas:  16,
					fee:  100,

					// Expected probability: 0.90
					expectedMinSelectedTimes: 80,
					expectedMaxSelectedTimes: 100,
				},
				{
					mass: 32,
					gas:  32,
					fee:  100,

					// Expected probability: 0
					expectedMinSelectedTimes: 0,
					expectedMaxSelectedTimes: 5,
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
					return 0, nil
				}
				index := tx.MsgTx().Payload[0]
				definition := test.txDefinitions[index]
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
					return 0, nil
				}
				index := tx.MsgTx().Payload[0]
				definition := test.txDefinitions[index]
				return definition.fee, nil
			})
			defer feePatch.Unpatch()

			// Load the txSource with transactions as defined in test.txDefinitions.
			// Note that we're saving the definition index in the msgTx payload
			// so that we may use it in massPatch and feePatch.
			// We also initialize a map that keeps track of how many times a tx
			// has been selected.
			txSource.txDescs = make([]*TxDesc, len(test.txDefinitions))
			selectedTxCountMap := make(map[*util.Tx]uint64, len(test.txDefinitions))
			for i, definition := range test.txDefinitions {
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

				definition.tx = tx
				selectedTxCountMap[tx] = 0
			}

			// Run selectTxs test.runTimes times
			for i := 0; i < test.runTimes; i++ {
				result, err := blockTemplateGenerator.selectTxs(OpTrueAddr)
				if err != nil {
					t.Errorf("selectTxs unexpectedly failed in test '%s': %s",
						test.name, err)
					return
				}

				// Increment the counts of all the selected transactions.
				// Ignore the first transactions because it's the coinbase.
				for _, selectedTx := range result.selectedTxs[1:] {
					selectedTxCountMap[selectedTx]++
				}
			}

			// Make sure that each transaction has not been selected either
			// too little or too much.
			for i, definition := range test.txDefinitions {
				tx := definition.tx
				count := selectedTxCountMap[tx]
				min := definition.expectedMinSelectedTimes
				max := definition.expectedMaxSelectedTimes
				if count < min || count > max {
					t.Errorf("unexpected selected tx count "+
						"in test '%s' for tx %d:%s. Want: %d <= count <= %d, got: %d",
						test.name, i, definition, min, max, count)
				}
			}
		}()
	}
}
