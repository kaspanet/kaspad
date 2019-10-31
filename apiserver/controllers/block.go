package controllers

import (
	"encoding/hex"
	"fmt"
	"github.com/daglabs/btcd/apiserver/apimodels"
	"github.com/daglabs/btcd/apiserver/dbmodels"
	"github.com/daglabs/btcd/httpserverutils"
	"net/http"

	"github.com/daglabs/btcd/apiserver/database"
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
func GetBlockByHashHandler(blockHash string) (interface{}, *httpserverutils.HandlerError) {
	if bytes, err := hex.DecodeString(blockHash); err != nil || len(bytes) != daghash.HashSize {
		return nil, httpserverutils.NewHandlerError(http.StatusUnprocessableEntity,
			fmt.Sprintf("The given block hash is not a hex-encoded %d-byte hash.", daghash.HashSize))
	}

	db, err := database.DB()
	if err != nil {
		return nil, httpserverutils.NewInternalServerHandlerError(err.Error())
	}

	block := &dbmodels.Block{}
	dbResult := db.Where(&dbmodels.Block{BlockHash: blockHash}).Preload("AcceptingBlock").First(block)
	dbErrors := dbResult.GetErrors()
	if httpserverutils.IsDBRecordNotFoundError(dbErrors) {
		return nil, httpserverutils.NewHandlerError(http.StatusNotFound, "No block with the given block hash was found.")
	}
	if httpserverutils.HasDBError(dbErrors) {
		return nil, httpserverutils.NewHandlerErrorFromDBErrors("Some errors were encountered when loading transactions from the database:", dbResult.GetErrors())
	}
	return convertBlockModelToBlockResponse(block), nil
}

// GetBlocksHandler searches for all blocks
func GetBlocksHandler(order string, skip uint64, limit uint64) (interface{}, *httpserverutils.HandlerError) {
	if limit > maxGetBlocksLimit {
		return nil, httpserverutils.NewHandlerError(http.StatusUnprocessableEntity, fmt.Sprintf("The maximum allowed value for the limit is %d", maxGetTransactionsLimit))
	}
	blocks := []*dbmodels.Block{}
	db, err := database.DB()
	if err != nil {
		return nil, httpserverutils.NewHandlerError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
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
		return nil, httpserverutils.NewHandlerError(http.StatusUnprocessableEntity, fmt.Sprintf("'%s' is not a valid order", order))
	}
	query.Find(&blocks)
	blockResponses := make([]*apimodels.BlockResponse, len(blocks))
	for i, block := range blocks {
		blockResponses[i] = convertBlockModelToBlockResponse(block)
	}
	return blockResponses, nil
}
