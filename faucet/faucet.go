package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/faucet/config"
	"github.com/kaspanet/kaspad/httpserverutils"
	"github.com/kaspanet/kaspad/kasparov/kasparovd/apimodels"
	"github.com/kaspanet/kaspad/txscript"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

const (
	sendAmount = 10000
	// Value 8 bytes + serialized varint size for the length of ScriptPubKey +
	// ScriptPubKey bytes.
	outputSize uint64 = 8 + 1 + 25
	minTxFee   uint64 = 3000

	requiredConfirmations = 10
)

type utxoSet map[wire.Outpoint]*blockdag.UTXOEntry

// apiURL returns a full concatenated URL from the base
// API server URL and the given path.
func apiURL(requestPath string) (string, error) {
	cfg, err := config.MainConfig()
	if err != nil {
		return "", err
	}
	u, err := url.Parse(cfg.KasparovdURL)
	if err != nil {
		return "", errors.WithStack(err)
	}
	u.Path = path.Join(u.Path, requestPath)
	return u.String(), nil
}

// getFromAPIServer makes an HTTP GET request to the API server
// to the given request path, and returns the response body.
func getFromAPIServer(requestPath string) ([]byte, error) {
	getAPIURL, err := apiURL(requestPath)
	if err != nil {
		return nil, err
	}
	resp, err := http.Get(getAPIURL)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(errors.WithStack(err))
		}
	}()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if resp.StatusCode != http.StatusOK {
		clientError := &httpserverutils.ClientError{}
		err := json.Unmarshal(body, &clientError)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return nil, errors.WithStack(clientError)
	}
	return body, nil
}

// getFromAPIServer makes an HTTP POST request to the API server
// to the given request path. It converts the given data to JSON,
// and post it as the POST data.
func postToAPIServer(requestPath string, data interface{}) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return errors.WithStack(err)
	}
	r := bytes.NewReader(dataBytes)
	postAPIURL, err := apiURL(requestPath)
	if err != nil {
		return err
	}
	resp, err := http.Post(postAPIURL, "application/json", r)
	if err != nil {
		return errors.WithStack(err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(errors.WithStack(err))
		}
	}()
	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.WithStack(err)
		}
		clientError := &httpserverutils.ClientError{}
		err = json.Unmarshal(body, &clientError)
		if err != nil {
			return errors.WithStack(err)
		}
		return errors.WithStack(clientError)
	}
	return nil
}

func isUTXOMatured(entry *blockdag.UTXOEntry, confirmations uint64) bool {
	if entry.IsCoinbase() {
		return confirmations >= config.ActiveNetParams().BlockCoinbaseMaturity
	}
	return confirmations >= requiredConfirmations
}

func getWalletUTXOSet() (utxoSet, error) {
	body, err := getFromAPIServer(fmt.Sprintf("utxos/address/%s", faucetAddress.EncodeAddress()))
	if err != nil {
		return nil, err
	}
	utxoResponses := []*apimodels.TransactionOutputResponse{}
	err = json.Unmarshal(body, &utxoResponses)
	if err != nil {
		return nil, err
	}
	walletUTXOSet := make(utxoSet)
	for _, utxoResponse := range utxoResponses {
		scriptPubKey, err := hex.DecodeString(utxoResponse.ScriptPubKey)
		if err != nil {
			return nil, err
		}
		txOut := &wire.TxOut{
			Value:        utxoResponse.Value,
			ScriptPubKey: scriptPubKey,
		}
		txID, err := daghash.NewTxIDFromStr(utxoResponse.TransactionID)
		if err != nil {
			return nil, err
		}
		outpoint := wire.NewOutpoint(txID, utxoResponse.Index)
		utxoEntry := blockdag.NewUTXOEntry(txOut, *utxoResponse.IsCoinbase, utxoResponse.AcceptingBlockBlueScore)
		if !isUTXOMatured(utxoEntry, *utxoResponse.Confirmations) {
			continue
		}
		walletUTXOSet[*outpoint] = utxoEntry
	}
	return walletUTXOSet, nil
}

func sendToAddress(address util.Address) (*wire.MsgTx, error) {
	tx, err := createTx(address)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))
	if err := tx.Serialize(buf); err != nil {
		return nil, err
	}
	rawTx := &apimodels.RawTransaction{RawTransaction: hex.EncodeToString(buf.Bytes())}
	return tx, postToAPIServer("transaction", rawTx)
}

func createTx(address util.Address) (*wire.MsgTx, error) {
	walletUTXOSet, err := getWalletUTXOSet()
	if err != nil {
		return nil, err
	}
	tx, err := createUnsignedTx(walletUTXOSet, address)
	if err != nil {
		return nil, err
	}
	err = signTx(walletUTXOSet, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func createUnsignedTx(walletUTXOSet utxoSet, address util.Address) (*wire.MsgTx, error) {
	tx := wire.NewNativeMsgTx(wire.TxVersion, nil, nil)
	netAmount, isChangeOutputRequired, err := fundTx(walletUTXOSet, tx, sendAmount)
	if err != nil {
		return nil, err
	}
	if isChangeOutputRequired {
		tx.AddTxOut(&wire.TxOut{
			Value:        sendAmount,
			ScriptPubKey: address.ScriptAddress(),
		})
		tx.AddTxOut(&wire.TxOut{
			Value:        netAmount - sendAmount,
			ScriptPubKey: faucetScriptPubKey,
		})
		return tx, nil
	}
	tx.AddTxOut(&wire.TxOut{
		Value:        netAmount,
		ScriptPubKey: address.ScriptAddress(),
	})
	return tx, nil
}

// signTx signs a transaction
func signTx(walletUTXOSet utxoSet, tx *wire.MsgTx) error {
	for i, txIn := range tx.TxIn {
		outpoint := txIn.PreviousOutpoint

		sigScript, err := txscript.SignatureScript(tx, i, walletUTXOSet[outpoint].ScriptPubKey(),
			txscript.SigHashAll, faucetPrivateKey, true)
		if err != nil {
			return errors.Errorf("Failed to sign transaction: %s", err)
		}
		txIn.SignatureScript = sigScript
	}

	return nil
}

func fundTx(walletUTXOSet utxoSet, tx *wire.MsgTx, amount uint64) (netAmount uint64, isChangeOutputRequired bool, err error) {
	amountSelected := uint64(0)
	isTxFunded := false
	for outpoint, entry := range walletUTXOSet {
		amountSelected += entry.Amount()

		// Add the selected output to the transaction
		tx.AddTxIn(wire.NewTxIn(&outpoint, nil))

		// Check if transaction has enough funds. If we don't have enough
		// coins from the current amount selected to pay the fee continue
		// to grab more coins.
		isTxFunded, isChangeOutputRequired, netAmount, err = isFundedAndIsChangeOutputRequired(tx, amountSelected, amount, walletUTXOSet)
		if err != nil {
			return 0, false, err
		}
		if isTxFunded {
			break
		}
	}

	if !isTxFunded {
		return 0, false, errors.Errorf("not enough funds for coin selection")
	}

	return netAmount, isChangeOutputRequired, nil
}

// isFundedAndIsChangeOutputRequired returns three values and an error:
// * isTxFunded is whether the transaction inputs cover the target amount + the required fee.
// * isChangeOutputRequired is whether it is profitable to add an additional change
//   output to the transaction.
// * netAmount is the amount of coins that will be eventually sent to the recipient. If no
//   change output is needed, the netAmount will be usually a little bit higher than the
//   targetAmount. Otherwise, it'll be the same as the targetAmount.
func isFundedAndIsChangeOutputRequired(tx *wire.MsgTx, amountSelected uint64, targetAmount uint64, walletUTXOSet utxoSet) (isTxFunded, isChangeOutputRequired bool, netAmount uint64, err error) {
	// First check if it can be funded with one output and the required fee for it.
	isFundedWithOneOutput, oneOutputFee, err := isFundedWithNumberOfOutputs(tx, 1, amountSelected, targetAmount, walletUTXOSet)
	if err != nil {
		return false, false, 0, err
	}
	if !isFundedWithOneOutput {
		return false, false, 0, nil
	}

	// Now check if it can be funded with two outputs and the required fee for it.
	isFundedWithTwoOutputs, twoOutputsFee, err := isFundedWithNumberOfOutputs(tx, 2, amountSelected, targetAmount, walletUTXOSet)
	if err != nil {
		return false, false, 0, err
	}

	// If it can be funded with two outputs, check if adding a change output worth it: i.e. check if
	// the amount you save by not sending the recipient the whole inputs amount (minus fees) is greater
	// than the additional fee that is required by adding a change output. If this is the case, return
	// isChangeOutputRequired as true.
	if isFundedWithTwoOutputs && twoOutputsFee-oneOutputFee < targetAmount-amountSelected {
		return true, true, amountSelected - twoOutputsFee, nil
	}
	return true, false, amountSelected - oneOutputFee, nil
}

// isFundedWithNumberOfOutputs returns whether the transaction inputs cover
// the target amount + the required fee with the assumed number of outputs.
func isFundedWithNumberOfOutputs(tx *wire.MsgTx, numberOfOutputs uint64, amountSelected uint64, targetAmount uint64, walletUTXOSet utxoSet) (isTxFunded bool, fee uint64, err error) {
	reqFee, err := calcFee(tx, numberOfOutputs, walletUTXOSet)
	if err != nil {
		return false, 0, err
	}
	return amountSelected > reqFee && amountSelected-reqFee >= targetAmount, reqFee, nil
}

func calcFee(msgTx *wire.MsgTx, numberOfOutputs uint64, walletUTXOSet utxoSet) (uint64, error) {
	txMass := calcTxMass(msgTx, walletUTXOSet)
	txMassWithOutputs := txMass + outputsTotalSize(numberOfOutputs)*blockdag.MassPerTxByte
	cfg, err := config.MainConfig()
	if err != nil {
		return 0, err
	}
	reqFee := uint64(float64(txMassWithOutputs) * cfg.FeeRate)
	if reqFee < minTxFee {
		return minTxFee, nil
	}
	return reqFee, nil
}

func outputsTotalSize(numberOfOutputs uint64) uint64 {
	return numberOfOutputs*outputSize + uint64(wire.VarIntSerializeSize(numberOfOutputs))
}

func calcTxMass(msgTx *wire.MsgTx, walletUTXOSet utxoSet) uint64 {
	previousScriptPubKeys := getPreviousScriptPubKeys(msgTx, walletUTXOSet)
	return blockdag.CalcTxMass(util.NewTx(msgTx), previousScriptPubKeys)
}

func getPreviousScriptPubKeys(msgTx *wire.MsgTx, walletUTXOSet utxoSet) [][]byte {
	previousScriptPubKeys := make([][]byte, len(msgTx.TxIn))
	for i, txIn := range msgTx.TxIn {
		outpoint := txIn.PreviousOutpoint
		previousScriptPubKeys[i] = walletUTXOSet[outpoint].ScriptPubKey()
	}
	return previousScriptPubKeys
}
