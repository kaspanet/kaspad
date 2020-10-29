package serialization

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// DomainBlockToDbBlock converts DomainBlocks to DbBlock
func DomainBlockToDbBlock(domainBlock *externalapi.DomainBlock) *DbBlock {
	dbTransactions := make([]*DbTransaction, len(domainBlock.Transactions))
	for i, domainTransaction := range domainBlock.Transactions {
		dbTransactions[i] = DomainTransactionToDbTransaction(domainTransaction)
	}

	return &DbBlock{
		Header:       DomainBlockHeaderToDbBlockHeader(domainBlock.Header),
		Transactions: dbTransactions,
	}
}

// DbBlockToDomainBlock converts DbBlock to DomainBlock
func DbBlockToDomainBlock(dbBlock *DbBlock) (*externalapi.DomainBlock, error) {
	domainBlockHeader, err := DbBlockHeaderToDomainBlockHeader(dbBlock.Header)
	if err != nil {
		return nil, err
	}

	domainTransactions := make([]*externalapi.DomainTransaction, len(dbBlock.Transactions))
	for i, dbTransaction := range dbBlock.Transactions {
		var err error
		domainTransactions[i], err = DbTransactionToDomainTransaction(dbTransaction)
		if err != nil {
			return nil, err
		}
	}

	return &externalapi.DomainBlock{
		Header:       domainBlockHeader,
		Transactions: domainTransactions,
	}, nil
}
