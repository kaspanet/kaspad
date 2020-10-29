package serialization

import "github.com/kaspanet/kaspad/domain/consensus/model"

func DomainAcceptanceDataToDbAcceptanceData(domainAcceptanceData model.AcceptanceData) *DbAcceptanceData {
	dbBlockAcceptanceData := make([]*DbBlockAcceptanceData, len(domainAcceptanceData))
	for i, blockAcceptanceData := range domainAcceptanceData {
		dbTransactionAcceptanceData := make([]*DbTransactionAcceptanceData,
			len(blockAcceptanceData.TransactionAcceptanceData))

		for j, transactionAcceptanceData := range blockAcceptanceData.TransactionAcceptanceData {
			dbTransactionAcceptanceData[j] = &DbTransactionAcceptanceData{
				Transaction: nil,
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

func DbAcceptanceDataToDomainAcceptanceData(dbAcceptanceData *DbAcceptanceData) (model.AcceptanceData, error) {
	panic("implement me")
}
