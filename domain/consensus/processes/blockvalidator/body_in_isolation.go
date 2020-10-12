package blockvalidator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactions"
	"github.com/kaspanet/kaspad/util"
)

// ValidateBodyInIsolation validates block bodies in isolation from the current
// consensus state
func (bv *BlockValidator) ValidateBodyInIsolation(block *model.DomainBlock) error {
	err := bv.checkBlockContainsAtLeastOneTransaction(block)
	if err != nil {
		return err
	}

	err = bv.checkFirstBlockTransactionIsCoinbase(block)
	if err != nil {
		return err
	}

	err = bv.checkBlockContainsOnlyOneCoinbase(block)
	if err != nil {
		return err
	}

	err = bv.checkTransactionsInIsolation(block)
	if err != nil {
		return err
	}

	return nil
}

func (bv *BlockValidator) checkBlockContainsAtLeastOneTransaction(block *model.DomainBlock) error {
	if len(block.Transactions) == 0 {
		return ruleerrors.Errorf(ruleerrors.ErrNoTransactions, "block does not contain "+
			"any transactions")
	}
	return nil
}

func (bv *BlockValidator) checkFirstBlockTransactionIsCoinbase(block *model.DomainBlock) error {
	if !transactions.IsCoinBase(block.Transactions[transactions.CoinbaseTransactionIndex]) {
		return ruleerrors.Errorf(ruleerrors.ErrFirstTxNotCoinbase, "first transaction in "+
			"block is not a coinbase")
	}
	return nil
}

func (bv *BlockValidator) checkBlockContainsOnlyOneCoinbase(block *model.DomainBlock) error {
	for i, tx := range block.Transactions[transactions.CoinbaseTransactionIndex+1:] {
		if transactions.IsCoinBase(tx) {
			return ruleerrors.Errorf(ruleerrors.ErrMultipleCoinbases, "block contains second coinbase at "+
				"index %d", i+transactions.CoinbaseTransactionIndex+1)
		}
	}
	return nil
}

func (bv *BlockValidator) checkBlockTransactionOrder(block *model.DomainBlock) error {
	for i, tx := range block.Transactions[util.CoinbaseTransactionIndex+1:] {
		if i != 0 && subnetworks.Less(tx.SubnetworkID, block.Transactions[i].SubnetworkID) {
			return ruleerrors.Errorf(ruleerrors.ErrTransactionsNotSorted, "transactions must be sorted by subnetwork")
		}
	}
	return nil
}

func (bv *BlockValidator) checkNoNonNativeTransactions(block *model.DomainBlock) error {
	// Disallow non-native/coinbase subnetworks in networks that don't allow them
	if !bv.enableNonNativeSubnetworks {
		for _, tx := range block.Transactions {
			if !(*tx.SubnetworkID == *subnetworks.SubnetworkIDNative ||
				*tx.SubnetworkID == *subnetworks.SubnetworkIDCoinbase) {
				return ruleerrors.Errorf(ruleerrors.ErrInvalidSubnetwork, "non-native/coinbase subnetworks are not allowed")
			}
		}
	}
	return nil
}

func (bv *BlockValidator) checkTransactionsInIsolation(block *model.DomainBlock) error {
	// TODO implement this
	panic("unimplemented")
}

func (bv *BlockValidator) checkBlockHashMerkleRoot(block *util.Block) error {
	// Build merkle tree and ensure the calculated merkle root matches the
	// entry in the block header. This also has the effect of caching all
	// of the transaction hashes in the block to speed up future hash
	// checks.
	hashMerkleTree := BuildHashMerkleTreeStore(block.Transactions())
	calculatedHashMerkleRoot := hashMerkleTree.Root()
	if !block.MsgBlock().Header.HashMerkleRoot.IsEqual(calculatedHashMerkleRoot) {
		return ruleerrors.Errorf(ruleerrors.ErrBadMerkleRoot, "block hash merkle root is invalid - block "+
			"header indicates %s, but calculated value is %s",
			block.MsgBlock().Header.HashMerkleRoot, calculatedHashMerkleRoot)
	}
	return nil
}
