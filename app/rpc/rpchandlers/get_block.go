package rpchandlers

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/domain/txscript"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/pointers"
	"github.com/kaspanet/kaspad/util/subnetworkid"
	"math/big"
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
			data, err := buildTransactionVerboseData(params, tx.MsgTx(), tx.ID().String(),
				&blockHeader, hash.String(), nil, false)
			if err != nil {
				return nil, err
			}
			transactionVerboseData[i] = data
		}
		result.TransactionVerboseData = transactionVerboseData
	}

	return result, nil
}

// getDifficultyRatio returns the proof-of-work difficulty as a multiple of the
// minimum difficulty using the passed bits field from the header of a block.
func getDifficultyRatio(bits uint32, params *dagconfig.Params) float64 {
	// The minimum difficulty is the max possible proof-of-work limit bits
	// converted back to a number. Note this is not the same as the proof of
	// work limit directly because the block difficulty is encoded in a block
	// with the compact form which loses precision.
	target := util.CompactToBig(bits)

	difficulty := new(big.Rat).SetFrac(params.PowMax, target)
	outString := difficulty.FloatString(8)
	diff, err := strconv.ParseFloat(outString, 64)
	if err != nil {
		log.Errorf("Cannot get difficulty: %s", err)
		return 0
	}
	return diff
}

func buildTransactionVerboseData(dagParams *dagconfig.Params, mtx *appmessage.MsgTx,
	txID string, blkHeader *appmessage.BlockHeader, blkHash string,
	acceptingBlock *daghash.Hash, isInMempool bool) (*appmessage.TransactionVerboseData, error) {

	mtxHex, err := msgTxToHex(mtx)
	if err != nil {
		return nil, err
	}

	var payloadHash string
	if mtx.PayloadHash != nil {
		payloadHash = mtx.PayloadHash.String()
	}

	txReply := &appmessage.TransactionVerboseData{
		Hex:          mtxHex,
		TxID:         txID,
		Hash:         mtx.TxHash().String(),
		Size:         int32(mtx.SerializeSize()),
		Vin:          buildVinList(mtx),
		Vout:         createVoutList(mtx, dagParams, nil),
		Version:      mtx.Version,
		LockTime:     mtx.LockTime,
		SubnetworkID: mtx.SubnetworkID.String(),
		Gas:          mtx.Gas,
		PayloadHash:  payloadHash,
		Payload:      hex.EncodeToString(mtx.Payload),
	}

	if blkHeader != nil {
		txReply.Time = uint64(blkHeader.Timestamp.UnixMilliseconds())
		txReply.BlockTime = uint64(blkHeader.Timestamp.UnixMilliseconds())
		txReply.BlockHash = blkHash
	}

	txReply.IsInMempool = isInMempool
	if acceptingBlock != nil {
		txReply.AcceptedBy = acceptingBlock.String()
	}

	return txReply, nil
}

// msgTxToHex serializes a transaction using the latest protocol version and
// returns a hex-encoded string of the result.
func msgTxToHex(msgTx *appmessage.MsgTx) (string, error) {
	var buf bytes.Buffer
	err := msgTx.KaspaEncode(&buf, 0)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(buf.Bytes()), nil
}

func buildVinList(mtx *appmessage.MsgTx) []*appmessage.Vin {
	vinList := make([]*appmessage.Vin, len(mtx.TxIn))
	for i, txIn := range mtx.TxIn {
		// The disassembled string will contain [error] inline
		// if the script doesn't fully parse, so ignore the
		// error here.
		disbuf, _ := txscript.DisasmString(txIn.SignatureScript)

		vinEntry := vinList[i]
		vinEntry.TxID = txIn.PreviousOutpoint.TxID.String()
		vinEntry.Vout = txIn.PreviousOutpoint.Index
		vinEntry.Sequence = txIn.Sequence
		vinEntry.ScriptSig = &appmessage.ScriptSig{
			Asm: disbuf,
			Hex: hex.EncodeToString(txIn.SignatureScript),
		}
	}

	return vinList
}

// createVoutList returns a slice of JSON objects for the outputs of the passed
// transaction.
func createVoutList(mtx *appmessage.MsgTx, dagParams *dagconfig.Params, filterAddrMap map[string]struct{}) []*appmessage.Vout {
	voutList := make([]*appmessage.Vout, 0, len(mtx.TxOut))
	for i, v := range mtx.TxOut {
		// The disassembled string will contain [error] inline if the
		// script doesn't fully parse, so ignore the error here.
		disbuf, _ := txscript.DisasmString(v.ScriptPubKey)

		// Ignore the error here since an error means the script
		// couldn't parse and there is no additional information about
		// it anyways.
		scriptClass, addr, _ := txscript.ExtractScriptPubKeyAddress(
			v.ScriptPubKey, dagParams)

		// Encode the addresses while checking if the address passes the
		// filter when needed.
		passesFilter := len(filterAddrMap) == 0
		var encodedAddr string
		if addr != nil {
			encodedAddr = *pointers.String(addr.EncodeAddress())

			// If the filter doesn't already pass, make it pass if
			// the address exists in the filter.
			if _, exists := filterAddrMap[encodedAddr]; exists {
				passesFilter = true
			}
		}

		if !passesFilter {
			continue
		}

		vout := &appmessage.Vout{}
		vout.N = uint32(i)
		vout.Value = v.Value
		vout.ScriptPubKey.Address = encodedAddr
		vout.ScriptPubKey.Asm = disbuf
		vout.ScriptPubKey.Hex = hex.EncodeToString(v.ScriptPubKey)
		vout.ScriptPubKey.Type = scriptClass.String()

		voutList = append(voutList, vout)
	}

	return voutList
}
