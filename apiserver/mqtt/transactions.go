package mqtt

import (
	"encoding/json"
	"fmt"
	"github.com/daglabs/btcd/apiserver/apimodels"
	"github.com/daglabs/btcd/apiserver/controllers"
	"github.com/daglabs/btcd/btcjson"
	"github.com/jinzhu/gorm"
)

// PublishTransactionsNotifications publishes notification for each transaction of the given block
func PublishTransactionsNotifications(db *gorm.DB, rawBlock *btcjson.GetBlockVerboseResult) error {
	transactionIds := make([]string, len(rawBlock.RawTx))
	for i, tx := range rawBlock.RawTx {
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
	addressesMap := make(map[string]bool)
	addresses := []string{}
	for _, output := range transaction.Outputs {
		if !addressesMap[output.Address] {
			addressesMap[output.Address] = true
			addresses = append(addresses, output.Address)
		}
	}
	return addresses
}

func publishTransactionNotificationForAddress(transaction *apimodels.TransactionResponse, address string) error {
	payload, err := json.Marshal(transaction)
	if err != nil {
		return err
	}

	token := client.Publish(transactionsTopic(address), 0, false, payload)
	token.Wait()
	if token.Error() != nil {
		return token.Error()
	}
	fmt.Printf("Published to topic: %v, message: %s\n\n", transactionsTopic(address), payload)
	return nil
}

func transactionsTopic(address string) string {
	return fmt.Sprintf("transactions/%s", address)
}
