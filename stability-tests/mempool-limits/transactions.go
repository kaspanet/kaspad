package main

import (
	"encoding/hex"
	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	utxopkg "github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"
	"github.com/kaspanet/kaspad/stability-tests/common/mine"
	"github.com/kaspanet/kaspad/util"
	"strings"
	"testing"
)

const (
	payAddress                       = "kaspasim:qr79e37hxdgkn4xjjmfxvqvayc5gsmsql2660d08u9ej9vnc8lzcywr265u64"
	payAddressPrivateKey             = "0ec5d7308f65717f3f0c3e4d962d73056c1c255a16593b3989589281b51ad5bc"
	fundingCoinbaseTransactionAmount = 1000
	outputsPerTransaction            = 3
	transactionFee                   = 1000
	coinbaseMaturity                 = 100
)

type fundingCoinbaseTransactions struct {
	transactions []*externalapi.DomainTransaction
}

func generateFundingCoinbaseTransactions(t *testing.T, rpcClient *rpcclient.RPCClient) *fundingCoinbaseTransactions {
	// Generate one coinbase transaction for its side effect:
	// the block containing it accepts the empty genesis coinbase
	mineBlockAndGetCoinbaseTransaction(t, rpcClient)

	log.Infof("Generating funding coinbase transactions")
	fundingCoinbaseTransactions := &fundingCoinbaseTransactions{
		transactions: make([]*externalapi.DomainTransaction, fundingCoinbaseTransactionAmount),
	}
	for i := 0; i < fundingCoinbaseTransactionAmount; i++ {
		fundingCoinbaseTransactions.transactions[i] = mineBlockAndGetCoinbaseTransaction(t, rpcClient)
	}

	log.Infof("Maturing funding coinbase transactions")
	for i := 0; i < coinbaseMaturity; i++ {
		mineBlockAndGetCoinbaseTransaction(t, rpcClient)
	}

	return fundingCoinbaseTransactions
}

func submitAnAmountOfTransactionsToTheMempool(t *testing.T, rpcClient *rpcclient.RPCClient,
	payAddressKeyPair *secp256k1.SchnorrKeyPair, payToPayAddressScript *externalapi.ScriptPublicKey,
	fundingTransactions *fundingCoinbaseTransactions, amountToSubmit int, ignoreOrphanRejects bool) {

	log.Infof("Generating %d transactions", amountToSubmit)
	transactions := make([]*externalapi.DomainTransaction, 0)
	for len(transactions) < amountToSubmit {
		var coinbaseTransaction *externalapi.DomainTransaction
		coinbaseTransaction, fundingTransactions.transactions = fundingTransactions.transactions[0], fundingTransactions.transactions[1:]

		unspentTransactions := []*externalapi.DomainTransaction{coinbaseTransaction}
		for len(transactions) < amountToSubmit && len(unspentTransactions) > 0 {
			var transactionToSpend *externalapi.DomainTransaction
			transactionToSpend, unspentTransactions = unspentTransactions[0], unspentTransactions[1:]
			spendingTransactions := generateTransactionsWithMultipleOutputs(t, payAddressKeyPair, payToPayAddressScript, transactionToSpend)
			transactions = append(transactions, spendingTransactions...)
			unspentTransactions = append(unspentTransactions, spendingTransactions...)
		}
		log.Infof("Generated %d transactions", len(transactions))
	}

	transactions = transactions[:amountToSubmit]
	log.Infof("Submitting %d transactions", len(transactions))

	for i, transaction := range transactions {
		rpcTransaction := appmessage.DomainTransactionToRPCTransaction(transaction)
		_, err := rpcClient.SubmitTransaction(rpcTransaction)
		if err != nil {
			if ignoreOrphanRejects && strings.Contains(err.Error(), "orphan") {
				continue
			}
			t.Fatalf("SubmitTransaction: %s", err)
		}
		log.Infof("Submitted %d transactions", i+1)
	}
}

func mineBlockAndGetCoinbaseTransaction(t *testing.T, rpcClient *rpcclient.RPCClient) *externalapi.DomainTransaction {
	getBlockTemplateResponse, err := rpcClient.GetBlockTemplate(payAddress)
	if err != nil {
		t.Fatalf("GetBlockTemplate: %s", err)
	}
	templateBlock, err := appmessage.RPCBlockToDomainBlock(getBlockTemplateResponse.Block)
	if err != nil {
		t.Fatalf("RPCBlockToDomainBlock: %s", err)
	}
	mine.SolveBlock(templateBlock)
	_, err = rpcClient.SubmitBlock(templateBlock)
	if err != nil {
		t.Fatalf("SubmitBlock: %s", err)
	}
	return templateBlock.Transactions[0]
}

func generateTransactionsWithMultipleOutputs(t *testing.T,
	payAddressKeyPair *secp256k1.SchnorrKeyPair, payToPayAddressScript *externalapi.ScriptPublicKey,
	fundingTransaction *externalapi.DomainTransaction) []*externalapi.DomainTransaction {

	var transactions []*externalapi.DomainTransaction
	for fundingTransactionOutputIndex, fundingTransactionOutput := range fundingTransaction.Outputs {
		if fundingTransactionOutput.Value < transactionFee {
			continue
		}
		outputValue := (fundingTransactionOutput.Value - transactionFee) / outputsPerTransaction

		fundingTransactionID := consensushashing.TransactionID(fundingTransaction)
		spendingTransactionInputs := []*externalapi.DomainTransactionInput{
			{
				PreviousOutpoint: externalapi.DomainOutpoint{
					TransactionID: *fundingTransactionID,
					Index:         uint32(fundingTransactionOutputIndex),
				},
				UTXOEntry: utxopkg.NewUTXOEntry(
					fundingTransactionOutput.Value,
					payToPayAddressScript,
					false,
					0),
			},
		}

		spendingTransactionOutputs := make([]*externalapi.DomainTransactionOutput, outputsPerTransaction)
		for i := 0; i < outputsPerTransaction; i++ {
			spendingTransactionOutputs[i] = &externalapi.DomainTransactionOutput{
				Value:           outputValue,
				ScriptPublicKey: payToPayAddressScript,
			}
		}

		spendingTransaction := &externalapi.DomainTransaction{
			Version:      constants.MaxTransactionVersion,
			Inputs:       spendingTransactionInputs,
			Outputs:      spendingTransactionOutputs,
			LockTime:     0,
			SubnetworkID: subnetworks.SubnetworkIDNative,
			Gas:          0,
			Payload:      nil,
		}

		for spendingTransactionInputIndex, spendingTransactionInput := range spendingTransactionInputs {
			signatureScript, err := txscript.SignatureScript(
				spendingTransaction,
				spendingTransactionInputIndex,
				consensushashing.SigHashAll,
				payAddressKeyPair,
				&consensushashing.SighashReusedValues{})
			if err != nil {
				t.Fatalf("SignatureScript: %s", err)
			}
			spendingTransactionInput.SignatureScript = signatureScript
		}

		transactions = append(transactions, spendingTransaction)
	}
	return transactions
}

func decodePayAddressKeyPair(t *testing.T) *secp256k1.SchnorrKeyPair {
	privateKeyBytes, err := hex.DecodeString(payAddressPrivateKey)
	if err != nil {
		t.Fatalf("DecodeString: %s", err)
	}
	keyPair, err := secp256k1.DeserializeSchnorrPrivateKeyFromSlice(privateKeyBytes)
	if err != nil {
		t.Fatalf("DeserializeSchnorrPrivateKeyFromSlice: %s", err)
	}
	return keyPair
}

func buildPayToPayAddressScript(t *testing.T) *externalapi.ScriptPublicKey {
	address, err := util.DecodeAddress(payAddress, dagconfig.SimnetParams.Prefix)
	if err != nil {
		t.Fatalf("DecodeAddress: %s", err)
	}
	script, err := txscript.PayToAddrScript(address)
	if err != nil {
		t.Fatalf("PayToAddrScript: %s", err)
	}
	return script
}
