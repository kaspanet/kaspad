package libkaspawallet_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/kaspanet/kaspad/util"
)

func forSchnorrAndECDSA(t *testing.T, testFunc func(t *testing.T, ecdsa bool)) {
	t.Run("schnorr", func(t *testing.T) {
		testFunc(t, false)
	})

	t.Run("ecdsa", func(t *testing.T) {
		testFunc(t, true)
	})
}

func TestMultisig(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		params := &consensusConfig.Params
		forSchnorrAndECDSA(t, func(t *testing.T, ecdsa bool) {
			consensusConfig.BlockCoinbaseMaturity = 0
			tc, teardown, err := consensus.NewFactory().NewTestConsensus(consensusConfig, "TestMultisig")
			if err != nil {
				t.Fatalf("Error setting up tc: %+v", err)
			}
			defer teardown(false)

			const numKeys = 3
			mnemonics := make([]string, numKeys)
			publicKeys := make([]string, numKeys)
			for i := 0; i < numKeys; i++ {
				var err error
				mnemonics[i], err = libkaspawallet.CreateMnemonic()
				if err != nil {
					t.Fatalf("CreateMnemonic: %+v", err)
				}

				publicKeys[i], err = libkaspawallet.MasterPublicKeyFromMnemonic(&consensusConfig.Params, mnemonics[i], true)
				if err != nil {
					t.Fatalf("MasterPublicKeyFromMnemonic: %+v", err)
				}
			}

			const minimumSignatures = 2
			path := "m/1/2/3"
			address, err := libkaspawallet.Address(params, publicKeys, minimumSignatures, path, ecdsa)
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

			fundingBlockHash, _, err := tc.AddBlock([]*externalapi.DomainHash{consensusConfig.GenesisHash}, coinbaseData, nil)
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
			selectedUTXOs := []*libkaspawallet.UTXO{
				{
					Outpoint: &externalapi.DomainOutpoint{
						TransactionID: *consensushashing.TransactionID(block1.Transactions[0]),
						Index:         0,
					},
					UTXOEntry:      utxo.NewUTXOEntry(block1TxOut.Value, block1TxOut.ScriptPublicKey, true, 0),
					DerivationPath: path,
				},
			}

			unsignedTransaction, err := libkaspawallet.CreateUnsignedTransaction(publicKeys, minimumSignatures,
				[]*libkaspawallet.Payment{{
					Address: address,
					Amount:  10,
				}}, selectedUTXOs)
			if err != nil {
				t.Fatalf("CreateUnsignedTransactions: %+v", err)
			}

			isFullySigned, err := libkaspawallet.IsTransactionFullySigned(unsignedTransaction)
			if err != nil {
				t.Fatalf("IsTransactionFullySigned: %+v", err)
			}

			if isFullySigned {
				t.Fatalf("Transaction is not expected to be signed")
			}

			_, err = libkaspawallet.ExtractTransaction(unsignedTransaction, ecdsa)
			if err == nil || !strings.Contains(err.Error(), fmt.Sprintf("missing %d signatures", minimumSignatures)) {
				t.Fatal("Unexpectedly succeed to extract a valid transaction out of unsigned transaction")
			}

			signedTxStep1, err := libkaspawallet.Sign(params, mnemonics[:1], unsignedTransaction, ecdsa)
			if err != nil {
				t.Fatalf("Sign: %+v", err)
			}

			isFullySigned, err = libkaspawallet.IsTransactionFullySigned(signedTxStep1)
			if err != nil {
				t.Fatalf("IsTransactionFullySigned: %+v", err)
			}

			if isFullySigned {
				t.Fatalf("Transaction is not expected to be fully signed")
			}

			signedTxStep2, err := libkaspawallet.Sign(params, mnemonics[1:2], signedTxStep1, ecdsa)
			if err != nil {
				t.Fatalf("Sign: %+v", err)
			}

			extractedSignedTxStep2, err := libkaspawallet.ExtractTransaction(signedTxStep2, ecdsa)
			if err != nil {
				t.Fatalf("ExtractTransaction: %+v", err)
			}

			signedTxOneStep, err := libkaspawallet.Sign(params, mnemonics[:2], unsignedTransaction, ecdsa)
			if err != nil {
				t.Fatalf("Sign: %+v", err)
			}

			extractedSignedTxOneStep, err := libkaspawallet.ExtractTransaction(signedTxOneStep, ecdsa)
			if err != nil {
				t.Fatalf("ExtractTransaction: %+v", err)
			}

			// We check IDs instead of comparing the actual transactions because the actual transactions have different
			// signature scripts due to non deterministic signature scheme.
			if !consensushashing.TransactionID(extractedSignedTxStep2).Equal(consensushashing.TransactionID(extractedSignedTxOneStep)) {
				t.Fatalf("Expected extractedSignedTxOneStep and extractedSignedTxStep2 IDs to be equal")
			}

			_, virtualChangeSet, err := tc.AddBlock([]*externalapi.DomainHash{block1Hash}, nil, []*externalapi.DomainTransaction{extractedSignedTxStep2})
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}

			addedUTXO := &externalapi.DomainOutpoint{
				TransactionID: *consensushashing.TransactionID(extractedSignedTxStep2),
				Index:         0,
			}
			if !virtualChangeSet.VirtualUTXODiff.ToAdd().Contains(addedUTXO) {
				t.Fatalf("Transaction wasn't accepted in the DAG")
			}
		})
	})
}

func TestP2PK(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		params := &consensusConfig.Params
		forSchnorrAndECDSA(t, func(t *testing.T, ecdsa bool) {
			consensusConfig.BlockCoinbaseMaturity = 0
			tc, teardown, err := consensus.NewFactory().NewTestConsensus(consensusConfig, "TestMultisig")
			if err != nil {
				t.Fatalf("Error setting up tc: %+v", err)
			}
			defer teardown(false)

			const numKeys = 1
			mnemonics := make([]string, numKeys)
			publicKeys := make([]string, numKeys)
			for i := 0; i < numKeys; i++ {
				var err error
				mnemonics[i], err = libkaspawallet.CreateMnemonic()
				if err != nil {
					t.Fatalf("CreateMnemonic: %+v", err)
				}

				publicKeys[i], err = libkaspawallet.MasterPublicKeyFromMnemonic(&consensusConfig.Params, mnemonics[i], false)
				if err != nil {
					t.Fatalf("MasterPublicKeyFromMnemonic: %+v", err)
				}
			}

			const minimumSignatures = 1
			path := "m/1/2/3"
			address, err := libkaspawallet.Address(params, publicKeys, minimumSignatures, path, ecdsa)
			if err != nil {
				t.Fatalf("Address: %+v", err)
			}

			if ecdsa {
				if _, ok := address.(*util.AddressPublicKeyECDSA); !ok {
					t.Fatalf("The address is of unexpected type")
				}
			} else {
				if _, ok := address.(*util.AddressPublicKey); !ok {
					t.Fatalf("The address is of unexpected type")
				}
			}

			scriptPublicKey, err := txscript.PayToAddrScript(address)
			if err != nil {
				t.Fatalf("PayToAddrScript: %+v", err)
			}

			coinbaseData := &externalapi.DomainCoinbaseData{
				ScriptPublicKey: scriptPublicKey,
				ExtraData:       nil,
			}

			fundingBlockHash, _, err := tc.AddBlock([]*externalapi.DomainHash{consensusConfig.GenesisHash}, coinbaseData, nil)
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
			selectedUTXOs := []*libkaspawallet.UTXO{
				{
					Outpoint: &externalapi.DomainOutpoint{
						TransactionID: *consensushashing.TransactionID(block1.Transactions[0]),
						Index:         0,
					},
					UTXOEntry:      utxo.NewUTXOEntry(block1TxOut.Value, block1TxOut.ScriptPublicKey, true, 0),
					DerivationPath: path,
				},
			}

			unsignedTransaction, err := libkaspawallet.CreateUnsignedTransaction(publicKeys, minimumSignatures,
				[]*libkaspawallet.Payment{{
					Address: address,
					Amount:  10,
				}}, selectedUTXOs)
			if err != nil {
				t.Fatalf("CreateUnsignedTransactions: %+v", err)
			}

			isFullySigned, err := libkaspawallet.IsTransactionFullySigned(unsignedTransaction)
			if err != nil {
				t.Fatalf("IsTransactionFullySigned: %+v", err)
			}

			if isFullySigned {
				t.Fatalf("Transaction is not expected to be signed")
			}

			_, err = libkaspawallet.ExtractTransaction(unsignedTransaction, ecdsa)
			if err == nil || !strings.Contains(err.Error(), "missing signature") {
				t.Fatal("Unexpectedly succeed to extract a valid transaction out of unsigned transaction")
			}

			signedTx, err := libkaspawallet.Sign(params, mnemonics, unsignedTransaction, ecdsa)
			if err != nil {
				t.Fatalf("Sign: %+v", err)
			}

			tx, err := libkaspawallet.ExtractTransaction(signedTx, ecdsa)
			if err != nil {
				t.Fatalf("ExtractTransaction: %+v", err)
			}

			_, virtualChangeSet, err := tc.AddBlock([]*externalapi.DomainHash{block1Hash}, nil, []*externalapi.DomainTransaction{tx})
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}

			addedUTXO := &externalapi.DomainOutpoint{
				TransactionID: *consensushashing.TransactionID(tx),
				Index:         0,
			}
			if !virtualChangeSet.VirtualUTXODiff.ToAdd().Contains(addedUTXO) {
				t.Fatalf("Transaction wasn't accepted in the DAG")
			}
		})
	})
}
