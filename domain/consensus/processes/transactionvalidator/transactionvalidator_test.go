package transactionvalidator_test

import (
	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/kaspanet/kaspad/util"

	"testing"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"github.com/pkg/errors"
)

type mocPastMedianTimeManager struct {
	pastMedianTimeForTest int64
}

func (mdf *mocPastMedianTimeManager) InvalidateVirtualPastMedianTimeCache() {
	// do nothing
}

// PastMedianTime returns the past median time for the test.
func (mdf *mocPastMedianTimeManager) PastMedianTime(*model.StagingArea, *externalapi.DomainHash) (int64, error) {
	return mdf.pastMedianTimeForTest, nil
}

func TestValidateTransactionInContextAndPopulateMassAndFee(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {

		factory := consensus.NewFactory()
		pastMedianManager := &mocPastMedianTimeManager{}
		factory.SetTestPastMedianTimeManager(func(int, model.DBReader, model.DAGTraversalManager, model.BlockHeaderStore,
			model.GHOSTDAGDataStore, *externalapi.DomainHash) model.PastMedianTimeManager {
			return pastMedianManager
		})
		tc, tearDown, err := factory.NewTestConsensus(consensusConfig,
			"TestValidateTransactionInContextAndPopulateMassAndFee")
		if err != nil {
			t.Fatalf("Failed create a NewTestConsensus: %s", err)
		}
		defer tearDown(false)

		pastMedianManager.pastMedianTimeForTest = 1
		privateKey, err := secp256k1.GenerateSchnorrKeyPair()
		if err != nil {
			t.Fatalf("Failed to generate a private key: %v", err)
		}
		publicKey, err := privateKey.SchnorrPublicKey()
		if err != nil {
			t.Fatalf("Failed to generate a public key: %v", err)
		}
		publicKeySerialized, err := publicKey.Serialize()
		if err != nil {
			t.Fatalf("Failed to serialize public key: %v", err)
		}
		addr, err := util.NewAddressPublicKey(publicKeySerialized[:], consensusConfig.Prefix)
		if err != nil {
			t.Fatalf("Failed to generate p2pk address: %v", err)
		}
		scriptPublicKey, err := txscript.PayToAddrScript(addr)
		if err != nil {
			t.Fatalf("PayToAddrScript: unexpected error: %v", err)
		}
		prevOutTxID := &externalapi.DomainTransactionID{}
		prevOutPoint := externalapi.DomainOutpoint{TransactionID: *prevOutTxID, Index: 1}

		txInput := externalapi.DomainTransactionInput{
			PreviousOutpoint: prevOutPoint,
			SignatureScript:  []byte{},
			Sequence:         constants.MaxTxInSequenceNum,
			UTXOEntry: utxo.NewUTXOEntry(
				100_000_000, // 1 KAS
				scriptPublicKey,
				true,
				uint64(5)),
		}
		immatureInput := externalapi.DomainTransactionInput{
			PreviousOutpoint: prevOutPoint,
			SignatureScript:  []byte{},
			Sequence:         constants.MaxTxInSequenceNum,
			UTXOEntry: utxo.NewUTXOEntry(
				100_000_000, // 1 KAS
				scriptPublicKey,
				true,
				uint64(6)),
		}
		txInputWithSequenceLockTimeIsSeconds := externalapi.DomainTransactionInput{
			PreviousOutpoint: prevOutPoint,
			SignatureScript:  []byte{},
			Sequence:         constants.SequenceLockTimeIsSeconds,
			UTXOEntry: utxo.NewUTXOEntry(
				100000000, // 1 KAS
				scriptPublicKey,
				true,
				uint64(5)),
		}
		txInputWithLargeAmount := externalapi.DomainTransactionInput{
			PreviousOutpoint: prevOutPoint,
			SignatureScript:  []byte{},
			Sequence:         constants.MaxTxInSequenceNum,
			UTXOEntry: utxo.NewUTXOEntry(
				constants.MaxSompi,
				scriptPublicKey,
				true,
				uint64(5)),
		}

		txOut := externalapi.DomainTransactionOutput{
			Value:           100000000, // 1 KAS
			ScriptPublicKey: scriptPublicKey,
		}
		txOutBigValue := externalapi.DomainTransactionOutput{
			Value:           200_000_000, // 2 KAS
			ScriptPublicKey: scriptPublicKey,
		}

		validTx := externalapi.DomainTransaction{
			Version:      constants.MaxTransactionVersion,
			Inputs:       []*externalapi.DomainTransactionInput{&txInputWithSequenceLockTimeIsSeconds},
			Outputs:      []*externalapi.DomainTransactionOutput{&txOut},
			SubnetworkID: subnetworks.SubnetworkIDRegistry,
			Gas:          0,
			LockTime:     0}

		for i, input := range validTx.Inputs {
			signatureScript, err := txscript.SignatureScript(&validTx, i, consensushashing.SigHashAll, privateKey,
				&consensushashing.SighashReusedValues{})
			if err != nil {
				t.Fatalf("Failed to create a sigScript: %v", err)
			}
			input.SignatureScript = signatureScript
		}

		txWithImmatureCoinbase := externalapi.DomainTransaction{
			Version:      constants.MaxTransactionVersion,
			Inputs:       []*externalapi.DomainTransactionInput{&immatureInput},
			Outputs:      []*externalapi.DomainTransactionOutput{&txOut},
			SubnetworkID: subnetworks.SubnetworkIDRegistry,
			Gas:          0,
			LockTime:     0}
		txWithLargeAmount := externalapi.DomainTransaction{
			Version:      constants.MaxTransactionVersion,
			Inputs:       []*externalapi.DomainTransactionInput{&txInput, &txInputWithLargeAmount},
			Outputs:      []*externalapi.DomainTransactionOutput{&txOut},
			SubnetworkID: subnetworks.SubnetworkIDRegistry,
			Gas:          0,
			LockTime:     0}
		txWithBigValue := externalapi.DomainTransaction{
			Version:      constants.MaxTransactionVersion,
			Inputs:       []*externalapi.DomainTransactionInput{&txInput},
			Outputs:      []*externalapi.DomainTransactionOutput{&txOutBigValue},
			SubnetworkID: subnetworks.SubnetworkIDRegistry,
			Gas:          0,
			LockTime:     0}
		txWithInvalidSignature := externalapi.DomainTransaction{
			Version:      constants.MaxTransactionVersion,
			Inputs:       []*externalapi.DomainTransactionInput{&txInput},
			Outputs:      []*externalapi.DomainTransactionOutput{&txOut},
			SubnetworkID: subnetworks.SubnetworkIDRegistry,
			Gas:          0,
			LockTime:     0}

		stagingArea := model.NewStagingArea()

		povBlockHash := externalapi.NewDomainHashFromByteArray(&[32]byte{0x01})
		tc.DAABlocksStore().StageDAAScore(stagingArea, povBlockHash, consensusConfig.BlockCoinbaseMaturity+txInput.UTXOEntry.BlockDAAScore())

		// Just use some stub ghostdag data
		tc.GHOSTDAGDataStore().Stage(stagingArea, povBlockHash, externalapi.NewBlockGHOSTDAGData(
			0,
			nil,
			consensusConfig.GenesisHash,
			nil,
			nil,
			nil))

		tests := []struct {
			name                     string
			tx                       *externalapi.DomainTransaction
			povBlockHash             *externalapi.DomainHash
			selectedParentMedianTime int64
			isValid                  bool
			expectedError            error
		}{
			{
				name:                     "Valid transaction",
				tx:                       &validTx,
				povBlockHash:             povBlockHash,
				selectedParentMedianTime: 1,
				isValid:                  true,
				expectedError:            nil,
			},
			{ // The calculated block coinbase maturity is smaller than the minimum expected blockCoinbaseMaturity.
				// The povBlockHash DAA score is 10 and the UTXO DAA score is 5, hence the The subtraction between
				// them will yield a smaller result than the required CoinbaseMaturity (currently set to 100).
				name:                     "checkTransactionCoinbaseMaturity",
				tx:                       &txWithImmatureCoinbase,
				povBlockHash:             povBlockHash,
				selectedParentMedianTime: 1,
				isValid:                  false,
				expectedError:            ruleerrors.ErrImmatureSpend,
			},
			{ // The total inputs amount is bigger than the allowed maximum (constants.MaxSompi)
				name:                     "checkTransactionInputAmounts",
				tx:                       &txWithLargeAmount,
				povBlockHash:             povBlockHash,
				selectedParentMedianTime: 1,
				isValid:                  false,
				expectedError:            ruleerrors.ErrBadTxOutValue,
			},
			{ // The total SompiIn (sum of inputs amount) is smaller than the total SompiOut (sum of outputs value) and hence invalid.
				name:                     "checkTransactionOutputAmounts",
				tx:                       &txWithBigValue,
				povBlockHash:             povBlockHash,
				selectedParentMedianTime: 1,
				isValid:                  false,
				expectedError:            ruleerrors.ErrSpendTooHigh,
			},
			{ // the selectedParentMedianTime is negative and hence invalid.
				name:                     "checkTransactionSequenceLock",
				tx:                       &validTx,
				povBlockHash:             povBlockHash,
				selectedParentMedianTime: -1,
				isValid:                  false,
				expectedError:            ruleerrors.ErrUnfinalizedTx,
			},
			{ // The SignatureScript (in the immatureInput) is empty and hence invalid.
				name:                     "checkTransactionScripts",
				tx:                       &txWithInvalidSignature,
				povBlockHash:             povBlockHash,
				selectedParentMedianTime: 1,
				isValid:                  false,
				expectedError:            ruleerrors.ErrScriptValidation,
			},
		}

		for _, test := range tests {
			err := tc.TransactionValidator().ValidateTransactionInContextAndPopulateMassAndFee(stagingArea, test.tx, test.povBlockHash, test.selectedParentMedianTime)

			if test.isValid {
				if err != nil {
					t.Fatalf("Unexpected error on TestValidateTransactionInContextAndPopulateMassAndFee"+
						" on test '%v': %+v", test.name, err)
				}
			} else {
				if err == nil || !errors.Is(err, test.expectedError) {
					t.Fatalf("TestValidateTransactionInContextAndPopulateMassAndFee: test %v:"+
						" Unexpected error: Expected to: %v, but got : %+v", test.name, test.expectedError, err)
				}
			}
		}
	})
}

func TestSigningTwoInputs(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		consensusConfig.BlockCoinbaseMaturity = 0
		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(consensusConfig, "TestSigningTwoInputs")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		privateKey, err := secp256k1.GenerateSchnorrKeyPair()
		if err != nil {
			t.Fatalf("Failed to generate a private key: %v", err)
		}
		publicKey, err := privateKey.SchnorrPublicKey()
		if err != nil {
			t.Fatalf("Failed to generate a public key: %v", err)
		}
		publicKeySerialized, err := publicKey.Serialize()
		if err != nil {
			t.Fatalf("Failed to serialize public key: %v", err)
		}
		addr, err := util.NewAddressPublicKey(publicKeySerialized[:], consensusConfig.Prefix)
		if err != nil {
			t.Fatalf("Failed to generate p2pk address: %v", err)
		}

		scriptPublicKey, err := txscript.PayToAddrScript(addr)
		if err != nil {
			t.Fatalf("PayToAddrScript: unexpected error: %v", err)
		}

		coinbaseData := &externalapi.DomainCoinbaseData{
			ScriptPublicKey: scriptPublicKey,
		}

		block1Hash, _, err := tc.AddBlock([]*externalapi.DomainHash{consensusConfig.GenesisHash}, coinbaseData, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		block2Hash, _, err := tc.AddBlock([]*externalapi.DomainHash{block1Hash}, coinbaseData, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		block3Hash, _, err := tc.AddBlock([]*externalapi.DomainHash{block2Hash}, coinbaseData, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		block2, err := tc.GetBlock(block2Hash)
		if err != nil {
			t.Fatalf("Error getting block2: %+v", err)
		}

		block3, err := tc.GetBlock(block3Hash)
		if err != nil {
			t.Fatalf("Error getting block3: %+v", err)
		}

		block2Tx := block2.Transactions[0]
		block2TxOut := block2Tx.Outputs[0]

		block3Tx := block3.Transactions[0]
		block3TxOut := block3Tx.Outputs[0]

		tx := &externalapi.DomainTransaction{
			Version: constants.MaxTransactionVersion,
			Inputs: []*externalapi.DomainTransactionInput{
				{
					PreviousOutpoint: externalapi.DomainOutpoint{
						TransactionID: *consensushashing.TransactionID(block2.Transactions[0]),
						Index:         0,
					},
					Sequence:  constants.MaxTxInSequenceNum,
					UTXOEntry: utxo.NewUTXOEntry(block2TxOut.Value, block2TxOut.ScriptPublicKey, true, 0),
				},
				{
					PreviousOutpoint: externalapi.DomainOutpoint{
						TransactionID: *consensushashing.TransactionID(block3.Transactions[0]),
						Index:         0,
					},
					Sequence:  constants.MaxTxInSequenceNum,
					UTXOEntry: utxo.NewUTXOEntry(block3TxOut.Value, block3TxOut.ScriptPublicKey, true, 0),
				},
			},
			Outputs: []*externalapi.DomainTransactionOutput{{
				Value: 1,
				ScriptPublicKey: &externalapi.ScriptPublicKey{
					Script:  nil,
					Version: 0,
				},
			}},
			SubnetworkID: subnetworks.SubnetworkIDNative,
			Gas:          0,
			LockTime:     0,
		}

		sighashReusedValues := &consensushashing.SighashReusedValues{}
		for i, input := range tx.Inputs {
			signatureScript, err := txscript.SignatureScript(tx, i, consensushashing.SigHashAll, privateKey,
				sighashReusedValues)
			if err != nil {
				t.Fatalf("Failed to create a sigScript: %v", err)
			}
			input.SignatureScript = signatureScript
		}

		_, insertionResult, err := tc.AddBlock([]*externalapi.DomainHash{block3Hash}, nil, []*externalapi.DomainTransaction{tx})
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		txOutpoint := &externalapi.DomainOutpoint{
			TransactionID: *consensushashing.TransactionID(tx),
			Index:         0,
		}
		if !insertionResult.VirtualUTXODiff.ToAdd().Contains(txOutpoint) {
			t.Fatalf("tx was not accepted by the DAG")
		}
	})
}

func TestSigningTwoInputsECDSA(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		consensusConfig.BlockCoinbaseMaturity = 0
		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(consensusConfig, "TestSigningTwoInputsECDSA")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		privateKey, err := secp256k1.GenerateECDSAPrivateKey()
		if err != nil {
			t.Fatalf("Failed to generate a private key: %v", err)
		}
		publicKey, err := privateKey.ECDSAPublicKey()
		if err != nil {
			t.Fatalf("Failed to generate a public key: %v", err)
		}
		publicKeySerialized, err := publicKey.Serialize()
		if err != nil {
			t.Fatalf("Failed to serialize public key: %v", err)
		}
		addr, err := util.NewAddressPublicKeyECDSA(publicKeySerialized[:], consensusConfig.Prefix)
		if err != nil {
			t.Fatalf("Failed to generate p2pk address: %v", err)
		}

		scriptPublicKey, err := txscript.PayToAddrScript(addr)
		if err != nil {
			t.Fatalf("PayToAddrScript: unexpected error: %v", err)
		}

		coinbaseData := &externalapi.DomainCoinbaseData{
			ScriptPublicKey: scriptPublicKey,
		}

		block1Hash, _, err := tc.AddBlock([]*externalapi.DomainHash{consensusConfig.GenesisHash}, coinbaseData, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		block2Hash, _, err := tc.AddBlock([]*externalapi.DomainHash{block1Hash}, coinbaseData, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		block3Hash, _, err := tc.AddBlock([]*externalapi.DomainHash{block2Hash}, coinbaseData, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		block2, err := tc.GetBlock(block2Hash)
		if err != nil {
			t.Fatalf("Error getting block2: %+v", err)
		}

		block3, err := tc.GetBlock(block3Hash)
		if err != nil {
			t.Fatalf("Error getting block3: %+v", err)
		}

		block2Tx := block2.Transactions[0]
		block2TxOut := block2Tx.Outputs[0]

		block3Tx := block3.Transactions[0]
		block3TxOut := block3Tx.Outputs[0]

		tx := &externalapi.DomainTransaction{
			Version: constants.MaxTransactionVersion,
			Inputs: []*externalapi.DomainTransactionInput{
				{
					PreviousOutpoint: externalapi.DomainOutpoint{
						TransactionID: *consensushashing.TransactionID(block2.Transactions[0]),
						Index:         0,
					},
					Sequence:  constants.MaxTxInSequenceNum,
					UTXOEntry: utxo.NewUTXOEntry(block2TxOut.Value, block2TxOut.ScriptPublicKey, true, 0),
				},
				{
					PreviousOutpoint: externalapi.DomainOutpoint{
						TransactionID: *consensushashing.TransactionID(block3.Transactions[0]),
						Index:         0,
					},
					Sequence:  constants.MaxTxInSequenceNum,
					UTXOEntry: utxo.NewUTXOEntry(block3TxOut.Value, block3TxOut.ScriptPublicKey, true, 0),
				},
			},
			Outputs: []*externalapi.DomainTransactionOutput{{
				Value: 1,
				ScriptPublicKey: &externalapi.ScriptPublicKey{
					Script:  nil,
					Version: 0,
				},
			}},
			SubnetworkID: subnetworks.SubnetworkIDNative,
			Gas:          0,
			LockTime:     0,
		}

		sighashReusedValues := &consensushashing.SighashReusedValues{}
		for i, input := range tx.Inputs {
			signatureScript, err := txscript.SignatureScriptECDSA(tx, i, consensushashing.SigHashAll, privateKey,
				sighashReusedValues)
			if err != nil {
				t.Fatalf("Failed to create a sigScript: %v", err)
			}
			input.SignatureScript = signatureScript
		}

		_, insertionResult, err := tc.AddBlock([]*externalapi.DomainHash{block3Hash}, nil, []*externalapi.DomainTransaction{tx})
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		txOutpoint := &externalapi.DomainOutpoint{
			TransactionID: *consensushashing.TransactionID(tx),
			Index:         0,
		}
		if !insertionResult.VirtualUTXODiff.ToAdd().Contains(txOutpoint) {
			t.Fatalf("tx was not accepted by the DAG")
		}
	})
}
