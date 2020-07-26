package rpc

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/kaspanet/kaspad/rpcmodel"
	"github.com/kaspanet/kaspad/txscript"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/pointers"
	"github.com/kaspanet/kaspad/wire"
	"math/big"
	"strconv"
)

var (
	// ErrRPCUnimplemented is an error returned to RPC clients when the
	// provided command is recognized, but not implemented.
	ErrRPCUnimplemented = &rpcmodel.RPCError{
		Code:    rpcmodel.ErrRPCUnimplemented,
		Message: "Command unimplemented",
	}
)

// internalRPCError is a convenience function to convert an internal error to
// an RPC error with the appropriate code set. It also logs the error to the
// RPC server subsystem since internal errors really should not occur. The
// context parameter is only used in the log message and may be empty if it's
// not needed.
func internalRPCError(errStr, context string) *rpcmodel.RPCError {
	logStr := errStr
	if context != "" {
		logStr = context + ": " + errStr
	}
	log.Error(logStr)
	return rpcmodel.NewRPCError(rpcmodel.ErrRPCInternal.Code, errStr)
}

// rpcDecodeHexError is a convenience function for returning a nicely formatted
// RPC error which indicates the provided hex string failed to decode.
func rpcDecodeHexError(gotHex string) *rpcmodel.RPCError {
	return rpcmodel.NewRPCError(rpcmodel.ErrRPCDecodeHexString,
		fmt.Sprintf("Argument must be hexadecimal string (not %q)",
			gotHex))
}

// rpcNoTxInfoError is a convenience function for returning a nicely formatted
// RPC error which indicates there is no information available for the provided
// transaction hash.
func rpcNoTxInfoError(txID *daghash.TxID) *rpcmodel.RPCError {
	return rpcmodel.NewRPCError(rpcmodel.ErrRPCNoTxInfo,
		fmt.Sprintf("No information available about transaction %s",
			txID))
}

// messageToHex serializes a message to the wire protocol encoding using the
// latest protocol version and returns a hex-encoded string of the result.
func messageToHex(msg wire.Message) (string, error) {
	var buf bytes.Buffer
	if err := msg.KaspaEncode(&buf, maxProtocolVersion); err != nil {
		context := fmt.Sprintf("Failed to encode msg of type %T", msg)
		return "", internalRPCError(err.Error(), context)
	}

	return hex.EncodeToString(buf.Bytes()), nil
}

// createVinList returns a slice of JSON objects for the inputs of the passed
// transaction.
func createVinList(mtx *wire.MsgTx) []rpcmodel.Vin {
	vinList := make([]rpcmodel.Vin, len(mtx.TxIn))
	for i, txIn := range mtx.TxIn {
		// The disassembled string will contain [error] inline
		// if the script doesn't fully parse, so ignore the
		// error here.
		disbuf, _ := txscript.DisasmString(txIn.SignatureScript)

		vinEntry := &vinList[i]
		vinEntry.TxID = txIn.PreviousOutpoint.TxID.String()
		vinEntry.Vout = txIn.PreviousOutpoint.Index
		vinEntry.Sequence = txIn.Sequence
		vinEntry.ScriptSig = &rpcmodel.ScriptSig{
			Asm: disbuf,
			Hex: hex.EncodeToString(txIn.SignatureScript),
		}
	}

	return vinList
}

// createVoutList returns a slice of JSON objects for the outputs of the passed
// transaction.
func createVoutList(mtx *wire.MsgTx, dagParams *dagconfig.Params, filterAddrMap map[string]struct{}) []rpcmodel.Vout {
	voutList := make([]rpcmodel.Vout, 0, len(mtx.TxOut))
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
		var encodedAddr *string
		if addr != nil {
			encodedAddr = pointers.String(addr.EncodeAddress())

			// If the filter doesn't already pass, make it pass if
			// the address exists in the filter.
			if _, exists := filterAddrMap[*encodedAddr]; exists {
				passesFilter = true
			}
		}

		if !passesFilter {
			continue
		}

		var vout rpcmodel.Vout
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

// createTxRawResult converts the passed transaction and associated parameters
// to a raw transaction JSON object.
func createTxRawResult(dagParams *dagconfig.Params, mtx *wire.MsgTx,
	txID string, blkHeader *wire.BlockHeader, blkHash string,
	acceptingBlock *daghash.Hash, isInMempool bool) (*rpcmodel.TxRawResult, error) {

	mtxHex, err := messageToHex(mtx)
	if err != nil {
		return nil, err
	}

	var payloadHash string
	if mtx.PayloadHash != nil {
		payloadHash = mtx.PayloadHash.String()
	}

	txReply := &rpcmodel.TxRawResult{
		Hex:         mtxHex,
		TxID:        txID,
		Hash:        mtx.TxHash().String(),
		Size:        int32(mtx.SerializeSize()),
		Vin:         createVinList(mtx),
		Vout:        createVoutList(mtx, dagParams, nil),
		Version:     mtx.Version,
		LockTime:    mtx.LockTime,
		Subnetwork:  mtx.SubnetworkID.String(),
		Gas:         mtx.Gas,
		PayloadHash: payloadHash,
		Payload:     hex.EncodeToString(mtx.Payload),
	}

	if blkHeader != nil {
		txReply.Time = uint64(blkHeader.Timestamp.UnixMilliseconds())
		txReply.BlockTime = uint64(blkHeader.Timestamp.UnixMilliseconds())
		txReply.BlockHash = blkHash
	}

	txReply.IsInMempool = isInMempool
	if acceptingBlock != nil {
		txReply.AcceptedBy = pointers.String(acceptingBlock.String())
	}

	return txReply, nil
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

// buildGetBlockVerboseResult takes a block and convert it to rpcmodel.GetBlockVerboseResult
//
// This function MUST be called with the DAG state lock held (for reads).
func buildGetBlockVerboseResult(s *Server, block *util.Block, isVerboseTx bool) (*rpcmodel.GetBlockVerboseResult, error) {
	hash := block.Hash()
	params := s.cfg.DAGParams
	blockHeader := block.MsgBlock().Header

	blockBlueScore, err := s.cfg.DAG.BlueScoreByBlockHash(hash)
	if err != nil {
		context := "Could not get block blue score"
		return nil, internalRPCError(err.Error(), context)
	}

	// Get the hashes for the next blocks unless there are none.
	childHashes, err := s.cfg.DAG.ChildHashesByHash(hash)
	if err != nil {
		context := "No next block"
		return nil, internalRPCError(err.Error(), context)
	}

	blockConfirmations, err := s.cfg.DAG.BlockConfirmationsByHashNoLock(hash)
	if err != nil {
		context := "Could not get block confirmations"
		return nil, internalRPCError(err.Error(), context)
	}

	selectedParentHash, err := s.cfg.DAG.SelectedParentHash(hash)
	if err != nil {
		context := "Could not get block selected parent"
		return nil, internalRPCError(err.Error(), context)
	}
	selectedParentHashStr := ""
	if selectedParentHash != nil {
		selectedParentHashStr = selectedParentHash.String()
	}

	isChainBlock, err := s.cfg.DAG.IsInSelectedParentChain(hash)
	if err != nil {
		context := "Could not get whether block is in the selected parent chain"
		return nil, internalRPCError(err.Error(), context)
	}

	acceptedBlockHashes, err := s.cfg.DAG.BluesByBlockHash(hash)
	if err != nil {
		context := fmt.Sprintf("Could not get block accepted blocks for block %s", hash)
		return nil, internalRPCError(err.Error(), context)
	}

	result := &rpcmodel.GetBlockVerboseResult{
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

	if isVerboseTx {
		transactions := block.Transactions()
		txNames := make([]string, len(transactions))
		for i, tx := range transactions {
			txNames[i] = tx.ID().String()
		}

		result.Tx = txNames
	} else {
		txns := block.Transactions()
		rawTxns := make([]rpcmodel.TxRawResult, len(txns))
		for i, tx := range txns {
			rawTxn, err := createTxRawResult(params, tx.MsgTx(), tx.ID().String(),
				&blockHeader, hash.String(), nil, false)
			if err != nil {
				return nil, err
			}
			rawTxns[i] = *rawTxn
		}
		result.RawTx = rawTxns
	}

	return result, nil
}

func collectChainBlocks(s *Server, hashes []*daghash.Hash) ([]rpcmodel.ChainBlock, error) {
	chainBlocks := make([]rpcmodel.ChainBlock, 0, len(hashes))
	for _, hash := range hashes {
		acceptanceData, err := s.cfg.AcceptanceIndex.TxsAcceptanceData(hash)
		if err != nil {
			return nil, &rpcmodel.RPCError{
				Code:    rpcmodel.ErrRPCInternal.Code,
				Message: fmt.Sprintf("could not retrieve acceptance data for block %s", hash),
			}
		}

		acceptedBlocks := make([]rpcmodel.AcceptedBlock, 0, len(acceptanceData))
		for _, blockAcceptanceData := range acceptanceData {
			acceptedTxIds := make([]string, 0, len(blockAcceptanceData.TxAcceptanceData))
			for _, txAcceptanceData := range blockAcceptanceData.TxAcceptanceData {
				if txAcceptanceData.IsAccepted {
					acceptedTxIds = append(acceptedTxIds, txAcceptanceData.Tx.ID().String())
				}
			}
			acceptedBlock := rpcmodel.AcceptedBlock{
				Hash:          blockAcceptanceData.BlockHash.String(),
				AcceptedTxIDs: acceptedTxIds,
			}
			acceptedBlocks = append(acceptedBlocks, acceptedBlock)
		}

		chainBlock := rpcmodel.ChainBlock{
			Hash:           hash.String(),
			AcceptedBlocks: acceptedBlocks,
		}
		chainBlocks = append(chainBlocks, chainBlock)
	}
	return chainBlocks, nil
}

// hashesToGetBlockVerboseResults takes block hashes and returns their
// correspondent block verbose.
//
// This function MUST be called with the DAG state lock held (for reads).
func hashesToGetBlockVerboseResults(s *Server, hashes []*daghash.Hash) ([]rpcmodel.GetBlockVerboseResult, error) {
	getBlockVerboseResults := make([]rpcmodel.GetBlockVerboseResult, 0, len(hashes))
	for _, blockHash := range hashes {
		block, err := s.cfg.DAG.BlockByHash(blockHash)
		if err != nil {
			return nil, &rpcmodel.RPCError{
				Code:    rpcmodel.ErrRPCInternal.Code,
				Message: fmt.Sprintf("could not retrieve block %s.", blockHash),
			}
		}
		getBlockVerboseResult, err := buildGetBlockVerboseResult(s, block, false)
		if err != nil {
			return nil, &rpcmodel.RPCError{
				Code:    rpcmodel.ErrRPCInternal.Code,
				Message: fmt.Sprintf("could not build getBlockVerboseResult for block %s: %s", blockHash, err),
			}
		}
		getBlockVerboseResults = append(getBlockVerboseResults, *getBlockVerboseResult)
	}
	return getBlockVerboseResults, nil
}
