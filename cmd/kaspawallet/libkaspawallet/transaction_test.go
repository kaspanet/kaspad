package libkaspawallet_test

import (
	"fmt"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/util"
	"strings"
	"testing"
)

func TestMultisig(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		params.BlockCoinbaseMaturity = 0
		tc, teardown, err := consensus.NewFactory().NewTestConsensus(params, false, "TestMultisig")
		if err != nil {
			t.Fatalf("Error setting up tc: %+v", err)
		}
		defer teardown(false)

		const numKeys = 3
		privateKeys := make([][]byte, numKeys)
		publicKeys := make([][]byte, numKeys)
		for i := 0; i < numKeys; i++ {
			privateKeys[i], publicKeys[i], err = libkaspawallet.CreateKeyPair()
			if err != nil {
				t.Fatalf("CreateKeyPair: %+v", err)
			}
		}

		const minimumSignatures = 2
		address, err := libkaspawallet.Address(params, publicKeys, minimumSignatures)
		if err != nil {
			t.Fatalf("Address: %+v", err)
		}

		if _, ok := address.(*util.AddressScriptHash); !ok {
			t.Fatalf("The address is of unexpected type")
		}

		scriptPublicKey, err := txscript.PayToAddrScript(address)
		if err != nil {
			t.Fatalf("PayToAddrScript: %+v", err)
		}

		coinbaseData := &externalapi.DomainCoinbaseData{
			ScriptPublicKey: scriptPublicKey,
			ExtraData:       nil,
		}

		fundingBlockHash, _, err := tc.AddBlock([]*externalapi.DomainHash{params.GenesisHash}, coinbaseData, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		block1Hash, _, err := tc.AddBlock([]*externalapi.DomainHash{fundingBlockHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		block1, err := tc.GetBlock(block1Hash)
		if err != nil {
			t.Fatalf("GetBlock: %+v", err)
		}

		block1Tx := block1.Transactions[0]
		block1TxOut := block1Tx.Outputs[0]
		selectedUTXOs := []*externalapi.OutpointAndUTXOEntryPair{{
			Outpoint: &externalapi.DomainOutpoint{
				TransactionID: *consensushashing.TransactionID(block1.Transactions[0]),
				Index:         0,
			},
			UTXOEntry: utxo.NewUTXOEntry(block1TxOut.Value, block1TxOut.ScriptPublicKey, true, 0),
		}}

		unsignedTransaction, err := libkaspawallet.CreateUnsignedTransaction(publicKeys, minimumSignatures, []*libkaspawallet.Payment{{
			Address: address,
			Amount:  10,
		}}, selectedUTXOs)
		if err != nil {
			t.Fatalf("CreateUnsignedTransaction: %+v", err)
		}

		isFullySigned, err := libkaspawallet.IsTransactionFullySigned(unsignedTransaction)
		if err != nil {
			t.Fatalf("IsTransactionFullySigned: %+v", err)
		}

		if isFullySigned {
			t.Fatalf("Transaction is not expected to be signed")
		}

		_, err = libkaspawallet.ExtractTransaction(unsignedTransaction)
		if err == nil || !strings.Contains(err.Error(), fmt.Sprintf("missing %d signatures", minimumSignatures)) {
			t.Fatal("Unexpectedly succeed to extract a valid transaction out of unsigned transaction")
		}

		signedTxStep1, err := libkaspawallet.Sign(privateKeys[:1], unsignedTransaction)
		if err != nil {
			t.Fatalf("IsTransactionFullySigned: %+v", err)
		}

		isFullySigned, err = libkaspawallet.IsTransactionFullySigned(signedTxStep1)
		if err != nil {
			t.Fatalf("IsTransactionFullySigned: %+v", err)
		}

		if isFullySigned {
			t.Fatalf("Transaction is not expected to be fully signed")
		}

		signedTxStep2, err := libkaspawallet.Sign(privateKeys[1:2], signedTxStep1)
		if err != nil {
			t.Fatalf("Sign: %+v", err)
		}

		extractedSignedTxStep2, err := libkaspawallet.ExtractTransaction(signedTxStep2)
		if err != nil {
			t.Fatalf("ExtractTransaction: %+v", err)
		}

		signedTxOneStep, err := libkaspawallet.Sign(privateKeys[:2], unsignedTransaction)
		if err != nil {
			t.Fatalf("Sign: %+v", err)
		}

		extractedSignedTxOneStep, err := libkaspawallet.ExtractTransaction(signedTxOneStep)
		if err != nil {
			t.Fatalf("ExtractTransaction: %+v", err)
		}

		// We check IDs instead of comparing the actual transactions because the actual transactions have different
		// signature scripts due to non deterministic signature scheme.
		if !consensushashing.TransactionID(extractedSignedTxStep2).Equal(consensushashing.TransactionID(extractedSignedTxOneStep)) {
			t.Fatalf("Expected extractedSignedTxOneStep and extractedSignedTxStep2 IDs to be equal")
		}

		_, insertionResult, err := tc.AddBlock([]*externalapi.DomainHash{block1Hash}, nil, []*externalapi.DomainTransaction{extractedSignedTxStep2})
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		addedUTXO := &externalapi.DomainOutpoint{
			TransactionID: *consensushashing.TransactionID(extractedSignedTxStep2),
			Index:         0,
		}
		if !insertionResult.VirtualUTXODiff.ToAdd().Contains(addedUTXO) {
			t.Fatalf("Transaction wasn't accepted in the DAG")
		}
	})
}

func TestP2PK(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		params.BlockCoinbaseMaturity = 0
		tc, teardown, err := consensus.NewFactory().NewTestConsensus(params, false, "TestMultisig")
		if err != nil {
			t.Fatalf("Error setting up tc: %+v", err)
		}
		defer teardown(false)

		const numKeys = 1
		privateKeys := make([][]byte, numKeys)
		publicKeys := make([][]byte, numKeys)
		for i := 0; i < numKeys; i++ {
			privateKeys[i], publicKeys[i], err = libkaspawallet.CreateKeyPair()
			if err != nil {
				t.Fatalf("CreateKeyPair: %+v", err)
			}
		}

		const minimumSignatures = 1
		address, err := libkaspawallet.Address(params, publicKeys, minimumSignatures)
		if err != nil {
			t.Fatalf("Address: %+v", err)
		}

		if _, ok := address.(*util.AddressPublicKey); !ok {
			t.Fatalf("The address is of unexpected type")
		}

		scriptPublicKey, err := txscript.PayToAddrScript(address)
		if err != nil {
			t.Fatalf("PayToAddrScript: %+v", err)
		}

		coinbaseData := &externalapi.DomainCoinbaseData{
			ScriptPublicKey: scriptPublicKey,
			ExtraData:       nil,
		}

		fundingBlockHash, _, err := tc.AddBlock([]*externalapi.DomainHash{params.GenesisHash}, coinbaseData, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		block1Hash, _, err := tc.AddBlock([]*externalapi.DomainHash{fundingBlockHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		block1, err := tc.GetBlock(block1Hash)
		if err != nil {
			t.Fatalf("GetBlock: %+v", err)
		}

		block1Tx := block1.Transactions[0]
		block1TxOut := block1Tx.Outputs[0]
		selectedUTXOs := []*externalapi.OutpointAndUTXOEntryPair{{
			Outpoint: &externalapi.DomainOutpoint{
				TransactionID: *consensushashing.TransactionID(block1.Transactions[0]),
				Index:         0,
			},
			UTXOEntry: utxo.NewUTXOEntry(block1TxOut.Value, block1TxOut.ScriptPublicKey, true, 0),
		}}

		unsignedTransaction, err := libkaspawallet.CreateUnsignedTransaction(publicKeys, minimumSignatures, []*libkaspawallet.Payment{{
			Address: address,
			Amount:  10,
		}}, selectedUTXOs)
		if err != nil {
			t.Fatalf("CreateUnsignedTransaction: %+v", err)
		}

		isFullySigned, err := libkaspawallet.IsTransactionFullySigned(unsignedTransaction)
		if err != nil {
			t.Fatalf("IsTransactionFullySigned: %+v", err)
		}

		if isFullySigned {
			t.Fatalf("Transaction is not expected to be signed")
		}

		_, err = libkaspawallet.ExtractTransaction(unsignedTransaction)
		if err == nil || !strings.Contains(err.Error(), "missing signature") {
			t.Fatal("Unexpectedly succeed to extract a valid transaction out of unsigned transaction")
		}

		signedTx, err := libkaspawallet.Sign(privateKeys, unsignedTransaction)
		if err != nil {
			t.Fatalf("IsTransactionFullySigned: %+v", err)
		}

		tx, err := libkaspawallet.ExtractTransaction(signedTx)
		if err != nil {
			t.Fatalf("ExtractTransaction: %+v", err)
		}

		_, insertionResult, err := tc.AddBlock([]*externalapi.DomainHash{block1Hash}, nil, []*externalapi.DomainTransaction{tx})
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		addedUTXO := &externalapi.DomainOutpoint{
			TransactionID: *consensushashing.TransactionID(tx),
			Index:         0,
		}
		if !insertionResult.VirtualUTXODiff.ToAdd().Contains(addedUTXO) {
			t.Fatalf("Transaction wasn't accepted in the DAG")
		}
	})
}
