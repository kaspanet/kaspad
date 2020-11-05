package blockvalidator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensusserialization"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/estimatedsize"
	"github.com/kaspanet/kaspad/domain/consensus/utils/merkle"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
)

// ValidateBodyInIsolation validates block bodies in isolation from the current
// consensus state
func (v *blockValidator) ValidateBodyInIsolation(blockHash *externalapi.DomainHash) error {
	block, err := v.blockStore.Block(v.databaseContext, blockHash)
	if err != nil {
		return err
	}

	err = v.checkBlockSize(block)
	if err != nil {
		return err
	}

	err = v.checkBlockContainsAtLeastOneTransaction(block)
	if err != nil {
		return err
	}

	err = v.checkFirstBlockTransactionIsCoinbase(block)
	if err != nil {
		return err
	}

	err = v.checkBlockContainsOnlyOneCoinbase(block)
	if err != nil {
		return err
	}

	err = v.checkCoinbase(block)
	if err != nil {
		return err
	}

	err = v.checkTransactionsInIsolation(block)
	if err != nil {
		return err
	}

	err = v.checkBlockHashMerkleRoot(block)
	if err != nil {
		return err
	}

	err = v.checkBlockDuplicateTransactions(block)
	if err != nil {
		return err
	}

	err = v.checkBlockDoubleSpends(block)
	if err != nil {
		return err
	}

	err = v.checkBlockHasNoChainedTransactions(block)
	if err != nil {
		return err
	}

	err = v.validateGasLimit(block)
	if err != nil {
		return err
	}

	return nil
}

func (v *blockValidator) checkCoinbase(block *externalapi.DomainBlock) error {
	_, _, err := v.coinbaseManager.ExtractCoinbaseDataAndBlueScore(block.
		Transactions[transactionhelper.CoinbaseTransactionIndex])
	if err != nil {
		return err
	}
	return nil
}

func (v *blockValidator) checkBlockContainsAtLeastOneTransaction(block *externalapi.DomainBlock) error {
	if len(block.Transactions) == 0 {
		return errors.Wrapf(ruleerrors.ErrNoTransactions, "block does not contain "+
			"any transactions")
	}
	return nil
}

func (v *blockValidator) checkFirstBlockTransactionIsCoinbase(block *externalapi.DomainBlock) error {
	if !transactionhelper.IsCoinBase(block.Transactions[transactionhelper.CoinbaseTransactionIndex]) {
		return errors.Wrapf(ruleerrors.ErrFirstTxNotCoinbase, "first transaction in "+
			"block is not a coinbase")
	}
	return nil
}

func (v *blockValidator) checkBlockContainsOnlyOneCoinbase(block *externalapi.DomainBlock) error {
	for i, tx := range block.Transactions[transactionhelper.CoinbaseTransactionIndex+1:] {
		if transactionhelper.IsCoinBase(tx) {
			return errors.Wrapf(ruleerrors.ErrMultipleCoinbases, "block contains second coinbase at "+
				"index %d", i+transactionhelper.CoinbaseTransactionIndex+1)
		}
	}
	return nil
}

func (v *blockValidator) checkBlockTransactionOrder(block *externalapi.DomainBlock) error {
	for i, tx := range block.Transactions[util.CoinbaseTransactionIndex+1:] {
		if i != 0 && subnetworks.Less(tx.SubnetworkID, block.Transactions[i].SubnetworkID) {
			return errors.Wrapf(ruleerrors.ErrTransactionsNotSorted, "transactions must be sorted by subnetwork")
		}
	}
	return nil
}

func (v *blockValidator) checkNoNonNativeTransactions(block *externalapi.DomainBlock) error {
	// Disallow non-native/coinbase subnetworks in networks that don't allow them
	if !v.enableNonNativeSubnetworks {
		for _, tx := range block.Transactions {
			if !(tx.SubnetworkID == subnetworks.SubnetworkIDNative ||
				tx.SubnetworkID == subnetworks.SubnetworkIDCoinbase) {
				return errors.Wrapf(ruleerrors.ErrInvalidSubnetwork, "non-native/coinbase subnetworks are not allowed")
			}
		}
	}
	return nil
}

func (v *blockValidator) checkTransactionsInIsolation(block *externalapi.DomainBlock) error {
	for _, tx := range block.Transactions {
		err := v.transactionValidator.ValidateTransactionInIsolation(tx)
		if err != nil {
			return errors.Wrapf(err, "transaction %s failed isolation "+
				"check", consensusserialization.TransactionID(tx))
		}
	}

	return nil
}

func (v *blockValidator) checkBlockHashMerkleRoot(block *externalapi.DomainBlock) error {
	calculatedHashMerkleRoot := merkle.CalculateHashMerkleRoot(block.Transactions)
	if block.Header.HashMerkleRoot != *calculatedHashMerkleRoot {
		return errors.Wrapf(ruleerrors.ErrBadMerkleRoot, "block hash merkle root is invalid - block "+
			"header indicates %s, but calculated value is %s",
			block.Header.HashMerkleRoot, calculatedHashMerkleRoot)
	}
	return nil
}

func (v *blockValidator) checkBlockDuplicateTransactions(block *externalapi.DomainBlock) error {
	existingTxIDs := make(map[externalapi.DomainTransactionID]struct{})
	for _, tx := range block.Transactions {
		id := consensusserialization.TransactionID(tx)
		if _, exists := existingTxIDs[*id]; exists {
			return errors.Wrapf(ruleerrors.ErrDuplicateTx, "block contains duplicate "+
				"transaction %s", id)
		}
		existingTxIDs[*id] = struct{}{}
	}
	return nil
}

func (v *blockValidator) checkBlockDoubleSpends(block *externalapi.DomainBlock) error {
	usedOutpoints := make(map[externalapi.DomainOutpoint]*externalapi.DomainTransactionID)
	for _, tx := range block.Transactions {
		for _, input := range tx.Inputs {
			txID := consensusserialization.TransactionID(tx)
			if spendingTxID, exists := usedOutpoints[input.PreviousOutpoint]; exists {
				return errors.Wrapf(ruleerrors.ErrDoubleSpendInSameBlock, "transaction %s spends "+
					"outpoint %s that was already spent by "+
					"transaction %s in this block", txID,
					input.PreviousOutpoint, spendingTxID)
			}
			usedOutpoints[input.PreviousOutpoint] = txID
		}
	}
	return nil
}

func (v *blockValidator) checkBlockHasNoChainedTransactions(block *externalapi.DomainBlock) error {

	transactions := block.Transactions
	transactionsSet := make(map[externalapi.DomainTransactionID]struct{}, len(transactions))
	for _, transaction := range transactions {
		txID := consensusserialization.TransactionID(transaction)
		transactionsSet[*txID] = struct{}{}
	}

	for _, transaction := range transactions {
		for i, transactionInput := range transaction.Inputs {
			if _, ok := transactionsSet[transactionInput.PreviousOutpoint.TransactionID]; ok {
				txID := consensusserialization.TransactionID(transaction)
				return errors.Wrapf(ruleerrors.ErrChainedTransactions, "block contains chained "+
					"transactions: Input %d of transaction %s spend "+
					"an output of transaction %s", i, txID, transactionInput.PreviousOutpoint.TransactionID)
			}
		}
	}

	return nil
}

func (v *blockValidator) validateGasLimit(block *externalapi.DomainBlock) error {
	// TODO: implement this
	return nil
}

func (v *blockValidator) checkBlockSize(block *externalapi.DomainBlock) error {
	size := uint64(0)
	size += v.headerEstimatedSerializedSize(block.Header)

	for _, tx := range block.Transactions {
		sizeBefore := size
		size += estimatedsize.TransactionEstimatedSerializedSize(tx)
		if size > constants.MaxBlockSize || size < sizeBefore {
			return errors.Wrapf(ruleerrors.ErrBlockSizeTooHigh, "block excceeded the size limit of %d",
				constants.MaxBlockSize)
		}
	}

	return nil
}
