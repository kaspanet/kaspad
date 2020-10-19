package blockdag

import (
	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/blocknode"
	"github.com/kaspanet/kaspad/domain/utxo"
	"github.com/pkg/errors"
)

// calcMultiset returns the multiset of the past UTXO of the given block.
func (dag *BlockDAG) calcMultiset(node *blocknode.Node, acceptanceData MultiBlockTxsAcceptanceData,
	selectedParentPastUTXO utxo.Set) (*secp256k1.MultiSet, error) {

	ms, err := dag.selectedParentMultiset(node)
	if err != nil {
		return nil, err
	}

	for _, blockAcceptanceData := range acceptanceData {
		for _, txAcceptanceData := range blockAcceptanceData.TxAcceptanceData {
			if !txAcceptanceData.IsAccepted {
				continue
			}

			tx := txAcceptanceData.Tx.MsgTx()

			var err error
			ms, err = addTxToMultiset(ms, tx, selectedParentPastUTXO, node.BlueScore)
			if err != nil {
				return nil, err
			}
		}
	}
	return ms, nil
}

// selectedParentMultiset returns the multiset of the node's selected
// parent. If the node is the genesis Node then it does not have
// a selected parent, in which case return a new, empty multiset.
func (dag *BlockDAG) selectedParentMultiset(node *blocknode.Node) (*secp256k1.MultiSet, error) {
	if node.IsGenesis() {
		return secp256k1.NewMultiset(), nil
	}

	ms, err := dag.multisetStore.MultisetByBlockNode(node.SelectedParent)
	if err != nil {
		return nil, err
	}

	return ms, nil
}

func addTxToMultiset(ms *secp256k1.MultiSet, tx *appmessage.MsgTx, pastUTXO utxo.Set, blockBlueScore uint64) (*secp256k1.MultiSet, error) {
	for _, txIn := range tx.TxIn {
		entry, ok := pastUTXO.Get(txIn.PreviousOutpoint)
		if !ok {
			return nil, errors.Errorf("Couldn't find entry for outpoint %s", txIn.PreviousOutpoint)
		}

		var err error
		ms, err = removeUTXOFromMultiset(ms, entry, &txIn.PreviousOutpoint)
		if err != nil {
			return nil, err
		}
	}

	isCoinbase := tx.IsCoinBase()
	for i, txOut := range tx.TxOut {
		outpoint := *appmessage.NewOutpoint(tx.TxID(), uint32(i))
		entry := utxo.NewEntry(txOut, isCoinbase, blockBlueScore)

		var err error
		ms, err = addUTXOToMultiset(ms, entry, &outpoint)
		if err != nil {
			return nil, err
		}
	}
	return ms, nil
}
