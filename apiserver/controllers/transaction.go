package controllers

import (
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/daglabs/btcd/apiserver/database"
	"github.com/daglabs/btcd/apiserver/models"
	"github.com/daglabs/btcd/apiserver/utils"
	"github.com/daglabs/btcd/util/daghash"
	"github.com/jinzhu/gorm"
)

const maximumGetTransactionsLimit = 1000

// GetTransactionByIDHandler returns a transaction by a given transaction ID.
func GetTransactionByIDHandler(txID string) (interface{}, *utils.HandlerError) {
	if bytes, err := hex.DecodeString(txID); err != nil || len(bytes) != daghash.TxIDSize {
		return nil, utils.NewHandlerError(http.StatusUnprocessableEntity,
			fmt.Sprintf("The given txid is not a hex-encoded %d-byte hash.", daghash.TxIDSize))
	}

	db, err := database.DB()
	if err != nil {
		return nil, utils.NewInternalServerHandlerError(err.Error())
	}

	tx := &models.Transaction{}
	query := db.Where(&models.Transaction{TransactionID: txID})
	dbResult := addTxPreloadedFields(query).First(&tx)
	if dbResult.RecordNotFound() && len(dbResult.GetErrors()) == 1 {
		return nil, utils.NewHandlerError(http.StatusNotFound, "No transaction with the given txid was found.")
	}
	if len(dbResult.GetErrors()) > 0 {
		return nil, utils.NewHandlerErrorFromDBErrors("Some errors where encountered when loading transaction from the database:", dbResult.GetErrors())
	}
	return convertTxModelToTxResponse(tx), nil
}

// GetTransactionByHashHandler returns a transaction by a given transaction hash.
func GetTransactionByHashHandler(txHash string) (interface{}, *utils.HandlerError) {
	if bytes, err := hex.DecodeString(txHash); err != nil || len(bytes) != daghash.HashSize {
		return nil, utils.NewHandlerError(http.StatusUnprocessableEntity,
			fmt.Sprintf("The given txhash is not a hex-encoded %d-byte hash.", daghash.HashSize))
	}

	db, err := database.DB()
	if err != nil {
		return nil, utils.NewInternalServerHandlerError(err.Error())
	}

	tx := &models.Transaction{}
	query := db.Where(&models.Transaction{TransactionHash: txHash})
	dbResult := addTxPreloadedFields(query).First(&tx)
	if dbResult.RecordNotFound() && len(dbResult.GetErrors()) == 1 {
		return nil, utils.NewHandlerError(http.StatusNotFound, "No transaction with the given txhash was found.")
	}
	if len(dbResult.GetErrors()) > 0 {
		return nil, utils.NewHandlerErrorFromDBErrors("Some errors where encountered when loading transaction from the database:", dbResult.GetErrors())
	}
	return convertTxModelToTxResponse(tx), nil
}

// GetTransactionsByAddressHandler searches for all transactions
// where the given address is either an input or an output.
func GetTransactionsByAddressHandler(address string, skip uint64, limit uint64) (interface{}, *utils.HandlerError) {
	if limit > maximumGetTransactionsLimit {
		return nil, utils.NewHandlerError(http.StatusUnprocessableEntity,
			fmt.Sprintf("The maximum allowed value for the limit is %d", maximumGetTransactionsLimit))
	}

	db, err := database.DB()
	if err != nil {
		return nil, utils.NewInternalServerHandlerError(err.Error())
	}

	txs := []*models.Transaction{}
	query := db.
		Joins("LEFT JOIN `transaction_outputs` ON `transaction_outputs`.`transaction_id` = `transactions`.`id`").
		Joins("LEFT JOIN `addresses` AS `out_addresses` ON `out_addresses`.`id` = `transaction_outputs`.`address_id`").
		Joins("LEFT JOIN `transaction_inputs` ON `transaction_inputs`.`transaction_id` = `transactions`.`id`").
		Joins("LEFT JOIN `transaction_outputs` AS `inputs_outs` ON `inputs_outs`.`id` = `transaction_inputs`.`transaction_output_id`").
		Joins("LEFT JOIN `addresses` AS `in_addresses` ON `in_addresses`.`id` = `inputs_outs`.`address_id`").
		Where("`out_addresses`.`address` = ?", address).
		Or("`in_addresses`.`address` = ?", address).
		Limit(limit).
		Offset(skip).
		Order("`transactions`.`id` ASC")
	dbErrors := addTxPreloadedFields(query).Find(&txs).GetErrors()
	if len(dbErrors) > 0 {
		return nil, utils.NewHandlerErrorFromDBErrors("Some errors where encountered when loading transactions from the database:", dbErrors)
	}
	txResponses := make([]*transactionResponse, len(txs))
	for i, tx := range txs {
		txResponses[i] = convertTxModelToTxResponse(tx)
	}
	return txResponses, nil
}

// GetUTXOsByAddressHandler searches for all UTXOs that belong to a certain address.
func GetUTXOsByAddressHandler(address string) (interface{}, *utils.HandlerError) {
	db, err := database.DB()
	if err != nil {
		return nil, utils.NewInternalServerHandlerError(err.Error())
	}

	var transactionOutputs []*models.TransactionOutput
	dbErrors := db.
		Joins("LEFT JOIN `addresses` ON `addresses`.`id` = `transaction_outputs`.`address_id`").
		Where("`addresses`.`address` = ? AND `transaction_outputs`.`is_spent` = 0", address).
		Preload("Transaction.AcceptingBlock").
		Find(&transactionOutputs).GetErrors()
	if len(dbErrors) > 0 {
		return nil, utils.NewHandlerErrorFromDBErrors("Some errors where encountered when loading UTXOs from the database:", dbErrors)
	}

	UTXOsResponses := make([]*transactionOutputResponse, len(transactionOutputs))
	for i, transactionOutput := range transactionOutputs {
		UTXOsResponses[i] = &transactionOutputResponse{
			Value:                   transactionOutput.Value,
			ScriptPubKey:            hex.EncodeToString(transactionOutput.ScriptPubKey),
			AcceptingBlockHash:      transactionOutput.Transaction.AcceptingBlock.BlockHash,
			AcceptingBlockBlueScore: transactionOutput.Transaction.AcceptingBlock.BlueScore,
		}
	}
	return UTXOsResponses, nil
}

func addTxPreloadedFields(query *gorm.DB) *gorm.DB {
	return query.Preload("AcceptingBlock").
		Preload("Subnetwork").
		Preload("TransactionOutputs").
		Preload("TransactionOutputs.Address").
		Preload("TransactionInputs.TransactionOutput.Transaction").
		Preload("TransactionInputs.TransactionOutput.Address")
}
