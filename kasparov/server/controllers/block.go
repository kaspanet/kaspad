package controllers

import (
	"encoding/hex"
	"github.com/daglabs/btcd/apiserver/database"
	"github.com/daglabs/btcd/apiserver/server/apimodels"
	"net/http"

	"github.com/daglabs/btcd/httpserverutils"
	"github.com/pkg/errors"

	"github.com/daglabs/btcd/util/daghash"
)

const (
	// OrderAscending is parameter that can be used
	// in a get list handler to get a list ordered
	// in an ascending order.
	OrderAscending = "asc"

	// OrderDescending is parameter that can be used
	// in a get list handler to get a list ordered
	// in an ascending order.
	OrderDescending = "desc"
)

const maxGetBlocksLimit = 100

// GetBlockByHashHandler returns a block by a given hash.
func GetBlockByHashHandler(blockHash string) (interface{}, error) {
	if bytes, err := hex.DecodeString(blockHash); err != nil || len(bytes) != daghash.HashSize {
		return nil, httpserverutils.NewHandlerError(http.StatusUnprocessableEntity,
			errors.Errorf("The given block hash is not a hex-encoded %d-byte hash.", daghash.HashSize))
	}

	db, err := database.DB()
	if err != nil {
		return nil, err
	}

	block := &database.Block{}
	dbResult := db.Where(&database.Block{BlockHash: blockHash}).Preload("AcceptingBlock").First(block)
	dbErrors := dbResult.GetErrors()
	if httpserverutils.IsDBRecordNotFoundError(dbErrors) {
		return nil, httpserverutils.NewHandlerError(http.StatusNotFound, errors.New("No block with the given block hash was found"))
	}
	if httpserverutils.HasDBError(dbErrors) {
		return nil, httpserverutils.NewErrorFromDBErrors("Some errors were encountered when loading transactions from the database:",
			dbResult.GetErrors())
	}
	return convertBlockModelToBlockResponse(block), nil
}

// GetBlocksHandler searches for all blocks
func GetBlocksHandler(order string, skip uint64, limit uint64) (interface{}, error) {
	if limit < 1 || limit > maxGetBlocksLimit {
		return nil, httpserverutils.NewHandlerError(http.StatusBadRequest,
			errors.Errorf("Limit higher than %d or lower than 1 was requested.", maxGetTransactionsLimit))
	}
	blocks := []*database.Block{}
	db, err := database.DB()
	if err != nil {
		return nil, err
	}
	query := db.
		Limit(limit).
		Offset(skip).
		Preload("AcceptingBlock")
	if order == OrderAscending {
		query = query.Order("`id` ASC")
	} else if order == OrderDescending {
		query = query.Order("`id` DESC")
	} else {
		return nil, httpserverutils.NewHandlerError(http.StatusUnprocessableEntity, errors.Errorf("'%s' is not a valid order", order))
	}
	query.Find(&blocks)
	blockResponses := make([]*apimodels.BlockResponse, len(blocks))
	for i, block := range blocks {
		blockResponses[i] = convertBlockModelToBlockResponse(block)
	}
	return blockResponses, nil
}

// GetAcceptedTransactionIDsByBlockHashHandler returns an array of transaction IDs for a given block hash
func GetAcceptedTransactionIDsByBlockHashHandler(blockHash string) ([]string, error) {
	db, err := database.DB()
	if err != nil {
		return nil, err
	}

	var transactions []database.Transaction
	dbResult := db.
		Joins("LEFT JOIN `blocks` ON `blocks`.`id` = `transactions`.`accepting_block_id`").
		Where("`blocks`.`block_hash` = ?", blockHash).
		Find(&transactions)

	dbErrors := dbResult.GetErrors()
	if httpserverutils.HasDBError(dbErrors) {
		return nil, httpserverutils.NewErrorFromDBErrors("Failed to find transactions: ", dbErrors)
	}

	result := make([]string, len(transactions))
	for _, transaction := range transactions {
		result = append(result, transaction.TransactionID)
	}

	return result, nil
}
