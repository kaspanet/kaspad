package rpchandlers

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/subnetworkid"
	"strconv"
)

// HandleGetBlock handles the respectively named RPC command
func HandleGetBlock(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	getBlockRequest := request.(*appmessage.GetBlockRequestMessage)

	// Load the raw block bytes from the database.
	hash, err := daghash.NewHashFromStr(getBlockRequest.Hash)
	if err != nil {
		errorMessage := &appmessage.GetBlockResponseMessage{}
		errorMessage.Error = &appmessage.RPCError{
			Message: fmt.Sprintf("Hash could not be parsed: %s", err),
		}
		return errorMessage, nil
	}
	if context.DAG.IsKnownInvalid(hash) {
		errorMessage := &appmessage.GetBlockResponseMessage{}
		errorMessage.Error = &appmessage.RPCError{
			Message: fmt.Sprintf("Block %s is known to be invalid", hash),
		}
		return errorMessage, nil
	}
	if context.DAG.IsKnownOrphan(hash) {
		errorMessage := &appmessage.GetBlockResponseMessage{}
		errorMessage.Error = &appmessage.RPCError{
			Message: fmt.Sprintf("Block %s is an orphan", hash),
		}
		return errorMessage, nil
	}
	block, err := context.DAG.BlockByHash(hash)
	if err != nil {
		errorMessage := &appmessage.GetBlockResponseMessage{}
		errorMessage.Error = &appmessage.RPCError{
			Message: fmt.Sprintf("Block %s not found", hash),
		}
		return errorMessage, nil
	}

	blockBytes, err := block.Bytes()
	if err != nil {
		return nil, err
	}

	// Handle partial blocks
	if getBlockRequest.SubnetworkID != "" {
		requestSubnetworkID, err := subnetworkid.NewFromStr(getBlockRequest.SubnetworkID)
		if err != nil {
			errorMessage := &appmessage.GetBlockResponseMessage{}
			errorMessage.Error = &appmessage.RPCError{
				Message: fmt.Sprintf("SubnetworkID could not be parsed: %s", err),
			}
			return errorMessage, nil
		}
		nodeSubnetworkID := context.Config.SubnetworkID

		if requestSubnetworkID != nil {
			if nodeSubnetworkID != nil {
				if !nodeSubnetworkID.IsEqual(requestSubnetworkID) {
					errorMessage := &appmessage.GetBlockResponseMessage{}
					errorMessage.Error = &appmessage.RPCError{
						Message: fmt.Sprintf("subnetwork %s does not match this partial node",
							getBlockRequest.SubnetworkID),
					}
					return errorMessage, nil
				}
				// nothing to do - partial node stores partial blocks
			} else {
				// Deserialize the block.
				msgBlock := block.MsgBlock()
				msgBlock.ConvertToPartial(requestSubnetworkID)
				var b bytes.Buffer
				err := msgBlock.Serialize(bufio.NewWriter(&b))
				if err != nil {
					return nil, err
				}
				blockBytes = b.Bytes()
			}
		}
	}

	response := appmessage.NewGetBlockResponseMessage()

	if getBlockRequest.IncludeBlockHex {
		response.BlockHex = hex.EncodeToString(blockBytes)
	}
	if getBlockRequest.IncludeBlockVerboseData {
		blockVerboseData, err := buildBlockVerboseData(context, block, true)
		if err != nil {
			return nil, err
		}
		response.BlockVerboseData = blockVerboseData
	}

	return response, nil
}

func buildBlockVerboseData(context *rpccontext.Context, block *util.Block,
	includeTransactionVerboseData bool) (*appmessage.BlockVerboseData, error) {

	hash := block.Hash()
	params := context.DAG.Params
	blockHeader := block.MsgBlock().Header

	blockBlueScore, err := context.DAG.BlueScoreByBlockHash(hash)
	if err != nil {
		return nil, err
	}

	// Get the hashes for the next blocks unless there are none.
	childHashes, err := context.DAG.ChildHashesByHash(hash)
	if err != nil {
		return nil, err
	}

	blockConfirmations, err := context.DAG.BlockConfirmationsByHashNoLock(hash)
	if err != nil {
		return nil, err
	}

	selectedParentHash, err := context.DAG.SelectedParentHash(hash)
	if err != nil {
		return nil, err
	}
	selectedParentHashStr := ""
	if selectedParentHash != nil {
		selectedParentHashStr = selectedParentHash.String()
	}

	isChainBlock, err := context.DAG.IsInSelectedParentChain(hash)
	if err != nil {
		return nil, err
	}

	acceptedBlockHashes, err := context.DAG.BluesByBlockHash(hash)
	if err != nil {
		return nil, err
	}

	result := &appmessage.BlockVerboseData{
		Hash:                 hash.String(),
		Version:              blockHeader.Version,
		VersionHex:           fmt.Sprintf("%08x", blockHeader.Version),
		HashMerkleRoot:       blockHeader.HashMerkleRoot.String(),
		AcceptedIDMerkleRoot: blockHeader.AcceptedIDMerkleRoot.String(),
		UTXOCommitment:       blockHeader.UTXOCommitment.String(),
		ParentHashes:         daghash.Strings(blockHeader.ParentHashes),
		SelectedParentHash:   selectedParentHashStr,
		Nonce:                blockHeader.Nonce,
		Time:                 blockHeader.Timestamp.UnixMilliseconds(),
		Confirmations:        blockConfirmations,
		BlueScore:            blockBlueScore,
		IsChainBlock:         isChainBlock,
		Size:                 int32(block.MsgBlock().SerializeSize()),
		Bits:                 strconv.FormatInt(int64(blockHeader.Bits), 16),
		Difficulty:           getDifficultyRatio(blockHeader.Bits, params),
		ChildHashes:          daghash.Strings(childHashes),
		AcceptedBlockHashes:  daghash.Strings(acceptedBlockHashes),
	}

	if includeTransactionVerboseData {
		transactions := block.Transactions()
		txIDs := make([]string, len(transactions))
		for i, tx := range transactions {
			txIDs[i] = tx.ID().String()
		}

		result.TxIDs = txIDs
	} else {
		transactions := block.Transactions()
		transactionVerboseData := make([]*appmessage.TransactionVerboseData, len(transactions))
		for i, tx := range transactions {
			rawTxn, err := createTxRawResult(params, tx.MsgTx(), tx.ID().String(),
				&blockHeader, hash.String(), nil, false)
			if err != nil {
				return nil, err
			}
			transactionVerboseData[i] = *rawTxn
		}
		result.TransactionVerboseData = transactionVerboseData
	}

	return result, nil
}
