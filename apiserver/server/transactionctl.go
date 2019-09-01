package server

import (
	"encoding/hex"
	"fmt"
	"github.com/daglabs/btcd/apiserver/database"
	"github.com/daglabs/btcd/apiserver/models"
	"github.com/daglabs/btcd/util/daghash"
	"net/http"
)

type transactionResponse struct {
	TransactionHash         string
	TransactionID           string
	AcceptingBlockHash      string
	AcceptingBlockBlueScore uint64
	SubnetworkID            string
	LockTime                uint64
	Gas                     uint64
	PayloadHash             string
	Payload                 string
	Inputs                  []*transactionInputResponse
	Outputs                 []*transactionOutputResponse
	Mass                    uint64
}

type transactionOutputResponse struct {
	TransactionID string
	Value         uint64
	PkScript      string
	Address       string
}

type transactionInputResponse struct {
	TransactionID                  string
	PreviousTransactionID          string
	PreviousTransactionOutputIndex uint32
	ScriptSig                      string
	Sequence                       uint64
}

func getTransactionByIDHandler(vars map[string]string, ctx *apiServerContext) (interface{}, *handlerError) {
	txID := vars["txID"]
	if len(txID) != daghash.TxIDSize*2 {
		return nil, newHandleError(http.StatusUnprocessableEntity, fmt.Sprintf("The given txid is not a hex-encoded %d-byte hash.", daghash.TxIDSize))
	}
	tx := &models.Transaction{}
	database.DB.Where("transaction_id = ?", txID).
		Preload("AcceptingBlock").
		Preload("Subnetwork").
		Preload("TransactionOutputs").
		Preload("TransactionInputs").
		Preload("TransactionInputs.TransactionOutput").
		Preload("TransactionInputs.TransactionOutput.Transaction").
		First(&tx)
	if tx.ID == 0 {
		return nil, newHandleError(http.StatusNotFound, "No transaction with the given txid was found.")
	}
	return convertTxModelToTxResponse(tx), nil
}

func convertTxModelToTxResponse(tx *models.Transaction) *transactionResponse {
	txRes := &transactionResponse{
		TransactionHash:         tx.TransactionHash,
		TransactionID:           tx.TransactionID,
		AcceptingBlockHash:      tx.AcceptingBlock.BlockHash,
		AcceptingBlockBlueScore: tx.AcceptingBlock.BlueScore,
		SubnetworkID:            tx.Subnetwork.SubnetworkID,
		LockTime:                tx.LockTime,
		Gas:                     tx.Gas,
		PayloadHash:             tx.PayloadHash,
		Payload:                 hex.EncodeToString(tx.Payload),
		Inputs:                  make([]*transactionInputResponse, len(tx.TransactionOutputs)),
		Outputs:                 make([]*transactionOutputResponse, len(tx.TransactionInputs)),
		Mass:                    tx.Mass,
	}
	for i, txOut := range tx.TransactionOutputs {
		txRes.Outputs[i] = &transactionOutputResponse{
			TransactionID: tx.TransactionID,
			Value:         txOut.Value,
			PkScript:      hex.EncodeToString(txOut.PkScript),
			Address:       "", // TODO: Fill it when there's an addrindex in the DB.
		}
	}
	for i, txIn := range tx.TransactionInputs {
		txRes.Inputs[i] = &transactionInputResponse{
			TransactionID:                  tx.TransactionID,
			PreviousTransactionID:          txIn.TransactionOutput.Transaction.TransactionID,
			PreviousTransactionOutputIndex: txIn.TransactionOutput.Index,
			ScriptSig:                      hex.EncodeToString(txIn.SignatureScript),
			Sequence:                       txIn.Sequence,
		}
	}
	return txRes
}
