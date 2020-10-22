package validator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"sort"
)

// ValidateAgainstPastUTXO validates the block against the UTXO of its past
func (v *validator) ValidateAgainstPastUTXO(block *externalapi.DomainBlock) error {
	acceptanceData, multiset := v.consensusStateManager.CalculateAcceptanceDataAndMultiset(block)

	err := v.validateAcceptedIDMerkleRoot(block, acceptanceData)
	if err != nil {
		return err
	}

	err = v.validateAcceptedIDMerkleRoot(block, acceptanceData)
	if err != nil {
		return err
	}

	err = v.validateUTXOCommitment(block, multiset)
	if err != nil {
		return err
	}

	err = v.validateCoinbaseTransaction()
	if err != nil {
		return err
	}

	return nil
}

func (v *validator) validateAcceptedIDMerkleRoot(block *externalapi.DomainBlock, consensusStateChanges model.ConsensusStateChanges) error {
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

func (v *validator) calculateAcceptedIDMerkleRoot(acceptanceData *model.BlockAcceptanceData) *daghash.Hash {
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

func (v *validator) validateUTXOCommitment(block *externalapi.DomainBlock, multiset model.Multiset) error {
	calculatedMultisetHash := multiset.Hash()
	if *calculatedMultisetHash != block.Header.UTXOCommitment {
		return ruleerrors.Errorf(ruleerrors.ErrBadUTXOCommitment, "UTXO commitment is invalid - block "+
			"header indicates %s, but calculated value is %s", block.Header.UTXOCommitment, calculatedMultisetHash)
	}

	return nil
}

// The following functions deal with building and validating the coinbase transaction
func (v *validator) validateCoinbaseTransaction(dag *BlockDAG, block *util.Block, txsAcceptanceData MultiBlockTxsAcceptanceData) error {
	panic("unimplemented")
}
