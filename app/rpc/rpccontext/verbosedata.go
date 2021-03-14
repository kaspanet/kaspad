package rpccontext

import (
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/util/difficulty"
	"github.com/pkg/errors"
	"math"
	"math/big"

	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"

	"github.com/kaspanet/kaspad/domain/consensus/utils/estimatedsize"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/dagconfig"
)

// ErrBuildBlockVerboseDataInvalidBlock indicates that a block that was given to BuildBlockVerboseData is invalid.
var ErrBuildBlockVerboseDataInvalidBlock = errors.New("ErrBuildBlockVerboseDataInvalidBlock")

// BuildBlockVerboseData builds a BlockVerboseData from the given block.
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
		Hash:           hash.String(),
		ChildrenHashes: hashes.ToStrings(childrenHashes),
		Difficulty:     ctx.GetDifficultyRatio(blockHeader.Bits(), ctx.Config.ActiveNetParams),
		BlueScore:      blockInfo.BlueScore,
		IsHeaderOnly:   blockInfo.BlockStatus == externalapi.StatusHeaderOnly,
	}

	if blockInfo.BlockStatus == externalapi.StatusHeaderOnly {
		block := &externalapi.DomainBlock{Header: blockHeader}
		result.Block = appmessage.DomainBlockToRPCBlock(block)
		return result, nil
	}

	if block == nil {
		block, err = ctx.Domain.Consensus().GetBlock(hash)
		if err != nil {
			return nil, err
		}
	}
	result.Block = appmessage.DomainBlockToRPCBlock(block)

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

	txReply := &appmessage.TransactionVerboseData{
		TxID:                      txID,
		Hash:                      consensushashing.TransactionHash(tx).String(),
		Size:                      estimatedsize.TransactionEstimatedSerializedSize(tx),
		TransactionVerboseInputs:  ctx.buildTransactionVerboseInputs(tx),
		TransactionVerboseOutputs: ctx.buildTransactionVerboseOutputs(tx, nil),
		Transaction:               appmessage.DomainTransactionToRPCTransaction(tx),
	}

	if blockHeader != nil {
		txReply.BlockTime = uint64(blockHeader.TimeInMilliseconds())
		txReply.BlockHash = blockHash
	}

	return txReply, nil
}

func (ctx *Context) buildTransactionVerboseInputs(tx *externalapi.DomainTransaction) []*appmessage.TransactionVerboseInput {
	inputs := make([]*appmessage.TransactionVerboseInput, len(tx.Inputs))
	for i := range tx.Inputs {
		inputs[i] = &appmessage.TransactionVerboseInput{}
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

		outputs[i] = &appmessage.TransactionVerboseOutput{
			ScriptPublicKeyType:    scriptClass.String(),
			ScriptPublicKeyAddress: encodedAddr,
		}
	}

	return outputs
}
