package rpccontext

import (
	"encoding/hex"
	"math"
	"math/big"

	difficultyPackage "github.com/kaspanet/kaspad/util/difficulty"
	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"

	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/dagconfig"
)

// ErrBuildBlockVerboseDataInvalidBlock indicates that a block that was given to BuildBlockVerboseData is invalid.
var ErrBuildBlockVerboseDataInvalidBlock = errors.New("ErrBuildBlockVerboseDataInvalidBlock")

// GetDifficultyRatio returns the proof-of-work difficulty as a multiple of the
// minimum difficulty using the passed bits field from the header of a block.
func (ctx *Context) GetDifficultyRatio(bits uint32, params *dagconfig.Params) float64 {
	// The minimum difficulty is the max possible proof-of-work limit bits
	// converted back to a number. Note this is not the same as the proof of
	// work limit directly because the block difficulty is encoded in a block
	// with the compact form which loses precision.
	target := difficultyPackage.CompactToBig(bits)

	difficulty := new(big.Rat).SetFrac(params.PowMax, target)
	diff, _ := difficulty.Float64()

	roundingPrecision := float64(100)
	diff = math.Round(diff*roundingPrecision) / roundingPrecision

	return diff
}

// PopulateBlockWithVerboseData populates the given `block` with verbose
// data from `domainBlockHeader` and optionally from `domainBlock`
func (ctx *Context) PopulateBlockWithVerboseData(block *appmessage.RPCBlock, domainBlockHeader externalapi.BlockHeader,
	domainBlock *externalapi.DomainBlock, includeTransactionVerboseData bool) error {

	blockHash := consensushashing.HeaderHash(domainBlockHeader)

	blockInfo, err := ctx.Domain.Consensus().GetBlockInfo(blockHash)
	if err != nil {
		return err
	}

	if blockInfo.BlockStatus == externalapi.StatusInvalid {
		return errors.Wrap(ErrBuildBlockVerboseDataInvalidBlock, "cannot build verbose data for "+
			"invalid block")
	}

	_, selectedParentHash, childrenHashes, err := ctx.Domain.Consensus().GetBlockRelations(blockHash)
	if err != nil {
		return err
	}

	block.VerboseData = &appmessage.RPCBlockVerboseData{
		Hash:           blockHash.String(),
		Difficulty:     ctx.GetDifficultyRatio(domainBlockHeader.Bits(), ctx.Config.ActiveNetParams),
		ChildrenHashes: hashes.ToStrings(childrenHashes),
		IsHeaderOnly:   blockInfo.BlockStatus == externalapi.StatusHeaderOnly,
		BlueScore:      blockInfo.BlueScore,
	}
	// selectedParentHash will be nil in the genesis block
	if selectedParentHash != nil {
		block.VerboseData.SelectedParentHash = selectedParentHash.String()
	}

	if blockInfo.BlockStatus == externalapi.StatusHeaderOnly {
		return nil
	}

	// Get the block if we didn't receive it previously
	if domainBlock == nil {
		domainBlock, err = ctx.Domain.Consensus().GetBlockEvenIfHeaderOnly(blockHash)
		if err != nil {
			return err
		}
	}

	transactionIDs := make([]string, len(domainBlock.Transactions))
	for i, transaction := range domainBlock.Transactions {
		transactionIDs[i] = consensushashing.TransactionID(transaction).String()
	}
	block.VerboseData.TransactionIDs = transactionIDs

	if includeTransactionVerboseData {
		for _, transaction := range block.Transactions {
			err := ctx.PopulateTransactionWithVerboseData(transaction, domainBlockHeader)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// PopulateTransactionWithVerboseData populates the given `transaction` with
// verbose data from `domainTransaction`
func (ctx *Context) PopulateTransactionWithVerboseData(
	transaction *appmessage.RPCTransaction, domainBlockHeader externalapi.BlockHeader) error {

	domainTransaction, err := appmessage.RPCTransactionToDomainTransaction(transaction)
	if err != nil {
		return err
	}

	transaction.VerboseData = &appmessage.RPCTransactionVerboseData{
		TransactionID: consensushashing.TransactionID(domainTransaction).String(),
		Hash:          consensushashing.TransactionHash(domainTransaction).String(),
		Mass:          ctx.Domain.Consensus().TransactionMass(domainTransaction),
	}
	if domainBlockHeader != nil {
		transaction.VerboseData.BlockHash = consensushashing.HeaderHash(domainBlockHeader).String()
		transaction.VerboseData.BlockTime = uint64(domainBlockHeader.TimeInMilliseconds())
	}
	for _, input := range transaction.Inputs {
		ctx.populateTransactionInputWithVerboseData(input)
	}
	for _, output := range transaction.Outputs {
		err := ctx.populateTransactionOutputWithVerboseData(output)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ctx *Context) populateTransactionInputWithVerboseData(transactionInput *appmessage.RPCTransactionInput) {
	transactionInput.VerboseData = &appmessage.RPCTransactionInputVerboseData{}
}

func (ctx *Context) populateTransactionOutputWithVerboseData(transactionOutput *appmessage.RPCTransactionOutput) error {
	scriptPublicKey, err := hex.DecodeString(transactionOutput.ScriptPublicKey.Script)
	if err != nil {
		return err
	}
	domainScriptPublicKey := &externalapi.ScriptPublicKey{
		Script:  scriptPublicKey,
		Version: transactionOutput.ScriptPublicKey.Version,
	}

	// Ignore the error here since an error means the script
	// couldn't be parsed and there's no additional information about
	// it anyways
	scriptPublicKeyType, scriptPublicKeyAddress, _ := txscript.ExtractScriptPubKeyAddress(
		domainScriptPublicKey, ctx.Config.ActiveNetParams)

	var encodedScriptPublicKeyAddress string
	if scriptPublicKeyAddress != nil {
		encodedScriptPublicKeyAddress = scriptPublicKeyAddress.EncodeAddress()
	}
	transactionOutput.VerboseData = &appmessage.RPCTransactionOutputVerboseData{
		ScriptPublicKeyType:    scriptPublicKeyType.String(),
		ScriptPublicKeyAddress: encodedScriptPublicKeyAddress,
	}
	return nil
}
