package mqtt

import (
	"fmt"
	"github.com/daglabs/btcd/apiserver/apimodels"
	"github.com/daglabs/btcd/apiserver/controllers"
	"github.com/daglabs/btcd/btcjson"
	"github.com/jinzhu/gorm"
)

// PublishTransactionsNotifications publishes notification for each transaction of the given block
func PublishTransactionsNotifications(db *gorm.DB, rawTransactions []btcjson.TxRawResult) error {
	if !isConnected() {
		return nil
	}

	transactionIds := make([]string, len(rawTransactions))
	for i, tx := range rawTransactions {
		transactionIds[i] = tx.TxID
	}

	transactions, err := controllers.GetTransactionsByIdsHandler(db, transactionIds)
	if err != nil {
		return err
	}

	for _, transaction := range transactions {
		err = publishTransactionNotifications(transaction)
		if err != nil {
			return err
		}
	}
	return nil
}

func publishTransactionNotifications(transaction *apimodels.TransactionResponse) error {
	addresses := uniqueAddressesForTransaction(transaction)
	for _, address := range addresses {
		err := publishTransactionNotificationForAddress(transaction, address)
		if err != nil {
			return err
		}
	}
	return nil
}

func uniqueAddressesForTransaction(transaction *apimodels.TransactionResponse) []string {
	addressesMap := make(map[string]struct{})
	addresses := []string{}
	for _, output := range transaction.Outputs {
		if _, exists := addressesMap[output.Address]; !exists {
			addresses = append(addresses, output.Address)
			addressesMap[output.Address] = struct{}{}
		}
	}
	return addresses
}

func publishTransactionNotificationForAddress(transaction *apimodels.TransactionResponse, address string) error {
	return publish(transactionsTopic(address), transaction)
}

func transactionsTopic(address string) string {
	return fmt.Sprintf("transactions/%s", address)
}
