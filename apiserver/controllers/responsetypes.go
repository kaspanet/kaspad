package controllers

import (
	"encoding/hex"
	"github.com/daglabs/btcd/apiserver/models"
	"github.com/daglabs/btcd/btcjson"
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
	TransactionID           string `json:"transactionId,omitempty"`
	Value                   uint64 `json:"value"`
	ScriptPubKey            string `json:"scriptPubKey"`
	Address                 string `json:"address,omitempty"`
	AcceptingBlockHash      string `json:"acceptingBlockHash,omitempty"`
	AcceptingBlockBlueScore uint64 `json:"acceptingBlockBlueScore,omitempty"`
}

type transactionInputResponse struct {
	TransactionID                  string `json:"transactionId,omitempty"`
	PreviousTransactionID          string `json:"previousTransactionId"`
	PreviousTransactionOutputIndex uint32 `json:"previousTransactionOutputIndex"`
	SignatureScript                string `json:"signatureScript"`
	Sequence                       uint64 `json:"sequence"`
	Address                        string `json:"address"`
}

type blockResponse struct {
	BlockHash            string
	Version              int32
	HashMerkleRoot       string
	AcceptedIDMerkleRoot string
	UTXOCommitment       string
	Timestamp            uint64
	Bits                 uint32
	Nonce                uint64
	AcceptingBlockHash   *string
	BlueScore            uint64
	IsChainBlock         bool
	Mass                 uint64
}

type feeEstimateResponse struct {
	HighPriority, NormalPriority, LowPriority float64
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
		Inputs:                  make([]*transactionInputResponse, len(tx.TransactionInputs)),
		Outputs:                 make([]*transactionOutputResponse, len(tx.TransactionOutputs)),
		Mass:                    tx.Mass,
	}
	for i, txOut := range tx.TransactionOutputs {
		txRes.Outputs[i] = &transactionOutputResponse{
			Value:        txOut.Value,
			ScriptPubKey: hex.EncodeToString(txOut.ScriptPubKey),
			Address:      txOut.Address.Address,
		}
	}
	for i, txIn := range tx.TransactionInputs {
		txRes.Inputs[i] = &transactionInputResponse{
			PreviousTransactionID:          txIn.TransactionOutput.Transaction.TransactionID,
			PreviousTransactionOutputIndex: txIn.TransactionOutput.Index,
			SignatureScript:                hex.EncodeToString(txIn.SignatureScript),
			Sequence:                       txIn.Sequence,
			Address:                        txIn.TransactionOutput.Address.Address,
		}
	}
	return txRes
}

func convertBlockModelToBlockResponse(block *models.Block) *blockResponse {
	blockRes := &blockResponse{
		BlockHash:            block.BlockHash,
		Version:              block.Version,
		HashMerkleRoot:       block.HashMerkleRoot,
		AcceptedIDMerkleRoot: block.AcceptedIDMerkleRoot,
		UTXOCommitment:       block.UTXOCommitment,
		Timestamp:            uint64(block.Timestamp.Unix()),
		Bits:                 block.Bits,
		Nonce:                block.Nonce,
		BlueScore:            block.BlueScore,
		IsChainBlock:         block.IsChainBlock,
		Mass:                 block.Mass,
	}
	if block.AcceptingBlock != nil {
		blockRes.AcceptingBlockHash = btcjson.String(block.AcceptingBlock.BlockHash)
	}
	return blockRes
}
