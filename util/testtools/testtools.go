package testtools

import (
	"fmt"

	"github.com/daglabs/btcd/dagconfig"

	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/mining"

	"github.com/daglabs/btcd/blockdag"

	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/util/subnetworkid"
	"github.com/daglabs/btcd/wire"
)

// RegisterSubnetworkForTest is used to register network on DAG with specified gas limit
func RegisterSubnetworkForTest(dag *blockdag.BlockDAG, params *dagconfig.Params, gasLimit uint64) (*subnetworkid.SubnetworkID, error) {
	buildNextBlock := func(parentHashes []*daghash.Hash, txs []*wire.MsgTx) (*util.Block, error) {
		msgBlock, err := mining.PrepareBlockForTest(dag, params, parentHashes, txs, false, 1)
		if err != nil {
			return nil, err
		}

		return util.NewBlock(msgBlock), nil
	}

	addBlockToDAG := func(block *util.Block) error {
		isOrphan, err := dag.ProcessBlock(block, blockdag.BFNoPoWCheck)
		if err != nil {
			return err
		}

		if isOrphan {
			return fmt.Errorf("ProcessBlock: unexpected returned orphan block")
		}

		return nil
	}

	// Create a block in order to fund later transactions
	fundsBlock, err := buildNextBlock(dag.TipHashes(), []*wire.MsgTx{})
	if err != nil {
		return nil, fmt.Errorf("could not build funds block: %s", err)
	}

	err = addBlockToDAG(fundsBlock)
	if err != nil {
		return nil, fmt.Errorf("could not add funds block to DAG: %s", err)
	}

	fundsBlockCbTx := fundsBlock.Transactions()[0].MsgTx()

	// Create a block with a valid subnetwork registry transaction
	txIn := &wire.TxIn{
		PreviousOutPoint: *wire.NewOutPoint(fundsBlockCbTx.TxID(), 0),
		Sequence:         wire.MaxTxInSequenceNum,
	}
	txOut := &wire.TxOut{
		PkScript: blockdag.OpTrueScript,
		Value:    fundsBlockCbTx.TxOut[0].Value,
	}
	registryTx := wire.NewRegistryMsgTx(1, []*wire.TxIn{txIn}, []*wire.TxOut{txOut}, gasLimit)

	// Add it to the DAG
	registryBlock, err := buildNextBlock([]*daghash.Hash{fundsBlock.Hash()}, []*wire.MsgTx{registryTx})
	if err != nil {
		return nil, fmt.Errorf("could not build registry block: %s", err)
	}
	err = addBlockToDAG(registryBlock)
	if err != nil {
		return nil, fmt.Errorf("could not add registry block to DAG: %s", err)
	}

	// Build a subnetwork ID from the registry transaction
	subnetworkID, err := blockdag.TxToSubnetworkID(registryTx)
	if err != nil {
		return nil, fmt.Errorf("could not build subnetwork ID: %s", err)
	}
	return subnetworkID, nil
}
