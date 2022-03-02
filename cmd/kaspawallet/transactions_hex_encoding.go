package main

import (
	"encoding/hex"
	"strings"
)

// hexTransactionsSeparator is used to mark the end of one transaction and the beggining of the next one.
// We use a separator that is not in the hex alphabet, but which will not split selection with a double click
const hexTransactionsSeparator = "_"

func encodeTransactionsToHex(transactions [][]byte) string {
	transactionsInHex := make([]string, len(transactions))
	for i, transaction := range transactions {
		transactionsInHex[i] = hex.EncodeToString(transaction)
	}
	return strings.Join(transactionsInHex, hexTransactionsSeparator)
}

func decodeTransactionsFromHex(transactionsHex string) ([][]byte, error) {
	splitTransactionsHexes := strings.Split(transactionsHex, hexTransactionsSeparator)
	transactions := make([][]byte, len(splitTransactionsHexes))

	var err error
	for i, transactionHex := range splitTransactionsHexes {
		transactions[i], err = hex.DecodeString(transactionHex)
		if err != nil {
			return nil, err
		}
	}

	return transactions, nil
}
