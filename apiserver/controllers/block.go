package controllers

import (
	"fmt"
	"github.com/daglabs/btcd/apiserver/database"
	"github.com/daglabs/btcd/apiserver/models"
	"github.com/daglabs/btcd/apiserver/utils"
	"github.com/daglabs/btcd/util/daghash"
	"net/http"
)

// GetBlockByHashHandler returns a block by a given hash.
func GetBlockByHashHandler(blockHash string) (interface{}, *utils.HandlerError) {
	if len(blockHash) != daghash.HashSize*2 {
		return nil, utils.NewHandlerError(http.StatusUnprocessableEntity, fmt.Sprintf("The given block hash is not a hex-encoded %d-byte hash.", daghash.HashSize))
	}
	if err := validateHex(blockHash); err != nil {
		return nil, utils.NewHandlerError(http.StatusUnprocessableEntity, fmt.Sprintf("Coulldn't parse the given block hash: %s", err))
	}
	block := &models.Block{}
	database.DB.Where(&models.Block{BlockHash: blockHash}).Preload("AcceptingBlock").First(block)
	if block.ID == 0 {
		return nil, utils.NewHandlerError(http.StatusNotFound, "No block with the given block hash was found.")
	}
	return convertBlockModelToBlockResponse(block), nil
}
