package serialization

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// DomainAcceptanceDataToDbAcceptanceData converts model.AcceptanceData to DbAcceptanceData
func DomainAcceptanceDataToDbAcceptanceData(domainAcceptanceData externalapi.AcceptanceData) *DbAcceptanceData {
	dbBlockAcceptanceData := make([]*DbBlockAcceptanceData, len(domainAcceptanceData))
	for i, blockAcceptanceData := range domainAcceptanceData {
		dbTransactionAcceptanceData := make([]*DbTransactionAcceptanceData,
			len(blockAcceptanceData.TransactionAcceptanceData))

		for j, transactionAcceptanceData := range blockAcceptanceData.TransactionAcceptanceData {
			dbTransaction := DomainTransactionToDbTransaction(transactionAcceptanceData.Transaction)
			dbTransactionAcceptanceData[j] = &DbTransactionAcceptanceData{
				Transaction: dbTransaction,
				Fee:         transactionAcceptanceData.Fee,
				IsAccepted:  transactionAcceptanceData.IsAccepted,
			}
		}

		blockHash := DomainHashToDbHash(blockAcceptanceData.BlockHash)

		dbBlockAcceptanceData[i] = &DbBlockAcceptanceData{
			BlockHash:                 blockHash,
			TransactionAcceptanceData: dbTransactionAcceptanceData,
		}
	}

	return &DbAcceptanceData{
		BlockAcceptanceData: dbBlockAcceptanceData,
	}
}

// DbAcceptanceDataToDomainAcceptanceData converts DbAcceptanceData to model.AcceptanceData
func DbAcceptanceDataToDomainAcceptanceData(dbAcceptanceData *DbAcceptanceData) (externalapi.AcceptanceData, error) {
	domainAcceptanceData := make(externalapi.AcceptanceData, len(dbAcceptanceData.BlockAcceptanceData))
	for i, dbBlockAcceptanceData := range dbAcceptanceData.BlockAcceptanceData {
		domainTransactionAcceptanceData := make([]*externalapi.TransactionAcceptanceData,
			len(dbBlockAcceptanceData.TransactionAcceptanceData))

		for j, dbTransactionAcceptanceData := range dbBlockAcceptanceData.TransactionAcceptanceData {
			domainTransaction, err := DbTransactionToDomainTransaction(dbTransactionAcceptanceData.Transaction)
			if err != nil {
				return nil, err
			}
			domainTransactionAcceptanceData[j] = &externalapi.TransactionAcceptanceData{
				Transaction: domainTransaction,
				Fee:         dbTransactionAcceptanceData.Fee,
				IsAccepted:  dbTransactionAcceptanceData.IsAccepted,
			}
		}

		blockHash, err := DbHashToDomainHash(dbBlockAcceptanceData.BlockHash)
		if err != nil {
			return nil, err
		}

		domainAcceptanceData[i] = &externalapi.BlockAcceptanceData{
			BlockHash:                 blockHash,
			TransactionAcceptanceData: domainTransactionAcceptanceData,
		}
	}

	return domainAcceptanceData, nil
}
