package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/daglabs/btcd/apiserver/apimodels"
	"github.com/daglabs/btcd/util"
	"github.com/pkg/errors"
)

func getUTXOs(apiAddress string, address string) ([]*apimodels.TransactionOutputResponse, error) {
	resp, err := http.Get(fmt.Sprintf("%s/utxos/%s", apiAddress, address))
	if err != nil {
		return nil, errors.Wrap(err, "Error getting utxos from API server")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Error reading utxos from API server response")
	}

	utxos := []*apimodels.TransactionOutputResponse{}

	err = json.Unmarshal(body, utxos)
	if err != nil {
		return nil, errors.Wrap(err, "Error unmarshalling utxos")
	}

	return utxos, nil
}

func balance(conf *balanceConfig) error {
	utxos, err := getUTXOs(conf.APIAddress, conf.Address)
	if err != nil {
		return err
	}

	var availableBalance, pendingBalance uint64 = 0, 0
	for _, utxo := range utxos {
		if utxo.IsSpendable != nil && *utxo.IsSpendable {
			availableBalance += utxo.Value
		} else {
			pendingBalance += utxo.Value
		}
	}

	fmt.Printf("Available balance is %f", float64(availableBalance)/util.SatoshiPerBitcoin)
	fmt.Printf("In addition, immature coinbase balance is %f", float64(availableBalance)/util.SatoshiPerBitcoin)

	return nil
}
