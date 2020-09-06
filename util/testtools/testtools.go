package testtools

import (
	"time"

	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/domain/mining"
	"github.com/kaspanet/kaspad/util/daghash"

	"github.com/kaspanet/kaspad/domain/blockdag"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/txscript"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/subnetworkid"
)

// RegisterSubnetworkForTest is used to register network on DAG with specified gas limit
func RegisterSubnetworkForTest(dag *blockdag.BlockDAG, params *dagconfig.Params, gasLimit uint64) (*subnetworkid.SubnetworkID, error) {
	buildNextBlock := func(parentHashes []*daghash.Hash, txs []*appmessage.MsgTx) (*util.Block, error) {
		msgBlock, err := mining.PrepareBlockForTest(dag, parentHashes, txs, false)
		if err != nil {
			return nil, err
		}

		return util.NewBlock(msgBlock), nil
	}

	addBlockToDAG := func(block *util.Block) error {
		isOrphan, isDelayed, err := dag.ProcessBlock(block, blockdag.BFNoPoWCheck)
		if err != nil {
			return err
		}

		if isDelayed {
			return errors.Errorf("ProcessBlock: block is is too far in the future")
		}

		if isOrphan {
			return errors.Errorf("ProcessBlock: unexpected returned orphan block")
		}

		return nil
	}

	// Create a block in order to fund later transactions
	fundsBlock, err := buildNextBlock(dag.VirtualParentHashes(), []*appmessage.MsgTx{})
	if err != nil {
		return nil, errors.Errorf("could not build funds block: %s", err)
	}

	err = addBlockToDAG(fundsBlock)
	if err != nil {
		return nil, errors.Errorf("could not add funds block to DAG: %s", err)
	}

	fundsBlockCbTx := fundsBlock.Transactions()[0].MsgTx()

	// Create a block with a valid subnetwork registry transaction
	signatureScript, err := txscript.PayToScriptHashSignatureScript(blockdag.OpTrueScript, nil)
	if err != nil {
		return nil, errors.Errorf("Failed to build signature script: %s", err)
	}
	txIn := &appmessage.TxIn{
		PreviousOutpoint: *appmessage.NewOutpoint(fundsBlockCbTx.TxID(), 0),
		Sequence:         appmessage.MaxTxInSequenceNum,
		SignatureScript:  signatureScript,
	}

	scriptPubKey, err := txscript.PayToScriptHashScript(blockdag.OpTrueScript)
	if err != nil {
		return nil, err
	}
	txOut := &appmessage.TxOut{
		ScriptPubKey: scriptPubKey,
		Value:        fundsBlockCbTx.TxOut[0].Value,
	}
	registryTx := appmessage.NewRegistryMsgTx(1, []*appmessage.TxIn{txIn}, []*appmessage.TxOut{txOut}, gasLimit)

	// Add it to the DAG
	registryBlock, err := buildNextBlock([]*daghash.Hash{fundsBlock.Hash()}, []*appmessage.MsgTx{registryTx})
	if err != nil {
		return nil, errors.Errorf("could not build registry block: %s", err)
	}
	err = addBlockToDAG(registryBlock)
	if err != nil {
		return nil, errors.Errorf("could not add registry block to DAG: %s", err)
	}

	// Build a subnetwork ID from the registry transaction
	subnetworkID, err := blockdag.TxToSubnetworkID(registryTx)
	if err != nil {
		return nil, errors.Errorf("could not build subnetwork ID: %s", err)
	}
	return subnetworkID, nil
}

// WaitTillAllCompleteOrTimeout waits until all the provided channels has been written to,
// or until a timeout period has passed.
// Returns true iff all channels returned in the allotted time.
func WaitTillAllCompleteOrTimeout(timeoutDuration time.Duration, chans ...chan struct{}) (ok bool) {
	timeout := time.After(timeoutDuration)

	for _, c := range chans {
		select {
		case <-c:
			continue
		case <-timeout:
			return false
		}
	}

	return true
}
