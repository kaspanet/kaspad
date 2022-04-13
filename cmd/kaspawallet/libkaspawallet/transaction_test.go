package libkaspawallet_test

import (
	"fmt"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/pkg/errors"
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

			_, err = libkaspawallet.DeserializedTransactionFromSerializedPartiallySigned(unsignedTransaction, ecdsa)
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

			extractedSignedTxStep2, err := libkaspawallet.DeserializedTransactionFromSerializedPartiallySigned(signedTxStep2, ecdsa)
			if err != nil {
				t.Fatalf("DeserializedTransactionFromSerializedPartiallySigned: %+v", err)
			}

			signedTxOneStep, err := libkaspawallet.Sign(params, mnemonics[:2], unsignedTransaction, ecdsa)
			if err != nil {
				t.Fatalf("Sign: %+v", err)
			}

			extractedSignedTxOneStep, err := libkaspawallet.DeserializedTransactionFromSerializedPartiallySigned(signedTxOneStep, ecdsa)
			if err != nil {
				t.Fatalf("DeserializedTransactionFromSerializedPartiallySigned: %+v", err)
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

			_, err = libkaspawallet.DeserializedTransactionFromSerializedPartiallySigned(unsignedTransaction, ecdsa)
			if err == nil || !strings.Contains(err.Error(), "missing signature") {
				t.Fatal("Unexpectedly succeed to extract a valid transaction out of unsigned transaction")
			}

			signedTx, err := libkaspawallet.Sign(params, mnemonics, unsignedTransaction, ecdsa)
			if err != nil {
				t.Fatalf("Sign: %+v", err)
			}

			tx, err := libkaspawallet.DeserializedTransactionFromSerializedPartiallySigned(signedTx, ecdsa)
			if err != nil {
				t.Fatalf("DeserializedTransactionFromSerializedPartiallySigned: %+v", err)
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

func TestMaxSompi(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		params := &consensusConfig.Params
		cfg := *consensusConfig
		cfg.BlockCoinbaseMaturity = 0
		cfg.PreDeflationaryPhaseBaseSubsidy = 20e6 * constants.SompiPerKaspa
		cfg.HF1DAAScore = cfg.GenesisBlock.Header.DAAScore() + 10
		tc, teardown, err := consensus.NewFactory().NewTestConsensus(&cfg, "TestMaxSompi")
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

			publicKeys[i], err = libkaspawallet.MasterPublicKeyFromMnemonic(&cfg.Params, mnemonics[i], false)
			if err != nil {
				t.Fatalf("MasterPublicKeyFromMnemonic: %+v", err)
			}
		}

		const minimumSignatures = 1
		path := "m/1/2/3"
		address, err := libkaspawallet.Address(params, publicKeys, minimumSignatures, path, false)
		if err != nil {
			t.Fatalf("Address: %+v", err)
		}

		scriptPublicKey, err := txscript.PayToAddrScript(address)
		if err != nil {
			t.Fatalf("PayToAddrScript: %+v", err)
		}

		coinbaseData := &externalapi.DomainCoinbaseData{
			ScriptPublicKey: scriptPublicKey,
			ExtraData:       nil,
		}

		fundingBlock1Hash, _, err := tc.AddBlock([]*externalapi.DomainHash{cfg.GenesisHash}, coinbaseData, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		fundingBlock2Hash, _, err := tc.AddBlock([]*externalapi.DomainHash{fundingBlock1Hash}, coinbaseData, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		fundingBlock3Hash, _, err := tc.AddBlock([]*externalapi.DomainHash{fundingBlock2Hash}, coinbaseData, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		fundingBlock4Hash, _, err := tc.AddBlock([]*externalapi.DomainHash{fundingBlock3Hash}, coinbaseData, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		fundingBlock2, err := tc.GetBlock(fundingBlock2Hash)
		if err != nil {
			t.Fatalf("GetBlock: %+v", err)
		}

		fundingBlock3, err := tc.GetBlock(fundingBlock3Hash)
		if err != nil {
			t.Fatalf("GetBlock: %+v", err)
		}

		fundingBlock4, err := tc.GetBlock(fundingBlock4Hash)
		if err != nil {
			t.Fatalf("GetBlock: %+v", err)
		}

		block1Hash, _, err := tc.AddBlock([]*externalapi.DomainHash{fundingBlock4Hash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		block1, err := tc.GetBlock(block1Hash)
		if err != nil {
			t.Fatalf("GetBlock: %+v", err)
		}

		txOut1 := fundingBlock2.Transactions[0].Outputs[0]
		txOut2 := fundingBlock3.Transactions[0].Outputs[0]
		txOut3 := fundingBlock4.Transactions[0].Outputs[0]
		txOut4 := block1.Transactions[0].Outputs[0]
		selectedUTXOsForTxWithLargeInputAmount := []*libkaspawallet.UTXO{
			{
				Outpoint: &externalapi.DomainOutpoint{
					TransactionID: *consensushashing.TransactionID(fundingBlock2.Transactions[0]),
					Index:         0,
				},
				UTXOEntry:      utxo.NewUTXOEntry(txOut1.Value, txOut1.ScriptPublicKey, true, 0),
				DerivationPath: path,
			},
			{
				Outpoint: &externalapi.DomainOutpoint{
					TransactionID: *consensushashing.TransactionID(fundingBlock3.Transactions[0]),
					Index:         0,
				},
				UTXOEntry:      utxo.NewUTXOEntry(txOut2.Value, txOut2.ScriptPublicKey, true, 0),
				DerivationPath: path,
			},
		}

		unsignedTxWithLargeInputAmount, err := libkaspawallet.CreateUnsignedTransaction(publicKeys, minimumSignatures,
			[]*libkaspawallet.Payment{{
				Address: address,
				Amount:  10,
			}}, selectedUTXOsForTxWithLargeInputAmount)
		if err != nil {
			t.Fatalf("CreateUnsignedTransactions: %+v", err)
		}

		signedTxWithLargeInputAmount, err := libkaspawallet.Sign(params, mnemonics, unsignedTxWithLargeInputAmount, false)
		if err != nil {
			t.Fatalf("Sign: %+v", err)
		}

		txWithLargeInputAmount, err := libkaspawallet.DeserializedTransactionFromSerializedPartiallySigned(signedTxWithLargeInputAmount, false)
		if err != nil {
			t.Fatalf("DeserializedTransactionFromSerializedPartiallySigned: %+v", err)
		}

		_, virtualChangeSet, err := tc.AddBlock([]*externalapi.DomainHash{block1Hash}, nil, []*externalapi.DomainTransaction{txWithLargeInputAmount})
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		addedUTXO1 := &externalapi.DomainOutpoint{
			TransactionID: *consensushashing.TransactionID(txWithLargeInputAmount),
			Index:         0,
		}
		if virtualChangeSet.VirtualUTXODiff.ToAdd().Contains(addedUTXO1) {
			t.Fatalf("Transaction was accepted in the DAG")
		}

		selectedUTXOsForTxWithLargeInputAndOutputAmount := []*libkaspawallet.UTXO{
			{
				Outpoint: &externalapi.DomainOutpoint{
					TransactionID: *consensushashing.TransactionID(fundingBlock4.Transactions[0]),
					Index:         0,
				},
				UTXOEntry:      utxo.NewUTXOEntry(txOut3.Value, txOut3.ScriptPublicKey, true, 0),
				DerivationPath: path,
			},
			{
				Outpoint: &externalapi.DomainOutpoint{
					TransactionID: *consensushashing.TransactionID(block1.Transactions[0]),
					Index:         0,
				},
				UTXOEntry:      utxo.NewUTXOEntry(txOut4.Value, txOut4.ScriptPublicKey, true, 0),
				DerivationPath: path,
			},
		}

		unsignedTxWithLargeInputAndOutputAmount, err := libkaspawallet.CreateUnsignedTransaction(publicKeys, minimumSignatures,
			[]*libkaspawallet.Payment{{
				Address: address,
				Amount:  22e6 * constants.SompiPerKaspa,
			}}, selectedUTXOsForTxWithLargeInputAndOutputAmount)
		if err != nil {
			t.Fatalf("CreateUnsignedTransactions: %+v", err)
		}

		signedTxWithLargeInputAndOutputAmount, err := libkaspawallet.Sign(params, mnemonics, unsignedTxWithLargeInputAndOutputAmount, false)
		if err != nil {
			t.Fatalf("Sign: %+v", err)
		}

		txWithLargeInputAndOutputAmount, err := libkaspawallet.DeserializedTransactionFromSerializedPartiallySigned(signedTxWithLargeInputAndOutputAmount, false)
		if err != nil {
			t.Fatalf("DeserializedTransactionFromSerializedPartiallySigned: %+v", err)
		}

		_, _, err = tc.AddBlock([]*externalapi.DomainHash{block1Hash}, nil, []*externalapi.DomainTransaction{txWithLargeInputAndOutputAmount})
		if !errors.Is(err, ruleerrors.ErrBadTxOutValue) {
			t.Fatalf("AddBlock: %+v", err)
		}

		tip := block1Hash
		for {
			tip, _, err = tc.AddBlock([]*externalapi.DomainHash{tip}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}

			selectedTip, err := tc.GetVirtualSelectedParent()
			if err != nil {
				t.Fatalf("GetVirtualDAAScore: %+v", err)
			}

			daaScore, err := tc.DAABlocksStore().DAAScore(tc.DatabaseContext(), model.NewStagingArea(), selectedTip)
			if err != nil {
				t.Fatalf("DAAScore: %+v", err)
			}

			if daaScore >= cfg.HF1DAAScore {
				break
			}
		}

		tip, virtualChangeSet, err = tc.AddBlock([]*externalapi.DomainHash{tip}, nil, []*externalapi.DomainTransaction{txWithLargeInputAndOutputAmount})
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		addedUTXO2 := &externalapi.DomainOutpoint{
			TransactionID: *consensushashing.TransactionID(txWithLargeInputAndOutputAmount),
			Index:         0,
		}

		if !virtualChangeSet.VirtualUTXODiff.ToAdd().Contains(addedUTXO2) {
			t.Fatalf("txWithLargeInputAndOutputAmount weren't accepted in the DAG")
		}

		_, virtualChangeSet, err = tc.AddBlock([]*externalapi.DomainHash{tip}, nil, []*externalapi.DomainTransaction{txWithLargeInputAmount})
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		if !virtualChangeSet.VirtualUTXODiff.ToAdd().Contains(addedUTXO1) {
			t.Fatalf("txWithLargeInputAmount wasn't accepted in the DAG")
		}
	})
}
