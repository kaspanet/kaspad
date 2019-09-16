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

// GetBlockByHashHandler returns a block by a given hash.
func GetBlockByHashHandler(blockHash string) (interface{}, *utils.HandlerError) {
	if bytes, err := hex.DecodeString(blockHash); err != nil || len(bytes) != daghash.HashSize {
		return nil, utils.NewHandlerError(http.StatusUnprocessableEntity,
			fmt.Sprintf("The given block hash is not a hex-encoded %d-byte hash.", daghash.HashSize))
	}

	db, err := database.DB()
	if err != nil {
		return nil, utils.NewInternalServerHandlerError(err.Error())
	}

	block := &models.Block{}
	dbErrors := db.Where(&models.Block{BlockHash: blockHash}).Preload("AcceptingBlock").First(block).GetErrors()
	if len(dbErrors) > 0 {
		return nil, utils.NewHandleErrorFromDBErrors("Some errors where encountered when loading transactions from the database:", dbErrors)
	}
	if block.ID == 0 {
		return nil, utils.NewHandlerError(http.StatusNotFound, "No block with the given block hash was found.")
	}
	return convertBlockModelToBlockResponse(block), nil
}
