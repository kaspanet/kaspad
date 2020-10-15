package validator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashserialization"
	"github.com/kaspanet/kaspad/domain/consensus/utils/merkle"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
)

// ValidateBodyInIsolation validates block bodies in isolation from the current
// consensus state
func (bv *Validator) ValidateBodyInIsolation(block *model.DomainBlock) error {
	err := bv.checkBlockSize(block)
	if err != nil {
		return err
	}

	err = bv.checkBlockContainsAtLeastOneTransaction(block)
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

	err = bv.checkBlockHashMerkleRoot(block)
	if err != nil {
		return err
	}

	err = bv.checkBlockDuplicateTransactions(block)
	if err != nil {
		return err
	}

	err = bv.checkBlockDoubleSpends(block)
	if err != nil {
		return err
	}

	err = bv.checkBlockHasNoChainedTransactions(block)
	if err != nil {
		return err
	}

	err = bv.validateGasLimit(block)
	if err != nil {
		return err
	}

	return nil
}

func (bv *Validator) checkBlockContainsAtLeastOneTransaction(block *model.DomainBlock) error {
	if len(block.Transactions) == 0 {
		return ruleerrors.Errorf(ruleerrors.ErrNoTransactions, "block does not contain "+
			"any transactions")
	}
	return nil
}

func (bv *Validator) checkFirstBlockTransactionIsCoinbase(block *model.DomainBlock) error {
	if !transactionhelper.IsCoinBase(block.Transactions[transactionhelper.CoinbaseTransactionIndex]) {
		return ruleerrors.Errorf(ruleerrors.ErrFirstTxNotCoinbase, "first transaction in "+
			"block is not a coinbase")
	}
	return nil
}

func (bv *Validator) checkBlockContainsOnlyOneCoinbase(block *model.DomainBlock) error {
	for i, tx := range block.Transactions[transactionhelper.CoinbaseTransactionIndex+1:] {
		if transactionhelper.IsCoinBase(tx) {
			return ruleerrors.Errorf(ruleerrors.ErrMultipleCoinbases, "block contains second coinbase at "+
				"index %d", i+transactionhelper.CoinbaseTransactionIndex+1)
		}
	}
	return nil
}

func (bv *Validator) checkBlockTransactionOrder(block *model.DomainBlock) error {
	for i, tx := range block.Transactions[util.CoinbaseTransactionIndex+1:] {
		if i != 0 && subnetworks.Less(tx.SubnetworkID, block.Transactions[i].SubnetworkID) {
			return ruleerrors.Errorf(ruleerrors.ErrTransactionsNotSorted, "transactions must be sorted by subnetwork")
		}
	}
	return nil
}

func (bv *Validator) checkNoNonNativeTransactions(block *model.DomainBlock) error {
	// Disallow non-native/coinbase subnetworks in networks that don't allow them
	if !bv.enableNonNativeSubnetworks {
		for _, tx := range block.Transactions {
			if !(tx.SubnetworkID == subnetworks.SubnetworkIDNative ||
				tx.SubnetworkID == subnetworks.SubnetworkIDCoinbase) {
				return ruleerrors.Errorf(ruleerrors.ErrInvalidSubnetwork, "non-native/coinbase subnetworks are not allowed")
			}
		}
	}
	return nil
}

func (bv *Validator) checkTransactionsInIsolation(block *model.DomainBlock) error {
	for _, tx := range block.Transactions {
		err := bv.checkTransactionInIsolation(tx)
		if err != nil {
			return errors.Wrapf(err, "transaction %s failed isolation "+
				"check", hashserialization.TransactionID(tx))
		}
	}

	return nil
}

func (bv *Validator) checkBlockHashMerkleRoot(block *model.DomainBlock) error {
	calculatedHashMerkleRoot := merkle.CalcHashMerkleRoot(block.Transactions)
	if *block.Header.HashMerkleRoot != *calculatedHashMerkleRoot {
		return ruleerrors.Errorf(ruleerrors.ErrBadMerkleRoot, "block hash merkle root is invalid - block "+
			"header indicates %s, but calculated value is %s",
			block.Header.HashMerkleRoot, calculatedHashMerkleRoot)
	}
	return nil
}

func (bv *Validator) checkBlockDuplicateTransactions(block *model.DomainBlock) error {
	existingTxIDs := make(map[model.DomainTransactionID]struct{})
	for _, tx := range block.Transactions {
		id := hashserialization.TransactionID(tx)
		if _, exists := existingTxIDs[*id]; exists {
			return ruleerrors.Errorf(ruleerrors.ErrDuplicateTx, "block contains duplicate "+
				"transaction %s", id)
		}
		existingTxIDs[*id] = struct{}{}
	}
	return nil
}

func (bv *Validator) checkBlockDoubleSpends(block *model.DomainBlock) error {
	usedOutpoints := make(map[model.DomainOutpoint]*model.DomainTransactionID)
	for _, tx := range block.Transactions {
		for _, input := range tx.Inputs {
			txID := hashserialization.TransactionID(tx)
			if spendingTxID, exists := usedOutpoints[input.PreviousOutpoint]; exists {
				return ruleerrors.Errorf(ruleerrors.ErrDoubleSpendInSameBlock, "transaction %s spends "+
					"outpoint %s that was already spent by "+
					"transaction %s in this block", txID, input.PreviousOutpoint, spendingTxID)
			}
			usedOutpoints[input.PreviousOutpoint] = txID
		}
	}
	return nil
}

func (bv *Validator) checkBlockHasNoChainedTransactions(block *model.DomainBlock) error {

	transactions := block.Transactions
	transactionsSet := make(map[model.DomainTransactionID]struct{}, len(transactions))
	for _, transaction := range transactions {
		txID := hashserialization.TransactionID(transaction)
		transactionsSet[*txID] = struct{}{}
	}

	for _, transaction := range transactions {
		for i, transactionInput := range transaction.Inputs {
			if _, ok := transactionsSet[transactionInput.PreviousOutpoint.ID]; ok {
				txID := hashserialization.TransactionID(transaction)
				return ruleerrors.Errorf(ruleerrors.ErrChainedTransactions, "block contains chained "+
					"transactions: Input %d of transaction %s spend "+
					"an output of transaction %s", i, txID, transactionInput.PreviousOutpoint.ID)
			}
		}
	}

	return nil
}

func (bv *Validator) validateGasLimit(block *model.DomainBlock) error {
	// TODO: implement this
	return nil
}

func (bv *Validator) checkBlockSize(block *model.DomainBlock) error {
	size := uint64(0)
	size += bv.headerEstimatedSerializedSize(block.Header)

	for _, tx := range block.Transactions {
		sizeBefore := size
		size += bv.transactionEstimatedSerializedSize(tx)
		const maxBlockSize = 1_000_000
		if size > maxBlockSize || size < sizeBefore {
			return ruleerrors.Errorf(ruleerrors.ErrBlockSizeTooHigh, "block excceeded the size limit of %d", maxBlockSize)
		}
	}

	return nil
}
