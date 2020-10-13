package validator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"sort"
)

// ValidateAgainstPastUTXO validates the block against the UTXO of its past
func (bv *Validator) ValidateAgainstPastUTXO(block *model.DomainBlock) error {
	consensusStateChanges := bv.consensusStateManager.CalculateConsensusStateChanges(block)
	return nil
}

func (bv *Validator) validateAcceptedIDMerkleRoot(block *model.DomainBlock, consensusStateChanges model.ConsensusStateChanges) error {
	// Genesis block doesn't have acceptance data to validate
	if len(block.Header.ParentHashes) == 0 {
		return nil
	}

	calculatedAccepetedIDMerkleRoot := calculateAcceptedIDMerkleRoot(txsAcceptanceData)
	header := node.Header()
	if !header.AcceptedIDMerkleRoot.IsEqual(calculatedAccepetedIDMerkleRoot) {
		str := fmt.Sprintf("block accepted ID merkle root is invalid - block "+
			"header indicates %s, but calculated value is %s",
			header.AcceptedIDMerkleRoot, calculatedAccepetedIDMerkleRoot)
		return ruleError(ErrBadMerkleRoot, str)
	}
	return nil
}

func (bv *Validator) calculateAcceptedIDMerkleRoot(acceptanceData *model.BlockAcceptanceData) *daghash.Hash {
	var acceptedTxs []*util.Tx
	for _, blockTxsAcceptanceData := range multiBlockTxsAcceptanceData {
		for _, txAcceptance := range blockTxsAcceptanceData.TxAcceptanceData {
			if !txAcceptance.IsAccepted {
				continue
			}
			acceptedTxs = append(acceptedTxs, txAcceptance.Tx)
		}
	}
	sort.Slice(acceptedTxs, func(i, j int) bool {
		return daghash.LessTxID(acceptedTxs[i].ID(), acceptedTxs[j].ID())
	})

	acceptedIDMerkleTree := BuildIDMerkleTreeStore(acceptedTxs)
	return acceptedIDMerkleTree.Root()
}
