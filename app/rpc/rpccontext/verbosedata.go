package rpccontext

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"

	"github.com/kaspanet/kaspad/domain/consensus/utils/estimatedsize"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"

	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashserialization"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/pointers"
)

// BuildBlockVerboseData builds a BlockVerboseData from the given block.
// This method must be called with the DAG lock held for reads
func (ctx *Context) BuildBlockVerboseData(block *externalapi.DomainBlock, includeTransactionVerboseData bool) (*appmessage.BlockVerboseData, error) {
	hash := hashserialization.BlockHash(block)
	blockHeader := block.Header

	result := &appmessage.BlockVerboseData{
		Hash:                 hash.String(),
		Version:              blockHeader.Version,
		VersionHex:           fmt.Sprintf("%08x", blockHeader.Version),
		HashMerkleRoot:       blockHeader.HashMerkleRoot.String(),
		AcceptedIDMerkleRoot: blockHeader.AcceptedIDMerkleRoot.String(),
		UTXOCommitment:       blockHeader.UTXOCommitment.String(),
		ParentHashes:         externalapi.DomainHashesToStrings(blockHeader.ParentHashes),
		Nonce:                blockHeader.Nonce,
		Time:                 blockHeader.TimeInMilliseconds,
		Bits:                 strconv.FormatInt(int64(blockHeader.Bits), 16),
		Difficulty:           ctx.GetDifficultyRatio(blockHeader.Bits, ctx.Config.ActiveNetParams),
	}

	txIDs := make([]string, len(block.Transactions))
	for i, tx := range block.Transactions {
		txIDs[i] = hashserialization.TransactionID(tx).String()
	}
	result.TxIDs = txIDs

	if includeTransactionVerboseData {
		transactionVerboseData := make([]*appmessage.TransactionVerboseData, len(block.Transactions))
		for i, tx := range block.Transactions {
			txID := hashserialization.TransactionID(tx).String()
			data, err := ctx.BuildTransactionVerboseData(tx, txID, blockHeader, hash.String())
			if err != nil {
				return nil, err
			}
			transactionVerboseData[i] = data
		}
		result.TransactionVerboseData = transactionVerboseData
	}

	return result, nil
}

// GetDifficultyRatio returns the proof-of-work difficulty as a multiple of the
// minimum difficulty using the passed bits field from the header of a block.
func (ctx *Context) GetDifficultyRatio(bits uint32, params *dagconfig.Params) float64 {
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

// BuildTransactionVerboseData builds a TransactionVerboseData from
// the given parameters
func (ctx *Context) BuildTransactionVerboseData(tx *externalapi.DomainTransaction, txID string,
	blockHeader *externalapi.DomainBlockHeader, blockHash string) (
	*appmessage.TransactionVerboseData, error) {

	var payloadHash string
	if tx.SubnetworkID != subnetworks.SubnetworkIDNative {
		payloadHash = tx.PayloadHash.String()
	}

	txReply := &appmessage.TransactionVerboseData{
		TxID:                      txID,
		Hash:                      hashserialization.TransactionHash(tx).String(),
		Size:                      estimatedsize.TransactionEstimatedSerializedSize(tx),
		TransactionVerboseInputs:  ctx.buildTransactionVerboseInputs(tx),
		TransactionVerboseOutputs: ctx.buildTransactionVerboseOutputs(tx, nil),
		Version:                   tx.Version,
		LockTime:                  tx.LockTime,
		SubnetworkID:              tx.SubnetworkID.String(),
		Gas:                       tx.Gas,
		PayloadHash:               payloadHash,
		Payload:                   hex.EncodeToString(tx.Payload),
	}

	if blockHeader != nil {
		txReply.Time = uint64(blockHeader.TimeInMilliseconds)
		txReply.BlockTime = uint64(blockHeader.TimeInMilliseconds)
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

func (ctx *Context) buildTransactionVerboseInputs(tx *externalapi.DomainTransaction) []*appmessage.TransactionVerboseInput {
	inputs := make([]*appmessage.TransactionVerboseInput, len(tx.Inputs))
	for i, transactionInput := range tx.Inputs {
		// The disassembled string will contain [error] inline
		// if the script doesn't fully parse, so ignore the
		// error here.
		disbuf, _ := txscript.DisasmString(transactionInput.SignatureScript)

		input := &appmessage.TransactionVerboseInput{}
		input.TxID = transactionInput.PreviousOutpoint.TransactionID.String()
		input.OutputIndex = transactionInput.PreviousOutpoint.Index
		input.Sequence = transactionInput.Sequence
		input.ScriptSig = &appmessage.ScriptSig{
			Asm: disbuf,
			Hex: hex.EncodeToString(transactionInput.SignatureScript),
		}
		inputs[i] = input
	}

	return inputs
}

// buildTransactionVerboseOutputs returns a slice of JSON objects for the outputs of the passed
// transaction.
func (ctx *Context) buildTransactionVerboseOutputs(tx *externalapi.DomainTransaction, filterAddrMap map[string]struct{}) []*appmessage.TransactionVerboseOutput {
	outputs := make([]*appmessage.TransactionVerboseOutput, len(tx.Outputs))
	for i, transactionOutput := range tx.Outputs {
		// The disassembled string will contain [error] inline if the
		// script doesn't fully parse, so ignore the error here.
		disbuf, _ := txscript.DisasmString(transactionOutput.ScriptPublicKey)

		// Ignore the error here since an error means the script
		// couldn't parse and there is no additional information about
		// it anyways.
		scriptClass, addr, _ := txscript.ExtractScriptPubKeyAddress(
			transactionOutput.ScriptPublicKey, ctx.Config.ActiveNetParams)

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

		output := &appmessage.TransactionVerboseOutput{}
		output.Index = uint32(i)
		output.Value = output.Value
		output.ScriptPubKey = &appmessage.ScriptPubKeyResult{
			Address: encodedAddr,
			Asm:     disbuf,
			Hex:     hex.EncodeToString(transactionOutput.ScriptPublicKey),
			Type:    scriptClass.String(),
		}
		outputs[i] = output
	}

	return outputs
}
