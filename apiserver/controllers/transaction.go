package controllers

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/daglabs/btcd/apiserver/apimodels"
	"github.com/daglabs/btcd/apiserver/dbmodels"
	"github.com/daglabs/btcd/blockdag"
	"github.com/daglabs/btcd/httpserverutils"
	"github.com/daglabs/btcd/util/subnetworkid"
	"net/http"

	"github.com/daglabs/btcd/apiserver/database"
	"github.com/daglabs/btcd/apiserver/jsonrpc"
	"github.com/daglabs/btcd/btcjson"
	"github.com/daglabs/btcd/util/daghash"
	"github.com/daglabs/btcd/wire"
	"github.com/jinzhu/gorm"
)

const maxGetTransactionsLimit = 1000

// GetTransactionByIDHandler returns a transaction by a given transaction ID.
func GetTransactionByIDHandler(txID string) (interface{}, *httpserverutils.HandlerError) {
	if bytes, err := hex.DecodeString(txID); err != nil || len(bytes) != daghash.TxIDSize {
		return nil, httpserverutils.NewHandlerError(http.StatusUnprocessableEntity,
			fmt.Sprintf("The given txid is not a hex-encoded %d-byte hash.", daghash.TxIDSize))
	}

	db, err := database.DB()
	if err != nil {
		return nil, httpserverutils.NewInternalServerHandlerError(err.Error())
	}

	tx := &dbmodels.Transaction{}
	query := db.Where(&dbmodels.Transaction{TransactionID: txID})
	dbResult := addTxPreloadedFields(query).First(&tx)
	dbErrors := dbResult.GetErrors()
	if httpserverutils.IsDBRecordNotFoundError(dbErrors) {
		return nil, httpserverutils.NewHandlerError(http.StatusNotFound, "No transaction with the given txid was found.")
	}
	if httpserverutils.HasDBError(dbErrors) {
		return nil, httpserverutils.NewHandlerErrorFromDBErrors("Some errors were encountered when loading transaction from the database:", dbErrors)
	}
	return convertTxDBModelToTxResponse(tx), nil
}

// GetTransactionByHashHandler returns a transaction by a given transaction hash.
func GetTransactionByHashHandler(txHash string) (interface{}, *httpserverutils.HandlerError) {
	if bytes, err := hex.DecodeString(txHash); err != nil || len(bytes) != daghash.HashSize {
		return nil, httpserverutils.NewHandlerError(http.StatusUnprocessableEntity,
			fmt.Sprintf("The given txhash is not a hex-encoded %d-byte hash.", daghash.HashSize))
	}

	db, err := database.DB()
	if err != nil {
		return nil, httpserverutils.NewHandlerError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	tx := &dbmodels.Transaction{}
	query := db.Where(&dbmodels.Transaction{TransactionHash: txHash})
	dbResult := addTxPreloadedFields(query).First(&tx)
	dbErrors := dbResult.GetErrors()
	if httpserverutils.IsDBRecordNotFoundError(dbErrors) {
		return nil, httpserverutils.NewHandlerError(http.StatusNotFound, "No transaction with the given txhash was found.")
	}
	if httpserverutils.HasDBError(dbErrors) {
		return nil, httpserverutils.NewHandlerErrorFromDBErrors("Some errors were encountered when loading transaction from the database:", dbErrors)
	}
	return convertTxDBModelToTxResponse(tx), nil
}

// GetTransactionsByAddressHandler searches for all transactions
// where the given address is either an input or an output.
func GetTransactionsByAddressHandler(address string, skip uint64, limit uint64) (interface{}, *httpserverutils.HandlerError) {
	if limit > maxGetTransactionsLimit {
		return nil, httpserverutils.NewHandlerError(http.StatusUnprocessableEntity,
			fmt.Sprintf("The maximum allowed value for the limit is %d", maxGetTransactionsLimit))
	}

	db, err := database.DB()
	if err != nil {
		return nil, httpserverutils.NewHandlerError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	txs := []*dbmodels.Transaction{}
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
	dbResult := addTxPreloadedFields(query).Find(&txs)
	dbErrors := dbResult.GetErrors()
	if httpserverutils.HasDBError(dbErrors) {
		return nil, httpserverutils.NewHandlerErrorFromDBErrors("Some errors were encountered when loading transactions from the database:", dbErrors)
	}
	txResponses := make([]*apimodels.TransactionResponse, len(txs))
	for i, tx := range txs {
		txResponses[i] = convertTxDBModelToTxResponse(tx)
	}
	return txResponses, nil
}

func fetchSelectedTipBlueScore() (uint64, *httpserverutils.HandlerError) {
	db, err := database.DB()
	if err != nil {
		return 0, httpserverutils.NewInternalServerHandlerError(err.Error())
	}
	block := &dbmodels.Block{}
	dbResult := db.Order("blue_score DESC").
		Where(&dbmodels.Block{IsChainBlock: true}).
		First(block)
	dbErrors := dbResult.GetErrors()
	if httpserverutils.HasDBError(dbErrors) {
		return 0, httpserverutils.NewHandlerErrorFromDBErrors("Some errors were encountered when loading transactions from the database:", dbErrors)
	}
	return block.BlueScore, nil
}

// GetUTXOsByAddressHandler searches for all UTXOs that belong to a certain address.
func GetUTXOsByAddressHandler(address string) (interface{}, *httpserverutils.HandlerError) {
	db, err := database.DB()
	if err != nil {
		return nil, httpserverutils.NewInternalServerHandlerError(err.Error())
	}

	var transactionOutputs []*dbmodels.TransactionOutput
	dbErrors := db.
		Joins("LEFT JOIN `addresses` ON `addresses`.`id` = `transaction_outputs`.`address_id`").
		Where("`addresses`.`address` = ? AND `transaction_outputs`.`is_spent` = 0", address).
		Preload("Transaction.AcceptingBlock").
		Preload("Transaction.Subnetwork").
		Find(&transactionOutputs).GetErrors()
	if len(dbErrors) > 0 {
		return nil, httpserverutils.NewHandlerErrorFromDBErrors("Some errors were encountered when loading UTXOs from the database:", dbErrors)
	}

	selectedTipBlueScore, hErr := fetchSelectedTipBlueScore()
	if hErr != nil {
		return nil, hErr
	}

	UTXOsResponses := make([]*apimodels.TransactionOutputResponse, len(transactionOutputs))
	for i, transactionOutput := range transactionOutputs {
		subnetworkID := &subnetworkid.SubnetworkID{}
		err := subnetworkid.Decode(subnetworkID, transactionOutput.Transaction.Subnetwork.SubnetworkID)
		if err != nil {
			return nil, httpserverutils.NewInternalServerHandlerError(fmt.Sprintf("Couldn't decode subnetwork id %s: %s", transactionOutput.Transaction.Subnetwork.SubnetworkID, err))
		}
		var acceptingBlockHash *string
		var confirmations uint64
		acceptingBlockBlueScore := blockdag.UnacceptedBlueScore
		if transactionOutput.Transaction.AcceptingBlock != nil {
			acceptingBlockHash = btcjson.String(transactionOutput.Transaction.AcceptingBlock.BlockHash)
			acceptingBlockBlueScore = transactionOutput.Transaction.AcceptingBlock.BlueScore
			confirmations = selectedTipBlueScore - acceptingBlockBlueScore
		}
		UTXOsResponses[i] = &apimodels.TransactionOutputResponse{
			TransactionID:           transactionOutput.Transaction.TransactionID,
			Value:                   transactionOutput.Value,
			ScriptPubKey:            hex.EncodeToString(transactionOutput.ScriptPubKey),
			AcceptingBlockHash:      acceptingBlockHash,
			AcceptingBlockBlueScore: acceptingBlockBlueScore,
			Index:                   transactionOutput.Index,
			IsCoinbase:              btcjson.Bool(subnetworkID.IsEqual(subnetworkid.SubnetworkIDCoinbase)),
			Confirmations:           btcjson.Uint64(confirmations),
		}
	}
	return UTXOsResponses, nil
}

func addTxPreloadedFields(query *gorm.DB) *gorm.DB {
	return query.Preload("AcceptingBlock").
		Preload("Subnetwork").
		Preload("TransactionOutputs").
		Preload("TransactionOutputs.Address").
		Preload("TransactionInputs.PreviousTransactionOutput.Transaction").
		Preload("TransactionInputs.PreviousTransactionOutput.Address")
}

// PostTransaction forwards a raw transaction to the JSON-RPC API server
func PostTransaction(requestBody []byte) *httpserverutils.HandlerError {
	client, err := jsonrpc.GetClient()
	if err != nil {
		return httpserverutils.NewInternalServerHandlerError(err.Error())
	}

	rawTx := &apimodels.RawTransaction{}
	err = json.Unmarshal(requestBody, rawTx)
	if err != nil {
		return httpserverutils.NewHandlerErrorWithCustomClientMessage(http.StatusUnprocessableEntity,
			fmt.Sprintf("Error unmarshalling request body: %s", err),
			"The request body is not json-formatted")
	}

	txBytes, err := hex.DecodeString(rawTx.RawTransaction)
	if err != nil {
		return httpserverutils.NewHandlerErrorWithCustomClientMessage(http.StatusUnprocessableEntity,
			fmt.Sprintf("Error decoding hex raw transaction: %s", err),
			"The raw transaction is not a hex-encoded transaction")
	}

	txReader := bytes.NewReader(txBytes)
	tx := &wire.MsgTx{}
	err = tx.BtcDecode(txReader, 0)
	if err != nil {
		return httpserverutils.NewHandlerErrorWithCustomClientMessage(http.StatusUnprocessableEntity,
			fmt.Sprintf("Error decoding raw transaction: %s", err),
			"Error decoding raw transaction")
	}

	_, err = client.SendRawTransaction(tx, true)
	if err != nil {
		if rpcErr, ok := err.(*btcjson.RPCError); ok && rpcErr.Code == btcjson.ErrRPCVerify {
			return httpserverutils.NewHandlerError(http.StatusInternalServerError, rpcErr.Message)
		}
		return httpserverutils.NewHandlerError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	return nil
}
