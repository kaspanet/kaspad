package rpc

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/kaspanet/kaspad/kaspajson"
	"github.com/kaspanet/kaspad/txscript"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
	"math/big"
	"strconv"
)

var (
	// ErrRPCUnimplemented is an error returned to RPC clients when the
	// provided command is recognized, but not implemented.
	ErrRPCUnimplemented = &kaspajson.RPCError{
		Code:    kaspajson.ErrRPCUnimplemented,
		Message: "Command unimplemented",
	}
)

// internalRPCError is a convenience function to convert an internal error to
// an RPC error with the appropriate code set. It also logs the error to the
// RPC server subsystem since internal errors really should not occur. The
// context parameter is only used in the log message and may be empty if it's
// not needed.
func internalRPCError(errStr, context string) *kaspajson.RPCError {
	logStr := errStr
	if context != "" {
		logStr = context + ": " + errStr
	}
	log.Error(logStr)
	return kaspajson.NewRPCError(kaspajson.ErrRPCInternal.Code, errStr)
}

// rpcDecodeHexError is a convenience function for returning a nicely formatted
// RPC error which indicates the provided hex string failed to decode.
func rpcDecodeHexError(gotHex string) *kaspajson.RPCError {
	return kaspajson.NewRPCError(kaspajson.ErrRPCDecodeHexString,
		fmt.Sprintf("Argument must be hexadecimal string (not %q)",
			gotHex))
}

// rpcNoTxInfoError is a convenience function for returning a nicely formatted
// RPC error which indicates there is no information available for the provided
// transaction hash.
func rpcNoTxInfoError(txID *daghash.TxID) *kaspajson.RPCError {
	return kaspajson.NewRPCError(kaspajson.ErrRPCNoTxInfo,
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
func createVinList(mtx *wire.MsgTx) []kaspajson.Vin {
	vinList := make([]kaspajson.Vin, len(mtx.TxIn))
	for i, txIn := range mtx.TxIn {
		// The disassembled string will contain [error] inline
		// if the script doesn't fully parse, so ignore the
		// error here.
		disbuf, _ := txscript.DisasmString(txIn.SignatureScript)

		vinEntry := &vinList[i]
		vinEntry.TxID = txIn.PreviousOutpoint.TxID.String()
		vinEntry.Vout = txIn.PreviousOutpoint.Index
		vinEntry.Sequence = txIn.Sequence
		vinEntry.ScriptSig = &kaspajson.ScriptSig{
			Asm: disbuf,
			Hex: hex.EncodeToString(txIn.SignatureScript),
		}
	}

	return vinList
}

// createVoutList returns a slice of JSON objects for the outputs of the passed
// transaction.
func createVoutList(mtx *wire.MsgTx, chainParams *dagconfig.Params, filterAddrMap map[string]struct{}) []kaspajson.Vout {
	voutList := make([]kaspajson.Vout, 0, len(mtx.TxOut))
	for i, v := range mtx.TxOut {
		// The disassembled string will contain [error] inline if the
		// script doesn't fully parse, so ignore the error here.
		disbuf, _ := txscript.DisasmString(v.ScriptPubKey)

		// Ignore the error here since an error means the script
		// couldn't parse and there is no additional information about
		// it anyways.
		scriptClass, addr, _ := txscript.ExtractScriptPubKeyAddress(
			v.ScriptPubKey, chainParams)

		// Encode the addresses while checking if the address passes the
		// filter when needed.
		passesFilter := len(filterAddrMap) == 0
		var encodedAddr *string
		if addr != nil {
			encodedAddr = kaspajson.String(addr.EncodeAddress())

			// If the filter doesn't already pass, make it pass if
			// the address exists in the filter.
			if _, exists := filterAddrMap[*encodedAddr]; exists {
				passesFilter = true
			}
		}

		if !passesFilter {
			continue
		}

		var vout kaspajson.Vout
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
	acceptingBlock *daghash.Hash, confirmations *uint64, isInMempool bool) (*kaspajson.TxRawResult, error) {

	mtxHex, err := messageToHex(mtx)
	if err != nil {
		return nil, err
	}

	var payloadHash string
	if mtx.PayloadHash != nil {
		payloadHash = mtx.PayloadHash.String()
	}

	txReply := &kaspajson.TxRawResult{
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
		// This is not a typo, they are identical in bitcoind as well.
		txReply.Time = uint64(blkHeader.Timestamp.Unix())
		txReply.BlockTime = uint64(blkHeader.Timestamp.Unix())
		txReply.BlockHash = blkHash
	}

	txReply.Confirmations = confirmations
	txReply.IsInMempool = isInMempool
	if acceptingBlock != nil {
		txReply.AcceptedBy = kaspajson.String(acceptingBlock.String())
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

// buildGetBlockVerboseResult takes a block and convert it to kaspajson.GetBlockVerboseResult
//
// This function MUST be called with the DAG state lock held (for reads).
func buildGetBlockVerboseResult(s *Server, block *util.Block, isVerboseTx bool) (*kaspajson.GetBlockVerboseResult, error) {
	hash := block.Hash()
	params := s.cfg.DAGParams
	blockHeader := block.MsgBlock().Header

	// Get the block chain height.
	blockChainHeight, err := s.cfg.DAG.BlockChainHeightByHash(hash)
	if err != nil {
		context := "Failed to obtain block height"
		return nil, internalRPCError(err.Error(), context)
	}

	// Get the hashes for the next blocks unless there are none.
	var nextHashStrings []string
	if blockChainHeight < s.cfg.DAG.ChainHeight() { //TODO: (Ori) This is probably wrong. Done only for compilation
		childHashes, err := s.cfg.DAG.ChildHashesByHash(hash)
		if err != nil {
			context := "No next block"
			return nil, internalRPCError(err.Error(), context)
		}
		nextHashStrings = daghash.Strings(childHashes)
	}

	blockConfirmations, err := s.cfg.DAG.BlockConfirmationsByHashNoLock(hash)
	if err != nil {
		context := "Could not get block confirmations"
		return nil, internalRPCError(err.Error(), context)
	}

	blockBlueScore, err := s.cfg.DAG.BlueScoreByBlockHash(hash)
	if err != nil {
		context := "Could not get block blue score"
		return nil, internalRPCError(err.Error(), context)
	}

	isChainBlock := s.cfg.DAG.IsInSelectedParentChain(hash)

	result := &kaspajson.GetBlockVerboseResult{
		Hash:                 hash.String(),
		Version:              blockHeader.Version,
		VersionHex:           fmt.Sprintf("%08x", blockHeader.Version),
		HashMerkleRoot:       blockHeader.HashMerkleRoot.String(),
		AcceptedIDMerkleRoot: blockHeader.AcceptedIDMerkleRoot.String(),
		UTXOCommitment:       blockHeader.UTXOCommitment.String(),
		ParentHashes:         daghash.Strings(blockHeader.ParentHashes),
		Nonce:                blockHeader.Nonce,
		Time:                 blockHeader.Timestamp.Unix(),
		Confirmations:        blockConfirmations,
		Height:               blockChainHeight,
		BlueScore:            blockBlueScore,
		IsChainBlock:         isChainBlock,
		Size:                 int32(block.MsgBlock().SerializeSize()),
		Bits:                 strconv.FormatInt(int64(blockHeader.Bits), 16),
		Difficulty:           getDifficultyRatio(blockHeader.Bits, params),
		NextHashes:           nextHashStrings,
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
		rawTxns := make([]kaspajson.TxRawResult, len(txns))
		for i, tx := range txns {
			rawTxn, err := createTxRawResult(params, tx.MsgTx(), tx.ID().String(),
				&blockHeader, hash.String(), nil, nil, false)
			if err != nil {
				return nil, err
			}
			rawTxns[i] = *rawTxn
		}
		result.RawTx = rawTxns
	}

	return result, nil
}

func collectChainBlocks(s *Server, hashes []*daghash.Hash) ([]kaspajson.ChainBlock, error) {
	chainBlocks := make([]kaspajson.ChainBlock, 0, len(hashes))
	for _, hash := range hashes {
		acceptanceData, err := s.cfg.AcceptanceIndex.TxsAcceptanceData(hash)
		if err != nil {
			return nil, &kaspajson.RPCError{
				Code:    kaspajson.ErrRPCInternal.Code,
				Message: fmt.Sprintf("could not retrieve acceptance data for block %s", hash),
			}
		}

		acceptedBlocks := make([]kaspajson.AcceptedBlock, 0, len(acceptanceData))
		for _, blockAcceptanceData := range acceptanceData {
			acceptedTxIds := make([]string, 0, len(blockAcceptanceData.TxAcceptanceData))
			for _, txAcceptanceData := range blockAcceptanceData.TxAcceptanceData {
				if txAcceptanceData.IsAccepted {
					acceptedTxIds = append(acceptedTxIds, txAcceptanceData.Tx.ID().String())
				}
			}
			acceptedBlock := kaspajson.AcceptedBlock{
				Hash:          blockAcceptanceData.BlockHash.String(),
				AcceptedTxIDs: acceptedTxIds,
			}
			acceptedBlocks = append(acceptedBlocks, acceptedBlock)
		}

		chainBlock := kaspajson.ChainBlock{
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
func hashesToGetBlockVerboseResults(s *Server, hashes []*daghash.Hash) ([]kaspajson.GetBlockVerboseResult, error) {
	getBlockVerboseResults := make([]kaspajson.GetBlockVerboseResult, 0, len(hashes))
	for _, blockHash := range hashes {
		block, err := s.cfg.DAG.BlockByHash(blockHash)
		if err != nil {
			return nil, &kaspajson.RPCError{
				Code:    kaspajson.ErrRPCInternal.Code,
				Message: fmt.Sprintf("could not retrieve block %s.", blockHash),
			}
		}
		getBlockVerboseResult, err := buildGetBlockVerboseResult(s, block, false)
		if err != nil {
			return nil, &kaspajson.RPCError{
				Code:    kaspajson.ErrRPCInternal.Code,
				Message: fmt.Sprintf("could not build getBlockVerboseResult for block %s: %s", blockHash, err),
			}
		}
		getBlockVerboseResults = append(getBlockVerboseResults, *getBlockVerboseResult)
	}
	return getBlockVerboseResults, nil
}

// txConfirmationsNoLock returns the confirmations number for the given transaction
// The confirmations number is defined as follows:
// If the transaction is in the mempool/in a red block/is a double spend -> 0
// Otherwise -> The confirmations number of the accepting block
//
// This function MUST be called with the DAG state lock held (for reads).
func txConfirmationsNoLock(s *Server, txID *daghash.TxID) (uint64, error) {
	if s.cfg.TxIndex == nil {
		return 0, errors.New("transaction index must be enabled (--txindex)")
	}

	acceptingBlock, err := s.cfg.TxIndex.BlockThatAcceptedTx(s.cfg.DAG, txID)
	if err != nil {
		return 0, errors.Errorf("could not get block that accepted tx %s: %s", txID, err)
	}
	if acceptingBlock == nil {
		return 0, nil
	}

	confirmations, err := s.cfg.DAG.BlockConfirmationsByHashNoLock(acceptingBlock)
	if err != nil {
		return 0, errors.Errorf("could not get confirmations for block that accepted tx %s: %s", txID, err)
	}

	return confirmations, nil
}

func txConfirmations(s *Server, txID *daghash.TxID) (uint64, error) {
	s.cfg.DAG.RLock()
	defer s.cfg.DAG.RUnlock()
	return txConfirmationsNoLock(s, txID)
}
