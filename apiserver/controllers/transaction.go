package controllers

import (
	"encoding/hex"
	"fmt"
	"github.com/daglabs/btcd/apiserver/database"
	"github.com/daglabs/btcd/apiserver/models"
	"github.com/daglabs/btcd/apiserver/utils"
	"github.com/daglabs/btcd/util/daghash"
	"net/http"
)

type transactionResponse struct {
	TransactionHash         string                       `json:"transactionHash"`
	TransactionID           string                       `json:"transactionId"`
	AcceptingBlockHash      string                       `json:"acceptingBlockHash,omitempty"`
	AcceptingBlockBlueScore uint64                       `json:"acceptingBlockBlueScore,omitempty"`
	SubnetworkID            string                       `json:"subnetworkId"`
	LockTime                uint64                       `json:"lockTime"`
	Gas                     uint64                       `json:"gas,omitempty"`
	PayloadHash             string                       `json:"payloadHash,omitempty"`
	Payload                 string                       `json:"payload,omitempty"`
	Inputs                  []*transactionInputResponse  `json:"inputs"`
	Outputs                 []*transactionOutputResponse `json:"outputs"`
	Mass                    uint64                       `json:"mass"`
}

type transactionOutputResponse struct {
	TransactionID string `json:"transactionId,omitempty"`
	Value         uint64 `json:"value"`
	PkScript      string `json:"pkScript"`
	Address       string `json:"address"`
}

type transactionInputResponse struct {
	TransactionID                  string `json:"transactionId,omitempty"`
	PreviousTransactionID          string `json:"previousTransactionId"`
	PreviousTransactionOutputIndex uint32 `json:"previousTransactionOutputIndex"`
	SignatureScript                string `json:"signatureScript"`
	Sequence                       uint64 `json:"sequence"`
}

func GetTransactionByIDHandler(txID string) (interface{}, *utils.HandlerError) {
	if len(txID) != daghash.TxIDSize*2 {
		return nil, utils.NewHandlerError(http.StatusUnprocessableEntity, fmt.Sprintf("The given txid is not a hex-encoded %d-byte hash.", daghash.TxIDSize))
	}
	tx := &models.Transaction{}
	database.DB.Where("transaction_id = ?", txID).
		Preload("AcceptingBlock").
		Preload("Subnetwork").
		Preload("TransactionOutputs").
		Preload("TransactionInputs.TransactionOutput.Transaction").
		First(&tx)
	if tx.ID == 0 {
		return nil, utils.NewHandlerError(http.StatusNotFound, "No transaction with the given txid was found.")
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
			Value:    txOut.Value,
			PkScript: hex.EncodeToString(txOut.PkScript),
			Address:  "", // TODO: Fill it when there's an addrindex in the DB.
		}
	}
	for i, txIn := range tx.TransactionInputs {
		txRes.Inputs[i] = &transactionInputResponse{
			PreviousTransactionID:          txIn.TransactionOutput.Transaction.TransactionID,
			PreviousTransactionOutputIndex: txIn.TransactionOutput.Index,
			SignatureScript:                hex.EncodeToString(txIn.SignatureScript),
			Sequence:                       txIn.Sequence,
		}
	}
	return txRes
}
