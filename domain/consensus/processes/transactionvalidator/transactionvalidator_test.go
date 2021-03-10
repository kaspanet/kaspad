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

	"math/big"

	"testing"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/pkg/errors"
)

type mocPastMedianTimeManager struct {
	pastMedianTimeForTest int64
}

// PastMedianTime returns the past median time for the test.
func (mdf *mocPastMedianTimeManager) PastMedianTime(*externalapi.DomainHash) (int64, error) {
	return mdf.pastMedianTimeForTest, nil
}

func TestValidateTransactionInContextAndPopulateMassAndFee(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {

		factory := consensus.NewFactory()
		pastMedianManager := &mocPastMedianTimeManager{}
		factory.SetTestPastMedianTimeManager(func(int, model.DBReader, model.DAGTraversalManager, model.BlockHeaderStore,
			model.GHOSTDAGDataStore) model.PastMedianTimeManager {
			return pastMedianManager
		})
		tc, tearDown, err := factory.NewTestConsensus(params, false,
			"TestValidateTransactionInContextAndPopulateMassAndFee")
		if err != nil {
			t.Fatalf("Failed create a NewTestConsensus: %s", err)
		}
		defer tearDown(false)

		pastMedianManager.pastMedianTimeForTest = 1
		privateKey, err := secp256k1.GeneratePrivateKey()
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
		addr, err := util.NewAddressPubKeyHashFromPublicKey(publicKeySerialized[:], params.Prefix)
		if err != nil {
			t.Fatalf("Failed to generate p2pkh address: %v", err)
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
		txInputWithMaxSequence := externalapi.DomainTransactionInput{
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
			Inputs:       []*externalapi.DomainTransactionInput{&txInputWithMaxSequence},
			Outputs:      []*externalapi.DomainTransactionOutput{&txOut},
			SubnetworkID: subnetworks.SubnetworkIDRegistry,
			Gas:          0,
			LockTime:     0}
		txWithImmatureCoinbase := externalapi.DomainTransaction{
			Version:      constants.MaxTransactionVersion,
			Inputs:       []*externalapi.DomainTransactionInput{&txInput},
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

		for i, input := range validTx.Inputs {
			signatureScript, err := txscript.SignatureScript(&validTx, i, scriptPublicKey, consensushashing.SigHashAll, privateKey)
			if err != nil {
				t.Fatalf("Failed to create a sigScript: %v", err)
			}
			input.SignatureScript = signatureScript
		}

		povBlockHash := externalapi.NewDomainHashFromByteArray(&[32]byte{0x01})
		genesisHash := params.GenesisHash
		tc.GHOSTDAGDataStore().Stage(model.VirtualBlockHash, model.NewBlockGHOSTDAGData(
			params.BlockCoinbaseMaturity+txInput.UTXOEntry.BlockBlueScore(),
			new(big.Int),
			genesisHash,
			make([]*externalapi.DomainHash, 1000),
			make([]*externalapi.DomainHash, 1),
			nil))
		tc.GHOSTDAGDataStore().Stage(povBlockHash, model.NewBlockGHOSTDAGData(
			10,
			new(big.Int),
			genesisHash,
			make([]*externalapi.DomainHash, 1000),
			make([]*externalapi.DomainHash, 1),
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
				povBlockHash:             model.VirtualBlockHash,
				selectedParentMedianTime: 1,
				isValid:                  true,
				expectedError:            nil,
			},
			{ // The calculated block coinbase maturity is smaller than the minimum expected blockCoinbaseMaturity.
				// The povBlockHash blue score is 10 and the UTXO blue score is 5, hence the The subtraction between
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
				povBlockHash:             model.VirtualBlockHash,
				selectedParentMedianTime: 1,
				isValid:                  false,
				expectedError:            ruleerrors.ErrBadTxOutValue,
			},
			{ // The total SompiIn (sum of inputs amount) is smaller than the total SompiOut (sum of outputs value) and hence invalid.
				name:                     "checkTransactionOutputAmounts",
				tx:                       &txWithBigValue,
				povBlockHash:             model.VirtualBlockHash,
				selectedParentMedianTime: 1,
				isValid:                  false,
				expectedError:            ruleerrors.ErrSpendTooHigh,
			},
			{ // the selectedParentMedianTime is negative and hence invalid.
				name:                     "checkTransactionSequenceLock",
				tx:                       &validTx,
				povBlockHash:             model.VirtualBlockHash,
				selectedParentMedianTime: -1,
				isValid:                  false,
				expectedError:            ruleerrors.ErrUnfinalizedTx,
			},
			{ // The SignatureScript (in the txInput) is empty and hence invalid.
				name:                     "checkTransactionScripts",
				tx:                       &txWithInvalidSignature,
				povBlockHash:             model.VirtualBlockHash,
				selectedParentMedianTime: 1,
				isValid:                  false,
				expectedError:            ruleerrors.ErrScriptValidation,
			},
		}

		for _, test := range tests {
			err := tc.TransactionValidator().ValidateTransactionInContextAndPopulateMassAndFee(test.tx,
				test.povBlockHash, test.selectedParentMedianTime)

			if test.isValid {
				if err != nil {
					t.Fatalf("Unexpected error on TestValidateTransactionInContextAndPopulateMassAndFee"+
						" on test %v: %v", test.name, err)
				}
			} else {
				if err == nil || !errors.Is(err, test.expectedError) {
					t.Fatalf("TestValidateTransactionInContextAndPopulateMassAndFee: test %v:"+
						" Unexpected error: Expected to: %v, but got : %v", test.name, test.expectedError, err)
				}
			}
		}
	})
}
