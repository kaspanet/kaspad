package rpccontext

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/blockheader"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/estimatedsize"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/util/difficulty"

	"encoding/hex"
	"fmt"
	"github.com/pkg/errors"
	"math"
	"math/big"
	"strconv"
)

// ErrBuildBlockVerboseDataInvalidBlock indicates that a block that was given to BuildBlockVerboseData is invalid.
var ErrBuildBlockVerboseDataInvalidBlock = errors.New("ErrBuildBlockVerboseDataInvalidBlock")

// BuildBlockVerboseData builds a BlockVerboseData from the given blockHeader.
// A block may optionally also be given if it's available in the calling context.
func (ctx *Context) BuildBlockVerboseData(blockHeader externalapi.BlockHeader, block *externalapi.DomainBlock,
	includeTransactionVerboseData bool) (*appmessage.BlockVerboseData, error) {

	onEnd := logger.LogAndMeasureExecutionTime(log, "BuildBlockVerboseData")
	defer onEnd()

	hash := consensushashing.HeaderHash(blockHeader)

	blockInfo, err := ctx.Domain.Consensus().GetBlockInfo(hash)
	if err != nil {
		return nil, err
	}

	if blockInfo.BlockStatus == externalapi.StatusInvalid {
		return nil, errors.Wrap(ErrBuildBlockVerboseDataInvalidBlock, "cannot build verbose data for "+
			"invalid block")
	}

	childrenHashes, err := ctx.Domain.Consensus().GetBlockChildren(hash)
	if err != nil {
		return nil, err
	}

	result := &appmessage.BlockVerboseData{
		Hash:                 hash.String(),
		Version:              blockHeader.Version(),
		VersionHex:           fmt.Sprintf("%08x", blockHeader.Version()),
		HashMerkleRoot:       blockHeader.HashMerkleRoot().String(),
		AcceptedIDMerkleRoot: blockHeader.AcceptedIDMerkleRoot().String(),
		UTXOCommitment:       blockHeader.UTXOCommitment().String(),
		ParentHashes:         hashes.ToStrings(blockHeader.ParentHashes()),
		ChildrenHashes:       hashes.ToStrings(childrenHashes),
		Nonce:                blockHeader.Nonce(),
		Time:                 blockHeader.TimeInMilliseconds(),
		Bits:                 strconv.FormatInt(int64(blockHeader.Bits()), 16),
		Difficulty:           ctx.GetDifficultyRatio(blockHeader.Bits(), ctx.Config.ActiveNetParams),
		BlueScore:            blockInfo.BlueScore,
		IsHeaderOnly:         blockInfo.BlockStatus == externalapi.StatusHeaderOnly,
	}

	if blockInfo.BlockStatus != externalapi.StatusHeaderOnly {
		if block == nil {
			block, err = ctx.Domain.Consensus().GetBlock(hash)
			if err != nil {
				return nil, err
			}
		}

		txIDs := make([]string, len(block.Transactions))
		for i, tx := range block.Transactions {
			txIDs[i] = consensushashing.TransactionID(tx).String()
		}
		result.TxIDs = txIDs

		if includeTransactionVerboseData {
			transactionVerboseData := make([]*appmessage.TransactionVerboseData, len(block.Transactions))
			for i, tx := range block.Transactions {
				txID := consensushashing.TransactionID(tx).String()
				data, err := ctx.BuildTransactionVerboseData(tx, txID, blockHeader, hash.String())
				if err != nil {
					return nil, err
				}
				transactionVerboseData[i] = data
			}
			result.TransactionVerboseData = transactionVerboseData
		}
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
	target := difficulty.CompactToBig(bits)

	difficulty := new(big.Rat).SetFrac(params.PowMax, target)
	diff, _ := difficulty.Float64()

	roundingPrecision := float64(100)
	diff = math.Round(diff*roundingPrecision) / roundingPrecision

	return diff
}

// BuildTransactionVerboseData builds a TransactionVerboseData from
// the given parameters
func (ctx *Context) BuildTransactionVerboseData(tx *externalapi.DomainTransaction, txID string,
	blockHeader externalapi.BlockHeader, blockHash string) (
	*appmessage.TransactionVerboseData, error) {

	onEnd := logger.LogAndMeasureExecutionTime(log, "BuildTransactionVerboseData")
	defer onEnd()

	var payloadHash string
	if tx.SubnetworkID != subnetworks.SubnetworkIDNative {
		payloadHash = tx.PayloadHash.String()
	}

	txReply := &appmessage.TransactionVerboseData{
		TxID:                      txID,
		Hash:                      consensushashing.TransactionHash(tx).String(),
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
		txReply.Time = uint64(blockHeader.TimeInMilliseconds())
		txReply.BlockTime = uint64(blockHeader.TimeInMilliseconds())
		txReply.BlockHash = blockHash
	}

	return txReply, nil
}

func (ctx *Context) buildTransactionVerboseInputs(tx *externalapi.DomainTransaction) []*appmessage.TransactionVerboseInput {
	inputs := make([]*appmessage.TransactionVerboseInput, len(tx.Inputs))
	for i, transactionInput := range tx.Inputs {
		// The disassembled string will contain [error] inline
		// if the script doesn't fully parse, so ignore the
		// error here.
		disbuf, _ := txscript.DisasmString(constants.MaxScriptPublicKeyVersion, transactionInput.SignatureScript)

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
			encodedAddr = addr.EncodeAddress()

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
		output.Value = transactionOutput.Value
		output.ScriptPubKey = &appmessage.ScriptPubKeyResult{
			Version: transactionOutput.ScriptPublicKey.Version,
			Address: encodedAddr,
			Hex:     hex.EncodeToString(transactionOutput.ScriptPublicKey.Script),
			Type:    scriptClass.String(),
		}
		outputs[i] = output
	}

	return outputs
}

// ConvertTransactionVerboseDataToDomainTransaction converts a appmessage.TransactionVerboseData to externalapi.DomainTransaction
func ConvertTransactionVerboseDataToDomainTransaction(transactionVerbose *appmessage.TransactionVerboseData) (*externalapi.DomainTransaction, error) {
	inputs := make([]*externalapi.DomainTransactionInput, len(transactionVerbose.TransactionVerboseInputs))
	for inputIndex, input := range transactionVerbose.TransactionVerboseInputs {
		previousTXID, err := externalapi.NewDomainTransactionIDFromString(input.TxID)
		if err != nil {
			return nil, err
		}
		scriptSig, err := hex.DecodeString(input.ScriptSig.Hex)
		if err != nil {
			return nil, err
		}
		inputs[inputIndex] = &externalapi.DomainTransactionInput{
			PreviousOutpoint: externalapi.DomainOutpoint{
				TransactionID: *previousTXID,
				Index:         input.OutputIndex,
			},
			SignatureScript: scriptSig,
			Sequence:        input.Sequence,
		}
	}
	outputs := make([]*externalapi.DomainTransactionOutput, len(transactionVerbose.TransactionVerboseOutputs))
	for outputIndex, output := range transactionVerbose.TransactionVerboseOutputs {
		scriptPubKey, err := hex.DecodeString(output.ScriptPubKey.Hex)
		if err != nil {
			return nil, err
		}
		outputs[outputIndex] = &externalapi.DomainTransactionOutput{
			Value: output.Value,
			ScriptPublicKey: &externalapi.ScriptPublicKey{
				Script:  scriptPubKey,
				Version: 0,
			},
		}
	}
	payload, err := hex.DecodeString(transactionVerbose.Payload)
	if err != nil {
		return nil, err
	}
	subnetworkID, err := subnetworks.FromString(transactionVerbose.SubnetworkID)
	if err != nil {
		return nil, err
	}
	payloadHash, err := externalapi.NewDomainHashFromString(transactionVerbose.PayloadHash)
	if err != nil {
		return nil, err
	}
	return &externalapi.DomainTransaction{
		Version:      transactionVerbose.Version,
		Inputs:       inputs,
		Outputs:      outputs,
		LockTime:     transactionVerbose.LockTime,
		SubnetworkID: *subnetworkID,
		Gas:          transactionVerbose.Gas,
		PayloadHash:  *payloadHash,
		Payload:      payload,
	}, nil
}

// ConvertBlockVerboseDataToDomainBlock converts a appmessage.BlockVerboseData to externalapi.DomainBlock
func ConvertBlockVerboseDataToDomainBlock(blockVerbose *appmessage.BlockVerboseData) (*externalapi.DomainBlock, error) {
	var err error
	parents := make([]*externalapi.DomainHash, len(blockVerbose.ParentHashes))
	for parentIndex, parent := range blockVerbose.ParentHashes {
		parents[parentIndex], err = externalapi.NewDomainHashFromString(parent)
		if err != nil {
			return nil, err
		}
	}
	hashMerkleRoot, err := externalapi.NewDomainHashFromString(blockVerbose.HashMerkleRoot)
	if err != nil {
		return nil, err
	}
	acceptedIDMerkleRoot, err := externalapi.NewDomainHashFromString(blockVerbose.AcceptedIDMerkleRoot)
	if err != nil {
		return nil, err
	}
	utxoCommitment, err := externalapi.NewDomainHashFromString(blockVerbose.UTXOCommitment)
	if err != nil {
		return nil, err
	}
	bits, err := strconv.ParseUint(blockVerbose.Bits, 16, 32)
	if err != nil {
		return nil, err
	}
	transactions := make([]*externalapi.DomainTransaction, len(blockVerbose.TransactionVerboseData))
	for txIndex, verboseTransaction := range blockVerbose.TransactionVerboseData {
		transaction, err := ConvertTransactionVerboseDataToDomainTransaction(verboseTransaction)
		if err != nil {
			return nil, err
		}
		transactions[txIndex] = transaction
	}

	return &externalapi.DomainBlock{
		Header: blockheader.NewImmutableBlockHeader(
			blockVerbose.Version,
			parents,
			hashMerkleRoot,
			acceptedIDMerkleRoot,
			utxoCommitment,
			blockVerbose.Time,
			uint32(bits),
			blockVerbose.Nonce),
		Transactions: transactions,
	}, nil

}
