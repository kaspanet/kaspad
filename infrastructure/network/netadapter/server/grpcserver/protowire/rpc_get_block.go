package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
)

func (x *KaspadMessage_GetBlockRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.GetBlockRequestMessage{
		Hash:                          x.GetBlockRequest.Hash,
		SubnetworkID:                  x.GetBlockRequest.SubnetworkId,
		IncludeBlockHex:               x.GetBlockRequest.IncludeBlockHex,
		IncludeBlockVerboseData:       x.GetBlockRequest.IncludeBlockVerboseData,
		IncludeTransactionVerboseData: x.GetBlockRequest.IncludeTransactionVerboseData,
	}, nil
}

func (x *KaspadMessage_GetBlockRequest) fromAppMessage(message *appmessage.GetBlockRequestMessage) error {
	x.GetBlockRequest = &GetBlockRequestMessage{
		Hash:                          message.Hash,
		SubnetworkId:                  message.SubnetworkID,
		IncludeBlockHex:               message.IncludeBlockHex,
		IncludeBlockVerboseData:       message.IncludeBlockVerboseData,
		IncludeTransactionVerboseData: message.IncludeTransactionVerboseData,
	}
	return nil
}

func (x *KaspadMessage_GetBlockResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.GetBlockResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.GetBlockResponse.Error.Message}
	}
	var blockVerboseData *appmessage.BlockVerboseData
	if x.GetBlockResponse.BlockVerboseData != nil {
		transactionVerboseData := make([]*appmessage.TransactionVerboseData, len(x.GetBlockResponse.BlockVerboseData.TransactionVerboseData))
		for i, data := range x.GetBlockResponse.BlockVerboseData.TransactionVerboseData {
			vin := make([]*appmessage.Vin, len(data.Vin))
			for j, item := range data.Vin {
				scriptSig := &appmessage.ScriptSig{
					Asm: item.ScriptSig.Asm,
					Hex: item.ScriptSig.Hex,
				}
				vin[j] = &appmessage.Vin{
					TxID:      item.TxId,
					Vout:      item.Vout,
					ScriptSig: scriptSig,
					Sequence:  item.Sequence,
				}
			}
			vout := make([]*appmessage.Vout, len(data.Vout))
			for j, item := range data.Vout {
				scriptPubKey := &appmessage.ScriptPubKeyResult{
					Asm:     item.ScriptPubKey.Asm,
					Hex:     item.ScriptPubKey.Hex,
					Type:    item.ScriptPubKey.Type,
					Address: item.ScriptPubKey.Address,
				}
				vout[j] = &appmessage.Vout{
					Value:        item.Value,
					N:            item.N,
					ScriptPubKey: scriptPubKey,
				}
			}
			transactionVerboseData[i] = &appmessage.TransactionVerboseData{
				Hex:          data.Hex,
				TxID:         data.TxId,
				Hash:         data.Hash,
				Size:         data.Size,
				Version:      data.Version,
				LockTime:     data.LockTime,
				SubnetworkID: data.SubnetworkId,
				Gas:          data.Gas,
				PayloadHash:  data.PayloadHash,
				Payload:      data.Payload,
				Vin:          vin,
				Vout:         vout,
				BlockHash:    data.BlockHash,
				AcceptedBy:   data.AcceptedBy,
				IsInMempool:  data.IsInMempool,
				Time:         data.Time,
				BlockTime:    data.BlockTime,
			}
		}
		blockVerboseData = &appmessage.BlockVerboseData{
			Hash:                   x.GetBlockResponse.BlockVerboseData.Hash,
			Confirmations:          x.GetBlockResponse.BlockVerboseData.Confirmations,
			Size:                   x.GetBlockResponse.BlockVerboseData.Size,
			BlueScore:              x.GetBlockResponse.BlockVerboseData.BlueScore,
			IsChainBlock:           x.GetBlockResponse.BlockVerboseData.IsChainBlock,
			Version:                x.GetBlockResponse.BlockVerboseData.Version,
			VersionHex:             x.GetBlockResponse.BlockVerboseData.VersionHex,
			HashMerkleRoot:         x.GetBlockResponse.BlockVerboseData.HashMerkleRoot,
			AcceptedIDMerkleRoot:   x.GetBlockResponse.BlockVerboseData.AcceptedIDMerkleRoot,
			UTXOCommitment:         x.GetBlockResponse.BlockVerboseData.UtxoCommitment,
			TxIDs:                  x.GetBlockResponse.BlockVerboseData.TransactionHex,
			TransactionVerboseData: transactionVerboseData,
			Time:                   x.GetBlockResponse.BlockVerboseData.Time,
			Nonce:                  x.GetBlockResponse.BlockVerboseData.Nonce,
			Bits:                   x.GetBlockResponse.BlockVerboseData.Bits,
			Difficulty:             x.GetBlockResponse.BlockVerboseData.Difficulty,
			ParentHashes:           x.GetBlockResponse.BlockVerboseData.ParentHashes,
			SelectedParentHash:     x.GetBlockResponse.BlockVerboseData.SelectedParentHash,
			ChildHashes:            x.GetBlockResponse.BlockVerboseData.ChildHashes,
			AcceptedBlockHashes:    x.GetBlockResponse.BlockVerboseData.AcceptedBlockHashes,
		}
	}
	return &appmessage.GetBlockResponseMessage{
		BlockHex:         x.GetBlockResponse.BlockHex,
		BlockVerboseData: blockVerboseData,
		Error:            err,
	}, nil
}

func (x *KaspadMessage_GetBlockResponse) fromAppMessage(message *appmessage.GetBlockResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	var blockVerboseData *BlockVerboseData
	if x.GetBlockResponse.BlockVerboseData != nil {
		transactionVerboseData := make([]*TransactionVerboseData, len(message.BlockVerboseData.TransactionVerboseData))
		for i, data := range message.BlockVerboseData.TransactionVerboseData {
			vin := make([]*Vin, len(data.Vin))
			for j, item := range data.Vin {
				scriptSig := &ScriptSig{
					Asm: item.ScriptSig.Asm,
					Hex: item.ScriptSig.Hex,
				}
				vin[j] = &Vin{
					TxId:      item.TxID,
					Vout:      item.Vout,
					ScriptSig: scriptSig,
					Sequence:  item.Sequence,
				}
			}
			vout := make([]*Vout, len(data.Vout))
			for j, item := range data.Vout {
				scriptPubKey := &ScriptPubKeyResult{
					Asm:     item.ScriptPubKey.Asm,
					Hex:     item.ScriptPubKey.Hex,
					Type:    item.ScriptPubKey.Type,
					Address: item.ScriptPubKey.Address,
				}
				vout[j] = &Vout{
					Value:        item.Value,
					N:            item.N,
					ScriptPubKey: scriptPubKey,
				}
			}
			transactionVerboseData[i] = &TransactionVerboseData{
				Hex:          data.Hex,
				TxId:         data.TxID,
				Hash:         data.Hash,
				Size:         data.Size,
				Version:      data.Version,
				LockTime:     data.LockTime,
				SubnetworkId: data.SubnetworkID,
				Gas:          data.Gas,
				PayloadHash:  data.PayloadHash,
				Payload:      data.Payload,
				Vin:          vin,
				Vout:         vout,
				BlockHash:    data.BlockHash,
				AcceptedBy:   data.AcceptedBy,
				IsInMempool:  data.IsInMempool,
				Time:         data.Time,
				BlockTime:    data.BlockTime,
			}
		}
		blockVerboseData = &BlockVerboseData{
			Hash:                   message.BlockVerboseData.Hash,
			Confirmations:          message.BlockVerboseData.Confirmations,
			Size:                   message.BlockVerboseData.Size,
			BlueScore:              message.BlockVerboseData.BlueScore,
			IsChainBlock:           message.BlockVerboseData.IsChainBlock,
			Version:                message.BlockVerboseData.Version,
			VersionHex:             message.BlockVerboseData.VersionHex,
			HashMerkleRoot:         message.BlockVerboseData.HashMerkleRoot,
			AcceptedIDMerkleRoot:   message.BlockVerboseData.AcceptedIDMerkleRoot,
			UtxoCommitment:         message.BlockVerboseData.UTXOCommitment,
			TransactionHex:         message.BlockVerboseData.TxIDs,
			TransactionVerboseData: transactionVerboseData,
			Time:                   message.BlockVerboseData.Time,
			Nonce:                  message.BlockVerboseData.Nonce,
			Bits:                   message.BlockVerboseData.Bits,
			Difficulty:             message.BlockVerboseData.Difficulty,
			ParentHashes:           message.BlockVerboseData.ParentHashes,
			SelectedParentHash:     message.BlockVerboseData.SelectedParentHash,
			ChildHashes:            message.BlockVerboseData.ChildHashes,
			AcceptedBlockHashes:    message.BlockVerboseData.AcceptedBlockHashes,
		}
	}
	x.GetBlockResponse = &GetBlockResponseMessage{
		BlockHex:         message.BlockHex,
		BlockVerboseData: blockVerboseData,
		Error:            err,
	}
	return nil
}
