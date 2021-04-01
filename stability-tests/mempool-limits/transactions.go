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
)

const (
	payAddress            = "kaspasim:qr79e37hxdgkn4xjjmfxvqvayc5gsmsql2660d08u9ej9vnc8lzcywr265u64"
	payAddressPrivateKey  = "0ec5d7308f65717f3f0c3e4d962d73056c1c255a16593b3989589281b51ad5bc"
	outputsPerTransaction = 3
	transactionFee        = 1000
	coinbaseMaturity      = 100
)

var (
	payAddressKeyPair     = decodePayAddressKeyPair()
	payToPayAddressScript = buildPayToPayAddressScript()
)

func submitAnAmountOfTransactionsToTheMempool(rpcClient *rpcclient.RPCClient, amountToSubmit int) {
	log.Infof("Generating %d transactions", amountToSubmit)
	transactions := make([]*externalapi.DomainTransaction, 0)
	for len(transactions) < amountToSubmit {
		log.Infof("Generating funding coinbase transaction")
		coinbaseTransaction := generateCoinbaseTransaction(rpcClient)

		unspentTransactions := []*externalapi.DomainTransaction{coinbaseTransaction}
		for len(transactions) < amountToSubmit && len(unspentTransactions) > 0 {
			var transactionToSpend *externalapi.DomainTransaction
			transactionToSpend, unspentTransactions = unspentTransactions[0], unspentTransactions[1:]
			spendingTransactions := generateTransactionsWithLotsOfOutputs(transactionToSpend)
			transactions = append(transactions, spendingTransactions...)
			unspentTransactions = append(unspentTransactions, spendingTransactions...)
		}
		log.Infof("Generated %d transactions", len(transactions))
	}

	log.Infof("Maturing funding coinbase transactions")
	for i := 0; i < coinbaseMaturity; i++ {
		generateCoinbaseTransaction(rpcClient)
	}

	transactions = transactions[:amountToSubmit]
	log.Infof("Submitting %d transactions", len(transactions))

	for i, transaction := range transactions {
		rpcTransaction := appmessage.DomainTransactionToRPCTransaction(transaction)
		_, err := rpcClient.SubmitTransaction(rpcTransaction)
		if err != nil {
			panic(err)
		}
		log.Infof("Submitted %d transactions", i+1)
	}
}

func generateCoinbaseTransaction(rpcClient *rpcclient.RPCClient) *externalapi.DomainTransaction {
	getBlockTemplateResponse, err := rpcClient.GetBlockTemplate(payAddress)
	if err != nil {
		panic(err)
	}
	templateBlock, err := appmessage.RPCBlockToDomainBlock(getBlockTemplateResponse.Block)
	if err != nil {
		panic(err)
	}
	mine.SolveBlock(templateBlock)
	_, err = rpcClient.SubmitBlock(templateBlock)
	if err != nil {
		panic(err)
	}
	return templateBlock.Transactions[0]
}

func generateTransactionsWithLotsOfOutputs(fundingTransaction *externalapi.DomainTransaction) []*externalapi.DomainTransaction {
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
				panic(err)
			}
			spendingTransactionInput.SignatureScript = signatureScript
		}

		transactions = append(transactions, spendingTransaction)
	}
	return transactions
}

func decodePayAddressKeyPair() *secp256k1.SchnorrKeyPair {
	privateKeyBytes, err := hex.DecodeString(payAddressPrivateKey)
	if err != nil {
		panic(err)
	}
	keyPair, err := secp256k1.DeserializeSchnorrPrivateKeyFromSlice(privateKeyBytes)
	if err != nil {
		panic(err)
	}
	return keyPair
}

func buildPayToPayAddressScript() *externalapi.ScriptPublicKey {
	address, err := util.DecodeAddress(payAddress, dagconfig.SimnetParams.Prefix)
	if err != nil {
		panic(err)
	}
	script, err := txscript.PayToAddrScript(address)
	if err != nil {
		panic(err)
	}
	return script
}
