package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/kaspanet/kasparov/apimodels"
	"github.com/pkg/errors"
)

const (
	getUTXOsEndpoint        = "utxos/address"
	sendTransactionEndpoint = "transaction"
)

// resourceURL returns a full concatenated URL from the base
// kasparov server URL and the given path.
func resourceURL(kasparovAddress string, pathElements ...string) (string, error) {
	kasparovURL, err := url.Parse(kasparovAddress)
	if err != nil {
		return "", errors.WithStack(err)
	}
	pathElements = append([]string{kasparovURL.Path}, pathElements...)
	kasparovURL.Path = path.Join(pathElements...)
	return kasparovURL.String(), nil
}

func getUTXOs(kasparovAddress string, address string) ([]*apimodels.TransactionOutputResponse, error) {
	requestURL, err := resourceURL(kasparovAddress, getUTXOsEndpoint, address)
	if err != nil {
		return nil, err
	}
	response, err := http.Get(requestURL)
	if err != nil {
		return nil, errors.Wrap(err, "Error getting UTXOs from Kasparov server")
	}
	body, err := readResponse(response)
	if err != nil {
		return nil, errors.Wrap(err, "Error reading UTXOs from Kasparov server response")
	}

	utxos := []*apimodels.TransactionOutputResponse{}

	err = json.Unmarshal(body, &utxos)
	if err != nil {
		return nil, errors.Wrap(err, "Error unmarshalling UTXOs")
	}

	return utxos, nil
}

func readResponse(response *http.Response) (body []byte, err error) {
	defer response.Body.Close()

	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Error reading response")
	}

	if response.StatusCode != http.StatusOK {
		return nil, errors.Errorf("Response status %s\nResponseBody:\n%s", response.Status, body)
	}

	return body, nil
}

func printErrorAndExit(err error) {
	fmt.Fprintf(os.Stderr, "%s\n", err)
	os.Exit(1)
}
