package consensusstatemanager

import (
	"sort"

	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"

	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"

	"github.com/kaspanet/kaspad/domain/consensus/utils/merkle"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/pkg/errors"
)

func (csm *consensusStateManager) verifyUTXO(stagingArea *model.StagingArea, block *externalapi.DomainBlock,
	blockHash *externalapi.DomainHash, pastUTXODiff externalapi.UTXODiff, acceptanceData externalapi.AcceptanceData,
	multiset model.Multiset) error {

	log.Tracef("verifyUTXO start for block %s", blockHash)
	defer log.Tracef("verifyUTXO end for block %s", blockHash)

	log.Debugf("Validating UTXO commitment for block %s", blockHash)
	err := csm.validateUTXOCommitment(block, blockHash, multiset)
	if err != nil {
		return err
	}
	log.Debugf("UTXO commitment validation passed for block %s", blockHash)

	log.Debugf("Validating acceptedIDMerkleRoot for block %s", blockHash)
	err = csm.validateAcceptedIDMerkleRoot(block, blockHash, acceptanceData)
	if err != nil {
		return err
	}
	log.Debugf("AcceptedIDMerkleRoot validation passed for block %s", blockHash)

	coinbaseTransaction := block.Transactions[0]
	log.Debugf("Validating coinbase transaction %s for block %s",
		consensushashing.TransactionID(coinbaseTransaction), blockHash)
	err = csm.validateCoinbaseTransaction(stagingArea, blockHash, coinbaseTransaction)
	if err != nil {
		return err
	}
	log.Debugf("Coinbase transaction validation passed for block %s", blockHash)

	log.Debugf("Validating transactions against past UTXO for block %s", blockHash)
	err = csm.validateBlockTransactionsAgainstPastUTXO(stagingArea, block, pastUTXODiff)
	if err != nil {
		return err
	}
	log.Tracef("Transactions against past UTXO validation passed for block %s", blockHash)

	return nil
}

func (csm *consensusStateManager) validateBlockTransactionsAgainstPastUTXO(stagingArea *model.StagingArea,
	block *externalapi.DomainBlock, pastUTXODiff externalapi.UTXODiff) error {

	blockHash := consensushashing.BlockHash(block)
	log.Tracef("validateBlockTransactionsAgainstPastUTXO start for block %s", blockHash)
	defer log.Tracef("validateBlockTransactionsAgainstPastUTXO end for block %s", blockHash)

	selectedParentMedianTime, err := csm.pastMedianTimeManager.PastMedianTime(stagingArea, blockHash)
	if err != nil {
		return err
	}
	log.Tracef("The past median time of %s is %d", blockHash, selectedParentMedianTime)

	for i, transaction := range block.Transactions {
		transactionID := consensushashing.TransactionID(transaction)
		log.Tracef("Validating transaction %s in block %s against "+
			"the block's past UTXO", transactionID, blockHash)
		if i == transactionhelper.CoinbaseTransactionIndex {
			log.Tracef("Skipping transaction %s because it is the coinbase", transactionID)
			continue
		}

		log.Tracef("Populating transaction %s with UTXO entries", transactionID)
		err = csm.populateTransactionWithUTXOEntriesFromVirtualOrDiff(stagingArea, transaction, pastUTXODiff)
		if err != nil {
			return err
		}

		log.Tracef("Validating transaction %s and populating it with fee", transactionID)
		err = csm.transactionValidator.ValidateTransactionInContextAndPopulateFee(
			stagingArea, transaction, blockHash)
		if err != nil {
			return err
		}
		log.Tracef("Validation against the block's past UTXO "+
			"passed for transaction %s in block %s", transactionID, blockHash)
	}
	return nil
}

func (csm *consensusStateManager) validateAcceptedIDMerkleRoot(block *externalapi.DomainBlock,
	blockHash *externalapi.DomainHash, acceptanceData externalapi.AcceptanceData) error {

	log.Tracef("validateAcceptedIDMerkleRoot start for block %s", blockHash)
	defer log.Tracef("validateAcceptedIDMerkleRoot end for block %s", blockHash)

	calculatedAcceptedIDMerkleRoot := calculateAcceptedIDMerkleRoot(acceptanceData)
	if !block.Header.AcceptedIDMerkleRoot().Equal(calculatedAcceptedIDMerkleRoot) {
		return errors.Wrapf(ruleerrors.ErrBadMerkleRoot, "block %s accepted ID merkle root is invalid - block "+
			"header indicates %s, but calculated value is %s",
			blockHash, block.Header.UTXOCommitment(), calculatedAcceptedIDMerkleRoot)
	}

	return nil
}

func (csm *consensusStateManager) validateUTXOCommitment(
	block *externalapi.DomainBlock, blockHash *externalapi.DomainHash, multiset model.Multiset) error {

	log.Tracef("validateUTXOCommitment start for block %s", blockHash)
	defer log.Tracef("validateUTXOCommitment end for block %s", blockHash)

	if blockHash.Equal(csm.genesisHash) {
		return nil
	}

	multisetHash := multiset.Hash()
	if !block.Header.UTXOCommitment().Equal(multisetHash) {
		return errors.Wrapf(ruleerrors.ErrBadUTXOCommitment, "block %s UTXO commitment is invalid - block "+
			"header indicates %s, but calculated value is %s", blockHash, block.Header.UTXOCommitment(), multisetHash)
	}

	return nil
}

func calculateAcceptedIDMerkleRoot(multiblockAcceptanceData externalapi.AcceptanceData) *externalapi.DomainHash {
	log.Tracef("calculateAcceptedIDMerkleRoot start")
	defer log.Tracef("calculateAcceptedIDMerkleRoot end")

	var acceptedTransactions []*externalapi.DomainTransaction

	for _, blockAcceptanceData := range multiblockAcceptanceData {
		for _, transactionAcceptance := range blockAcceptanceData.TransactionAcceptanceData {
			if !transactionAcceptance.IsAccepted {
				continue
			}
			acceptedTransactions = append(acceptedTransactions, transactionAcceptance.Transaction)
		}
	}
	sort.Slice(acceptedTransactions, func(i, j int) bool {
		return consensushashing.TransactionID(acceptedTransactions[i]).Less(
			consensushashing.TransactionID(acceptedTransactions[j]))
	})

	return merkle.CalculateIDMerkleRoot(acceptedTransactions)
}

func (csm *consensusStateManager) validateCoinbaseTransaction(stagingArea *model.StagingArea,
	blockHash *externalapi.DomainHash, coinbaseTransaction *externalapi.DomainTransaction) error {

	log.Tracef("validateCoinbaseTransaction start for block %s", blockHash)
	defer log.Tracef("validateCoinbaseTransaction end for block %s", blockHash)

	log.Tracef("Extracting coinbase data for coinbase transaction %s in block %s",
		consensushashing.TransactionID(coinbaseTransaction), blockHash)
	_, coinbaseData, _, err := csm.coinbaseManager.ExtractCoinbaseDataBlueScoreAndSubsidy(coinbaseTransaction)
	if err != nil {
		return err
	}

	log.Tracef("Calculating the expected coinbase transaction for the given coinbase data and block %s", blockHash)
	expectedCoinbaseTransaction, _, err :=
		csm.coinbaseManager.ExpectedCoinbaseTransaction(stagingArea, blockHash, coinbaseData)
	if err != nil {
		return err
	}

	coinbaseTransactionHash := consensushashing.TransactionHash(coinbaseTransaction)
	expectedCoinbaseTransactionHash := consensushashing.TransactionHash(expectedCoinbaseTransaction)
	log.Tracef("given coinbase hash: %s, expected coinbase hash: %s",
		coinbaseTransactionHash, expectedCoinbaseTransactionHash)

	if !coinbaseTransactionHash.Equal(expectedCoinbaseTransactionHash) {
		return errors.Wrap(ruleerrors.ErrBadCoinbaseTransaction, "coinbase transaction is not built as expected")
	}

	return nil
}
