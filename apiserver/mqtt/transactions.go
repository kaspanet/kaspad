package mqtt

import (
	"encoding/json"
	"github.com/daglabs/btcd/apiserver/apimodels"
	"github.com/daglabs/btcd/apiserver/controllers"
	"github.com/daglabs/btcd/btcjson"
	"github.com/daglabs/btcd/rpcclient"
	"github.com/daglabs/btcd/util/daghash"
)

// PublishTransactionsNotifications publishes notification for each transaction of the given block
func PublishTransactionsNotifications(rawTransactions []btcjson.TxRawResult) error {
	if !isConnected() {
		return nil
	}

	transactionIDs := make([]string, len(rawTransactions))
	for i, tx := range rawTransactions {
		transactionIDs[i] = tx.TxID
	}

	transactions, err := controllers.GetTransactionsByIDsHandler(transactionIDs)
	if err != nil {
		return err
	}

	for _, transaction := range transactions {
		err = publishTransactionNotifications(transaction, "transactions/")
		if err != nil {
			return err
		}
	}
	return nil
}

func publishTransactionNotifications(transaction *apimodels.TransactionResponse, topic string) error {
	addresses := uniqueAddressesForTransaction(transaction)
	for _, address := range addresses {
		err := publishTransactionNotificationForAddress(transaction, address, topic)
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

func publishTransactionNotificationForAddress(transaction *apimodels.TransactionResponse, address string, topic string) error {
	payload, err := json.Marshal(transaction)
	if err != nil {
		return err
	}

	token := client.Publish(topic+address, 2, false, payload)
	token.Wait()
	if token.Error() != nil {
		return token.Error()
	}

	return nil
}

// PublishAcceptedTransactionsNotifications publishes notification for each accepted transaction of the given chain-block
func PublishAcceptedTransactionsNotifications(addedChainBlocks []*rpcclient.ChainBlock) error {
	for _, addedChainBlock := range addedChainBlocks {
		for _, acceptedBlock := range addedChainBlock.AcceptedBlocks {
			transactionIDs := make([]string, len(acceptedBlock.AcceptedTxIDs))
			for i, acceptedTxID := range acceptedBlock.AcceptedTxIDs {
				transactionIDs[i] = acceptedTxID.String()
			}

			transactions, err := controllers.GetTransactionsByIDsHandler(transactionIDs)
			if err != nil {
				return err
			}

			for _, transaction := range transactions {
				err = publishTransactionNotifications(transaction, "transactions/accepted/")
				if err != nil {
					return err
				}
			}
			return nil
		}
	}
	return nil
}

// PublishUnacceptedTransactionsNotifications publishes notification for each unaccepted transaction of the given chain-block
func PublishUnacceptedTransactionsNotifications(removedChainHashes []*daghash.Hash) error {
	for _, removedHash := range removedChainHashes {
		transactionIDs, err := controllers.GetTransactionIDsByBlockHashHandler(removedHash)
		if err != nil {
			return err
		}

		transactions, err := controllers.GetTransactionsByIDsHandler(transactionIDs)
		if err != nil {
			return err
		}

		for _, transaction := range transactions {
			err = publishTransactionNotifications(transaction, "transactions/unaccepted/")
			if err != nil {
				return err
			}
		}
	}

	return nil
}
