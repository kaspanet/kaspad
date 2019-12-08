package controllers

import (
	"encoding/hex"
	"github.com/daglabs/btcd/btcjson"
	"github.com/daglabs/btcd/kasparov/dbmodels"
	"github.com/daglabs/btcd/kasparov/server/models"
)

func convertTxDBModelToTxResponse(tx *dbmodels.Transaction) *models.TransactionResponse {
	txRes := &models.TransactionResponse{
		TransactionHash: tx.TransactionHash,
		TransactionID:   tx.TransactionID,
		SubnetworkID:    tx.Subnetwork.SubnetworkID,
		LockTime:        tx.LockTime,
		Gas:             tx.Gas,
		PayloadHash:     tx.PayloadHash,
		Payload:         hex.EncodeToString(tx.Payload),
		Inputs:          make([]*models.TransactionInputResponse, len(tx.TransactionInputs)),
		Outputs:         make([]*models.TransactionOutputResponse, len(tx.TransactionOutputs)),
		Mass:            tx.Mass,
	}
	if tx.AcceptingBlock != nil {
		txRes.AcceptingBlockHash = &tx.AcceptingBlock.BlockHash
		txRes.AcceptingBlockBlueScore = &tx.AcceptingBlock.BlueScore
	}
	for i, txOut := range tx.TransactionOutputs {
		txRes.Outputs[i] = &models.TransactionOutputResponse{
			Value:        txOut.Value,
			ScriptPubKey: hex.EncodeToString(txOut.ScriptPubKey),
			Address:      txOut.Address.Address,
			Index:        txOut.Index,
		}
	}
	for i, txIn := range tx.TransactionInputs {
		txRes.Inputs[i] = &models.TransactionInputResponse{
			PreviousTransactionID:          txIn.PreviousTransactionOutput.Transaction.TransactionID,
			PreviousTransactionOutputIndex: txIn.PreviousTransactionOutput.Index,
			SignatureScript:                hex.EncodeToString(txIn.SignatureScript),
			Sequence:                       txIn.Sequence,
			Address:                        txIn.PreviousTransactionOutput.Address.Address,
		}
	}
	return txRes
}

func convertBlockModelToBlockResponse(block *dbmodels.Block) *models.BlockResponse {
	blockRes := &models.BlockResponse{
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
