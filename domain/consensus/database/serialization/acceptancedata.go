package serialization

import "github.com/kaspanet/kaspad/domain/consensus/model"

// DomainAcceptanceDataToDbAcceptanceData converts model.AcceptanceData to DbAcceptanceData
func DomainAcceptanceDataToDbAcceptanceData(domainAcceptanceData model.AcceptanceData) *DbAcceptanceData {
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

		dbBlockAcceptanceData[i] = &DbBlockAcceptanceData{
			TransactionAcceptanceData: dbTransactionAcceptanceData,
		}
	}

	return &DbAcceptanceData{
		BlockAcceptanceData: dbBlockAcceptanceData,
	}
}

// DbAcceptanceDataToDomainAcceptanceData converts DbAcceptanceData to model.AcceptanceData
func DbAcceptanceDataToDomainAcceptanceData(dbAcceptanceData *DbAcceptanceData) (model.AcceptanceData, error) {
	domainAcceptanceData := make(model.AcceptanceData, len(dbAcceptanceData.BlockAcceptanceData))
	for i, dbBlockAcceptanceData := range dbAcceptanceData.BlockAcceptanceData {
		domainTransactionAcceptanceData := make([]*model.TransactionAcceptanceData,
			len(dbBlockAcceptanceData.TransactionAcceptanceData))

		for j, dbTransactionAcceptanceData := range dbBlockAcceptanceData.TransactionAcceptanceData {
			domainTransaction, err := DbTransactionToDomainTransaction(dbTransactionAcceptanceData.Transaction)
			if err != nil {
				return nil, err
			}
			domainTransactionAcceptanceData[j] = &model.TransactionAcceptanceData{
				Transaction: domainTransaction,
				Fee:         dbTransactionAcceptanceData.Fee,
				IsAccepted:  dbTransactionAcceptanceData.IsAccepted,
			}
		}

		domainAcceptanceData[i] = &model.BlockAcceptanceData{
			TransactionAcceptanceData: domainTransactionAcceptanceData,
		}
	}

	return domainAcceptanceData, nil
}
