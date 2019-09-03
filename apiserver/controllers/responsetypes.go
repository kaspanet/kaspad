package controllers

import (
	"encoding/hex"
	"github.com/daglabs/btcd/apiserver/models"
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
