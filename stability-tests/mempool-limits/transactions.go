package main

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"
	"github.com/kaspanet/kaspad/stability-tests/common/mine"
)

func submitAnAmountOfTransactionsToTheMempool(rpcClient *rpcclient.RPCClient, amountToSubmit int) {
	fundingBlocksToGenerate := amountToSubmit / maxTransactionsInBlock

	coinbaseTransactions := generateCoinbaseTransactions(rpcClient, fundingBlocksToGenerate)
	for _, coinbaseTransaction := range coinbaseTransactions {
		spendingTransactions := generateTransactionsSpendingCoinbaseTransaction(coinbaseTransaction)
		for _, transaction := range spendingTransactions {
			rpcTransaction := appmessage.DomainTransactionToRPCTransaction(transaction)
			_, err := rpcClient.SubmitTransaction(rpcTransaction)
			if err != nil {
				panic(err)
			}
		}
	}
}

func generateCoinbaseTransactions(rpcClient *rpcclient.RPCClient, coinbaseTransactionAmountToGenerate int) []*externalapi.DomainTransaction {
	coinbaseTransactions := make([]*externalapi.DomainTransaction, coinbaseTransactionAmountToGenerate)
	for i := 0; i < coinbaseTransactionAmountToGenerate; i++ {
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
		coinbaseTransactions[i] = templateBlock.Transactions[0]
	}
	return coinbaseTransactions
}

func generateTransactionsSpendingCoinbaseTransaction(coinbaseTransaction *externalapi.DomainTransaction) []*externalapi.DomainTransaction {
	transactions := make([]*externalapi.DomainTransaction, maxTransactionsInBlock)
	for i := 0; i < maxTransactionsInBlock; i++ {
		transactions[i] = &externalapi.DomainTransaction{}
	}
	return transactions
}
