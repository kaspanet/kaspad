package controllers

import (
	"fmt"
	"github.com/daglabs/btcd/apiserver/database"
	"github.com/daglabs/btcd/apiserver/models"
	"github.com/daglabs/btcd/apiserver/utils"
	"github.com/daglabs/btcd/util/daghash"
	"github.com/jinzhu/gorm"
	"net/http"
)

const maximumGetTransactionsLimit = 1000

// GetTransactionByIDHandler returns a transaction by a given transaction ID.
func GetTransactionByIDHandler(txID string) (interface{}, *utils.HandlerError) {
	if len(txID) != daghash.TxIDSize*2 {
		return nil, utils.NewHandlerError(http.StatusUnprocessableEntity, fmt.Sprintf("The given txid is not a hex-encoded %d-byte hash.", daghash.TxIDSize))
	}
	tx := &models.Transaction{}
	db := database.DB.Where("transaction_id = ?", txID)
	addTxPreloadedFields(db).First(&tx)
	if tx.ID == 0 {
		return nil, utils.NewHandlerError(http.StatusNotFound, "No transaction with the given txid was found.")
	}
	return convertTxModelToTxResponse(tx), nil
}

// GetTransactionByHashHandler returns a transaction by a given transaction hash.
func GetTransactionByHashHandler(txHash string) (interface{}, *utils.HandlerError) {
	if len(txHash) != daghash.HashSize*2 {
		return nil, utils.NewHandlerError(http.StatusUnprocessableEntity, fmt.Sprintf("The given txhash is not a hex-encoded %d-byte hash.", daghash.HashSize))
	}
	tx := &models.Transaction{}
	db := database.DB.
		Where("transaction_hash = ?", txHash)
	addTxPreloadedFields(db).First(&tx)
	if tx.ID == 0 {
		return nil, utils.NewHandlerError(http.StatusNotFound, "No transaction with the given txhash was found.")
	}
	return convertTxModelToTxResponse(tx), nil
}

// GetTransactionsByAddressHandler searches for all transactions
// where the given address is either an input or an output.
func GetTransactionsByAddressHandler(address string, skip uint64, limit uint64) (interface{}, *utils.HandlerError) {
	if limit > 1000 {
		return nil, utils.NewHandlerError(http.StatusUnprocessableEntity, fmt.Sprintf("The maximum allowed value for the limit is %d", maximumGetTransactionsLimit))
	}
	txs := []*models.Transaction{}
	db := database.DB.
		Joins("INNER JOIN `transaction_outputs` ON `transaction_outputs.transaction_id` = `transactions.id`").
		Joins("INNER JOIN `addresses` ON `addresses.id` = `transaction_outputs.address_id`").
		Where("addresses.address = ?", address).
		Limit(limit).
		Offset(skip)
	addTxPreloadedFields(db).Find(&txs)
	txResponses := make([]*transactionResponse, len(txs))
	for i, tx := range txs {
		txResponses[i] = convertTxModelToTxResponse(tx)
	}
	return txs, nil
}

func addTxPreloadedFields(db *gorm.DB) *gorm.DB {
	return db.Preload("AcceptingBlock").
		Preload("Subnetwork").
		Preload("TransactionOutputs").
		Preload("TransactionInputs.TransactionOutput.Transaction")
}
