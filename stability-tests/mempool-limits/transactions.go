package mempoollimits

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	utxopkg "github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"
	"github.com/kaspanet/kaspad/stability-tests/common/mine"
	"github.com/kaspanet/kaspad/util"
)

const (
	payAddress                       = "kaspasim:qzuax2jhawd354e54thhpd9m9wg03pdzwjlpr4vtq3k7xrpumhhtwa2hkr3ep"
	payAddressPrivateKey             = "05d8f681e954a550395ee2297fc1a14f6e801f554c0b9d48cd7165a7ea72ff77"
	fundingCoinbaseTransactionAmount = 1000
	outputsPerTransaction            = 3
	transactionFee                   = 1000
	coinbaseMaturity                 = 100
)

// fundingCoinbaseTransactions contains a collection of transactions
// to be utilized when generating further transactions to fill up
// the mempool.
// It's a separate type because we modify the transactions in place
// whenever we pass an instance of this type into
// submitAnAmountOfTransactionsToTheMempool.
type fundingCoinbaseTransactions struct {
	transactions []*externalapi.DomainTransaction
}

func generateFundingCoinbaseTransactions(t *testing.T, rpcClient *rpcclient.RPCClient) *fundingCoinbaseTransactions {
	// Mine a block, since we need at least one block above the genesis
	// to create a spendable UTXO
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
		_, err := rpcClient.SubmitTransaction(rpcTransaction, false)
		if err != nil {
			if ignoreOrphanRejects && strings.Contains(err.Error(), "orphan") {
				continue
			}
			t.Fatalf("SubmitTransaction: %+v", err)
		}
		log.Infof("Submitted %d transactions", i+1)
	}
}

func mineBlockAndGetCoinbaseTransaction(t *testing.T, rpcClient *rpcclient.RPCClient) *externalapi.DomainTransaction {
	getBlockTemplateResponse, err := rpcClient.GetBlockTemplate(payAddress)
	if err != nil {
		t.Fatalf("GetBlockTemplate: %+v", err)
	}
	templateBlock, err := appmessage.RPCBlockToDomainBlock(getBlockTemplateResponse.Block)
	if err != nil {
		t.Fatalf("RPCBlockToDomainBlock: %+v", err)
	}
	mine.SolveBlock(templateBlock)
	_, err = rpcClient.SubmitBlock(templateBlock)
	if err != nil {
		t.Fatalf("SubmitBlock: %+v", err)
	}
	return templateBlock.Transactions[transactionhelper.CoinbaseTransactionIndex]
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
				t.Fatalf("SignatureScript: %+v", err)
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
		t.Fatalf("DecodeString: %+v", err)
	}
	keyPair, err := secp256k1.DeserializeSchnorrPrivateKeyFromSlice(privateKeyBytes)
	if err != nil {
		t.Fatalf("DeserializeSchnorrPrivateKeyFromSlice: %+v", err)
	}
	return keyPair
}

func buildPayToPayAddressScript(t *testing.T) *externalapi.ScriptPublicKey {
	address, err := util.DecodeAddress(payAddress, dagconfig.SimnetParams.Prefix)
	if err != nil {
		t.Fatalf("DecodeAddress: %+v", err)
	}
	script, err := txscript.PayToAddrScript(address)
	if err != nil {
		t.Fatalf("PayToAddrScript: %+v", err)
	}
	return script
}
