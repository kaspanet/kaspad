package consensusstatemanager_test

import (
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus/utils/multiset"

	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"

	"github.com/kaspanet/kaspad/domain/consensus/utils/consensusserialization"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"

	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/dagconfig"
)

// opTrueScript is script returning TRUE
var opTrueScript = []byte{txscript.OpTrue}

func TestUTXOCommitment(t *testing.T) {
	params := dagconfig.SimnetParams
	params.BlockCoinbaseMaturity = 0
	factory := consensus.NewFactory()

	consensus, teardown, err := factory.NewTestConsensus(&params, "TestUTXOCommitment")
	if err != nil {
		t.Fatalf("Error setting up consensus: %+v", err)
	}
	defer teardown()

	createTransaction := func(txToSpend *externalapi.DomainTransaction) *externalapi.DomainTransaction {
		scriptPubKey, err := txscript.PayToScriptHashScript(opTrueScript)
		if err != nil {
			t.Fatalf("TestUTXOCommitment: failed to build script pub key: %s", err)
		}
		signatureScript, err := txscript.PayToScriptHashSignatureScript(opTrueScript, nil)
		if err != nil {
			t.Fatalf("TestUTXOCommitment: failed to build signature script: %s", err)
		}
		input := &externalapi.DomainTransactionInput{
			PreviousOutpoint: externalapi.DomainOutpoint{
				TransactionID: *consensusserialization.TransactionID(txToSpend),
				Index:         0,
			},
			SignatureScript: signatureScript,
			Sequence:        appmessage.MaxTxInSequenceNum,
		}
		output := &externalapi.DomainTransactionOutput{
			ScriptPublicKey: scriptPubKey,
			Value:           uint64(1),
		}
		return &externalapi.DomainTransaction{
			Version: constants.TransactionVersion,
			Inputs:  []*externalapi.DomainTransactionInput{input},
			Outputs: []*externalapi.DomainTransactionOutput{output},
		}
	}
	coinbaseData := &externalapi.DomainCoinbaseData{ScriptPublicKey: opTrueScript, ExtraData: []byte{}}

	// Build the following DAG:
	// G <- A <- B <- D
	//        <- C <-
	genesisHash := params.GenesisHash

	// Block A:
	blockAHash, err := consensus.AddBlock(
		[]*externalapi.DomainHash{genesisHash}, coinbaseData, []*externalapi.DomainTransaction{})
	if err != nil {
		t.Fatalf("Error creating block A: %+v", err)
	}
	blockA, err := consensus.GetBlock(blockAHash)
	if err != nil {
		t.Fatalf("Error getting block A: %+v", err)
	}
	// Block B:
	blockBHash, err := consensus.AddBlock(
		[]*externalapi.DomainHash{blockAHash}, coinbaseData, []*externalapi.DomainTransaction{})
	if err != nil {
		t.Fatalf("Error creating block B: %+v", err)
	}
	// Block C:
	blockCTransaction := createTransaction(blockA.Transactions[0])
	blockCHash, err := consensus.AddBlock(
		[]*externalapi.DomainHash{blockAHash}, coinbaseData, []*externalapi.DomainTransaction{blockCTransaction})
	if err != nil {
		t.Fatalf("Error creating block C: %+v", err)
	}
	// Block D:
	blockDHash, err := consensus.AddBlock(
		[]*externalapi.DomainHash{blockBHash, blockCHash}, coinbaseData, []*externalapi.DomainTransaction{})
	if err != nil {
		t.Fatalf("Error creating block D: %+v", err)
	}
	blockD, err := consensus.GetBlock(blockDHash)
	if err != nil {
		t.Fatalf("Error getting block D: %+v", err)
	}

	// Get the past UTXO set of block D
	csm := consensus.ConsensusStateManager()
	utxoSetIterator, err := csm.RestorePastUTXOSetIterator(blockDHash)
	if err != nil {
		t.Fatalf("Error restoring past UTXO of block D: %+v", err)
	}

	// Build a Multiset for block D
	ms := multiset.New()
	for utxoSetIterator.Next() {
		outpoint, entry, err := utxoSetIterator.Get()
		if err != nil {
			t.Fatalf("Error getting from UTXOSet iterator: %+v", err)
		}
		err = consensus.ConsensusStateManager().AddUTXOToMultiset(ms, entry, outpoint)
		if err != nil {
			t.Fatalf("Error adding utxo to multiset: %+v", err)
		}
	}

	// Turn the multiset into a UTXO commitment
	utxoCommitment := ms.Hash()

	// Make sure that the two commitments are equal
	if *utxoCommitment != blockD.Header.UTXOCommitment {
		t.Fatalf("TestUTXOCommitment: calculated UTXO commitment and "+
			"actual UTXO commitment don't match. Want: %s, got: %s",
			utxoCommitment, blockD.Header.UTXOCommitment)
	}

}
