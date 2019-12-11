package controllers

import (
	"encoding/hex"

	"github.com/kaspanet/kaspad/btcjson"
	"github.com/kaspanet/kaspad/kasparov/dbmodels"
	"github.com/kaspanet/kaspad/kasparov/kasparovd/apimodels"
)

func convertTxDBModelToTxResponse(tx *dbmodels.Transaction) *apimodels.TransactionResponse {
	txRes := &apimodels.TransactionResponse{
		TransactionHash: tx.TransactionHash,
		TransactionID:   tx.TransactionID,
		SubnetworkID:    tx.Subnetwork.SubnetworkID,
		LockTime:        tx.LockTime,
		Gas:             tx.Gas,
		PayloadHash:     tx.PayloadHash,
		Payload:         hex.EncodeToString(tx.Payload),
		Inputs:          make([]*apimodels.TransactionInputResponse, len(tx.TransactionInputs)),
		Outputs:         make([]*apimodels.TransactionOutputResponse, len(tx.TransactionOutputs)),
		Mass:            tx.Mass,
	}
	if tx.AcceptingBlock != nil {
		txRes.AcceptingBlockHash = &tx.AcceptingBlock.BlockHash
		txRes.AcceptingBlockBlueScore = &tx.AcceptingBlock.BlueScore
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
