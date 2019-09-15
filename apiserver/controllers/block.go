package controllers

import (
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/daglabs/btcd/apiserver/database"
	"github.com/daglabs/btcd/apiserver/models"
	"github.com/daglabs/btcd/apiserver/utils"
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

const maximumGetBlocksLimit = 100

// GetBlockByHashHandler returns a block by a given hash.
func GetBlockByHashHandler(blockHash string) (interface{}, *utils.HandlerError) {
	if bytes, err := hex.DecodeString(blockHash); err != nil || len(bytes) != daghash.HashSize {
		return nil, utils.NewHandlerError(http.StatusUnprocessableEntity,
			fmt.Sprintf("The given block hash is not a hex-encoded %d-byte hash.", daghash.HashSize))
	}

	db, err := database.DB()
	if err != nil {
		return nil, utils.NewHandlerError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	block := &models.Block{}
	db.Where(&models.Block{BlockHash: blockHash}).Preload("AcceptingBlock").First(block)
	if block.ID == 0 {
		return nil, utils.NewHandlerError(http.StatusNotFound, "No block with the given block hash was found.")
	}
	return convertBlockModelToBlockResponse(block), nil
}

// GetBlocksHandler searches for all blocks
func GetBlocksHandler(order string, skip uint64, limit uint64) (interface{}, *utils.HandlerError) {
	if limit > maximumGetBlocksLimit {
		return nil, utils.NewHandlerError(http.StatusUnprocessableEntity, fmt.Sprintf("The maximum allowed value for the limit is %d", maximumGetTransactionsLimit))
	}
	blocks := []*models.Block{}
	db, err := database.DB()
	if err != nil {
		return nil, utils.NewHandlerError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
	query := db.
		Limit(limit).
		Offset(skip).
		Preload("AcceptingBlock")
	if order == OrderAscending {
		query = query.Order("`id` ASC")
	} else {
		query = query.Order("`id` DESC")
	}
	query.Find(&blocks)
	blockResponses := make([]*blockResponse, len(blocks))
	for i, block := range blocks {
		blockResponses[i] = convertBlockModelToBlockResponse(block)
	}
	return blockResponses, nil
}
