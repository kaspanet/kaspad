package controllers

import (
	"encoding/hex"
	"github.com/daglabs/btcd/apiserver/apimodels"
	"github.com/daglabs/btcd/apiserver/dbmodels"
	"github.com/daglabs/btcd/btcjson"
)

func convertTxDBModelToTxResponse(tx *dbmodels.Transaction) *apimodels.TransactionResponse {
	var acceptingBlockHash *string
	var acceptingBlockBlueScore *uint64

	if tx.AcceptingBlock != nil {
		acceptingBlockHash = &tx.AcceptingBlock.BlockHash
		acceptingBlockBlueScore = &tx.AcceptingBlock.BlueScore
	}

	txRes := &apimodels.TransactionResponse{
		TransactionHash:         tx.TransactionHash,
		TransactionID:           tx.TransactionID,
		AcceptingBlockHash:      acceptingBlockHash,
		AcceptingBlockBlueScore: acceptingBlockBlueScore,
		SubnetworkID:            tx.Subnetwork.SubnetworkID,
		LockTime:                tx.LockTime,
		Gas:                     tx.Gas,
		PayloadHash:             tx.PayloadHash,
		Payload:                 hex.EncodeToString(tx.Payload),
		Inputs:                  make([]*apimodels.TransactionInputResponse, len(tx.TransactionInputs)),
		Outputs:                 make([]*apimodels.TransactionOutputResponse, len(tx.TransactionOutputs)),
		Mass:                    tx.Mass,
	}
	for i, txOut := range tx.TransactionOutputs {
		txRes.Outputs[i] = &apimodels.TransactionOutputResponse{
			Value:        txOut.Value,
			ScriptPubKey: hex.EncodeToString(txOut.ScriptPubKey),
			Address:      txOut.Address.Address,
			Index:        txOut.Index,
		}
	}
	for i, txIn := range tx.TransactionInputs {
		txRes.Inputs[i] = &apimodels.TransactionInputResponse{
			PreviousTransactionID:          txIn.PreviousTransactionOutput.Transaction.TransactionID,
			PreviousTransactionOutputIndex: txIn.PreviousTransactionOutput.Index,
			SignatureScript:                hex.EncodeToString(txIn.SignatureScript),
			Sequence:                       txIn.Sequence,
			Address:                        txIn.PreviousTransactionOutput.Address.Address,
		}
	}
	return txRes
}

func convertBlockModelToBlockResponse(block *dbmodels.Block) *apimodels.BlockResponse {
	blockRes := &apimodels.BlockResponse{
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
