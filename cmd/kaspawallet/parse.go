package main

import (
	"encoding/hex"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"strings"
)

func parse(conf *parseConfig) error {
	if conf.Transaction == "" && conf.TransactionFile == "" {
		return errors.Errorf("Either --transaction or --transaction-file is required")
	}
	if conf.Transaction != "" && conf.TransactionFile != "" {
		return errors.Errorf("Both --transaction and --transaction-file cannot be passed at the same time")
	}

	transactionHex := conf.Transaction
	if conf.TransactionFile != "" {
		transactionHexBytes, err := ioutil.ReadFile(conf.TransactionFile)
		if err != nil {
			return errors.Wrapf(err, "Could not read hex from %s", conf.TransactionFile)
		}
		transactionHex = strings.TrimSpace(string(transactionHexBytes))
	}

	transaction, err := hex.DecodeString(transactionHex)
	if err != nil {
		return err
	}

	fmt.Printf("Transaction length: \t%d\n", len(transaction))

	return nil
}
