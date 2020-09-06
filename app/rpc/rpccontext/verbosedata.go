package rpccontext

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/domain/txscript"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/pointers"
	"math/big"
	"strconv"
)

func (ctx *Context) BuildBlockVerboseData(block *util.Block, includeTransactionVerboseData bool) (*appmessage.BlockVerboseData, error) {
	hash := block.Hash()
	params := ctx.DAG.Params
	blockHeader := block.MsgBlock().Header

	blockBlueScore, err := ctx.DAG.BlueScoreByBlockHash(hash)
	if err != nil {
		return nil, err
	}

	// Get the hashes for the next blocks unless there are none.
	childHashes, err := ctx.DAG.ChildHashesByHash(hash)
	if err != nil {
		return nil, err
	}

	blockConfirmations, err := ctx.DAG.BlockConfirmationsByHashNoLock(hash)
	if err != nil {
		return nil, err
	}

	selectedParentHash, err := ctx.DAG.SelectedParentHash(hash)
	if err != nil {
		return nil, err
	}
	selectedParentHashStr := ""
	if selectedParentHash != nil {
		selectedParentHashStr = selectedParentHash.String()
	}

	isChainBlock, err := ctx.DAG.IsInSelectedParentChain(hash)
	if err != nil {
		return nil, err
	}

	acceptedBlockHashes, err := ctx.DAG.BluesByBlockHash(hash)
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

	transactions := block.Transactions()
	txIDs := make([]string, len(transactions))
	for i, tx := range transactions {
		txIDs[i] = tx.ID().String()
	}
	result.TxIDs = txIDs

	if includeTransactionVerboseData {
		transactions := block.Transactions()
		transactionVerboseData := make([]*appmessage.TransactionVerboseData, len(transactions))
		for i, tx := range transactions {
			data, err := ctx.buildTransactionVerboseData(tx.MsgTx(), tx.ID().String(),
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

func (ctx *Context) buildTransactionVerboseData(mtx *appmessage.MsgTx,
	txID string, blockHeader *appmessage.BlockHeader, blockHash string,
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
		Vin:          ctx.buildVinList(mtx),
		Vout:         ctx.createVoutList(mtx, nil),
		Version:      mtx.Version,
		LockTime:     mtx.LockTime,
		SubnetworkID: mtx.SubnetworkID.String(),
		Gas:          mtx.Gas,
		PayloadHash:  payloadHash,
		Payload:      hex.EncodeToString(mtx.Payload),
	}

	if blockHeader != nil {
		txReply.Time = uint64(blockHeader.Timestamp.UnixMilliseconds())
		txReply.BlockTime = uint64(blockHeader.Timestamp.UnixMilliseconds())
		txReply.BlockHash = blockHash
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

func (ctx *Context) buildVinList(mtx *appmessage.MsgTx) []*appmessage.Vin {
	vinList := make([]*appmessage.Vin, len(mtx.TxIn))
	for i, txIn := range mtx.TxIn {
		// The disassembled string will contain [error] inline
		// if the script doesn't fully parse, so ignore the
		// error here.
		disbuf, _ := txscript.DisasmString(txIn.SignatureScript)

		vinEntry := &appmessage.Vin{}
		vinEntry.TxID = txIn.PreviousOutpoint.TxID.String()
		vinEntry.Vout = txIn.PreviousOutpoint.Index
		vinEntry.Sequence = txIn.Sequence
		vinEntry.ScriptSig = &appmessage.ScriptSig{
			Asm: disbuf,
			Hex: hex.EncodeToString(txIn.SignatureScript),
		}
		vinList[i] = vinEntry
	}

	return vinList
}

// createVoutList returns a slice of JSON objects for the outputs of the passed
// transaction.
func (ctx *Context) createVoutList(mtx *appmessage.MsgTx, filterAddrMap map[string]struct{}) []*appmessage.Vout {
	voutList := make([]*appmessage.Vout, len(mtx.TxOut))
	for i, v := range mtx.TxOut {
		// The disassembled string will contain [error] inline if the
		// script doesn't fully parse, so ignore the error here.
		disbuf, _ := txscript.DisasmString(v.ScriptPubKey)

		// Ignore the error here since an error means the script
		// couldn't parse and there is no additional information about
		// it anyways.
		scriptClass, addr, _ := txscript.ExtractScriptPubKeyAddress(
			v.ScriptPubKey, ctx.DAG.Params)

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
		vout.ScriptPubKey = &appmessage.ScriptPubKeyResult{
			Address: encodedAddr,
			Asm:     disbuf,
			Hex:     hex.EncodeToString(v.ScriptPubKey),
			Type:    scriptClass.String(),
		}
		voutList[i] = vout
	}

	return voutList
}
