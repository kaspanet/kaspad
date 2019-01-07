package blockdag

import (
	"fmt"
	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/database"
	"github.com/daglabs/btcd/wire"
)

// validateAndExtractSubNetworkRegistryTxs filters the given input and extracts a list
// of valid sub-network registry transactions.
func validateAndExtractSubNetworkRegistryTxs(txs []*TxWithBlockHash) ([]*wire.MsgTx, error) {
	validSubNetworkRegistryTxs := make([]*wire.MsgTx, 0, len(txs))
	for _, txData := range txs {
		tx := txData.Tx.MsgTx()
		if tx.SubNetworkID == wire.SubNetworkRegistry {
			err := validateSubNetworkRegistryTransaction(tx)
			if err != nil {
				return nil, err
			}
			validSubNetworkRegistryTxs = append(validSubNetworkRegistryTxs, tx)
		}
	}

	return validSubNetworkRegistryTxs, nil
}

// validateSubNetworkRegistryTransaction makes sure that a given sub-network registry
// transaction is valid. Such a transaction is valid iff:
// - Its entire payload is a uint64 (8 bytes)
func validateSubNetworkRegistryTransaction(tx *wire.MsgTx) error {
	if len(tx.Payload) != 8 {
		return ruleError(ErrSubNetworkRegistry, fmt.Sprintf("validation failed: subnetwork registry"+
			"tx '%s' has an invalid payload", tx.TxHash()))
	}

	return nil
}

// registerPendingSubNetworks attempts to register all the pending sub-networks that
// had previously been defined between the initial finality point and the new one.
func (dag *BlockDAG) registerPendingSubNetworks(dbTx database.Tx, initialFinalityPoint *blockNode, newFinalityPoint *blockNode) error {
	var stack []*blockNode
	for currentNode := newFinalityPoint; currentNode != initialFinalityPoint; currentNode = currentNode.selectedParent {
		stack = append(stack, currentNode)
	}

	for i := len(stack) - 1; i >= 0; i-- {
		currentNode := stack[i]
		for _, blue := range currentNode.blues {
			err := dag.registerPendingSubNetworksInBlock(dbTx, blue.hash)
			if err != nil {
				return fmt.Errorf("failed to register pending sub-networks: %s", err)
			}
		}
		err := dag.registerPendingSubNetworksInBlock(dbTx, currentNode.hash)
		if err != nil {
			return fmt.Errorf("failed to register pending sub-networks: : %s", err)
		}
	}

	return nil
}

// registerPendingSubNetworksInBlock attempts to register all the sub-networks
// that had been defined in a given block.
func (dag *BlockDAG) registerPendingSubNetworksInBlock(dbTx database.Tx, blockHash daghash.Hash) error {
	pendingSubNetworkTxs, err := dbGetPendingSubNetworkTxs(dbTx, blockHash)
	if err != nil {
		return fmt.Errorf("failed to retrieve pending sub-network txs in block '%s': %s", blockHash, err)
	}
	for _, tx := range pendingSubNetworkTxs {
		if !dbIsRegisteredSubNetworkTx(dbTx, tx.TxHash()) {
			createdSubNetwork := newSubNetwork(tx)
			err := dbRegisterSubNetwork(dbTx, dag.lastSubNetworkID, createdSubNetwork)
			if err != nil {
				return fmt.Errorf("failed registering sub-network"+
					"for tx '%s' in block '%s': %s", tx.TxHash(), blockHash, err)
			}

			err = dbPutRegisteredSubNetworkTx(dbTx, tx.TxHash(), dag.lastSubNetworkID)
			if err != nil {
				return fmt.Errorf("failed to put registered sub-network tx '%s'"+
					" in block '%s': %s", tx.TxHash(), blockHash, err)
			}

			dag.lastSubNetworkID++
		}
	}

	err = dbRemovePendingSubNetworkTxs(dbTx, blockHash)
	if err != nil {
		return fmt.Errorf("failed to remove block '%s'"+
			"from pending sub-networks: %s", blockHash, err)
	}

	return nil
}

// subNetwork returns a registered-and-finalized sub-network. If the sub-network
// does not exist this method returns an error.
func (dag *BlockDAG) subNetwork(subNetworkID uint64) (*subNetwork, error) {
	var sNet *subNetwork
	var err error
	dbErr := dag.db.View(func(dbTx database.Tx) error {
		sNet, err = dbGetSubNetwork(dbTx, subNetworkID)
		return nil
	})
	if dbErr != nil {
		return nil, fmt.Errorf("could not retrieve sub-network '%d': %s", subNetworkID, dbErr)
	}
	if err != nil {
		return nil, fmt.Errorf("could not retrieve sub-network '%d': %s", subNetworkID, err)
	}

	return sNet, nil
}

// GasLimit returns the gas limit of a registered-and-finalized sub-network. If
// the sub-network does not exist this method returns an error.
func (dag *BlockDAG) GasLimit(subNetworkID uint64) (uint64, error) {
	sNet, err := dag.subNetwork(subNetworkID)
	if err != nil {
		return 0, err
	}

	return sNet.gasLimit, nil
}
