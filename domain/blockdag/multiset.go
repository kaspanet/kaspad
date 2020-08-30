package blockdag

import (
	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

// calcMultiset returns the multiset of the past UTXO of the given block.
func (node *blockNode) calcMultiset(acceptanceData MultiBlockTxsAcceptanceData,
	selectedParentPastUTXO UTXOSet) (*secp256k1.MultiSet, error) {

	ms, err := node.selectedParentMultiset()
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
			ms, err = addTxToMultiset(ms, tx, selectedParentPastUTXO, node.blueScore)
			if err != nil {
				return nil, err
			}
		}
	}
	return ms, nil
}

// selectedParentMultiset returns the multiset of the node's selected
// parent. If the node is the genesis blockNode then it does not have
// a selected parent, in which case return a new, empty multiset.
func (node *blockNode) selectedParentMultiset() (*secp256k1.MultiSet, error) {
	if node.isGenesis() {
		return secp256k1.NewMultiset(), nil
	}

	ms, err := node.dag.multisetStore.multisetByBlockNode(node.selectedParent)
	if err != nil {
		return nil, err
	}

	return ms, nil
}

func addTxToMultiset(ms *secp256k1.MultiSet, tx *appmessage.MsgTx, pastUTXO UTXOSet, blockBlueScore uint64) (*secp256k1.MultiSet, error) {
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
		entry := NewUTXOEntry(txOut, isCoinbase, blockBlueScore)

		var err error
		ms, err = addUTXOToMultiset(ms, entry, &outpoint)
		if err != nil {
			return nil, err
		}
	}
	return ms, nil
}
