package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/daglabs/btcd/apiserver/apimodels"
	"github.com/daglabs/btcd/blockdag"
	"github.com/daglabs/btcd/faucet/config"
	"github.com/daglabs/btcd/httpserverutils"
	"github.com/daglabs/btcd/txscript"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/util/daghash"
	"github.com/daglabs/btcd/wire"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
)

const (
	sendAmount = 10000
	// Value 8 bytes + serialized varint size for the length of ScriptPubKey +
	// ScriptPubKey bytes.
	outputSize uint64 = 8 + 1 + 25
	minTxFee   uint64 = 3000

	requiredConfirmations                       = 10
	approximateConfirmationsForCoinbaseMaturity = 150
)

type utxoSet map[wire.Outpoint]*blockdag.UTXOEntry

func apiURL(serverPath string) (string, error) {
	cfg, err := config.MainConfig()
	if err != nil {
		return "", err
	}
	u, err := url.Parse(cfg.APIServerURL)
	if err != nil {
		return "", errors.WithStack(err)
	}
	u.Path = path.Join(u.Path, serverPath)
	return u.String(), nil
}

func getFromAPIServer(serverPath string) ([]byte, error) {
	getAPIURL, err := apiURL(serverPath)
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

func postToAPIServer(serverPath string, data interface{}) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return errors.WithStack(err)
	}
	r := bytes.NewReader(dataBytes)
	postAPIURL, err := apiURL(serverPath)
	if err != nil {
		return err
	}
	resp, err := http.Post(postAPIURL, "encoding/json", r)
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
	if !entry.IsCoinbase() {
		return confirmations >= requiredConfirmations
	}
	return confirmations >= approximateConfirmationsForCoinbaseMaturity
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
		if isUTXOMatured(utxoEntry, *utxoResponse.Confirmations) {
			walletUTXOSet[*outpoint] = utxoEntry
		}
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
			return fmt.Errorf("Failed to sign transaction: %s", err)
		}
		txIn.SignatureScript = sigScript
	}

	return nil
}

func fundTx(walletUTXOSet utxoSet, tx *wire.MsgTx, amount uint64) (netAmount uint64, isChangeOutputRequired bool, err error) {
	amountSelected := uint64(0)
	for outpoint, entry := range walletUTXOSet {
		amountSelected += entry.Amount()

		// Add the selected output to the transaction
		tx.AddTxIn(wire.NewTxIn(&outpoint, nil))

		// Check if transaction has enough funds. If we don't have enough
		// coins from the current amount selected to pay the fee continue
		// to grab more coins.
		isTxFunded, _, _, err := isFunded(tx, amountSelected, amount, walletUTXOSet)
		if err != nil {
			return 0, false, err
		}
		if isTxFunded {
			break
		}
	}

	isTxFunded, isChangeOutputRequired, netAmount, err := isFunded(tx, amountSelected, amount, walletUTXOSet)
	if err != nil {
		return 0, false, err
	}
	if !isTxFunded {
		return 0, false, errors.Errorf("not enough funds for coin selection")
	}

	return netAmount, isChangeOutputRequired, nil
}

// Check if the transaction has enough funds to cover the fee
// required for the txn.
func isFunded(tx *wire.MsgTx, amountSelected uint64, targetAmount uint64, walletUTXOSet utxoSet) (isTxFunded, isChangeOutputRequired bool, netAmount uint64, err error) {
	isFundedWithOneOutput, oneOutputFee, err := isFundedWithTargetOutputs(tx, 1, amountSelected, targetAmount, walletUTXOSet)
	if err != nil {
		return false, false, 0, err
	}
	if !isFundedWithOneOutput {
		return false, false, 0, nil
	}
	isFundedWithTwoOutputs, twoOutputsFee, err := isFundedWithTargetOutputs(tx, 2, amountSelected, targetAmount, walletUTXOSet)
	if err != nil {
		return false, false, 0, err
	}
	if isFundedWithTwoOutputs && twoOutputsFee-oneOutputFee < targetAmount-amountSelected {
		return true, true, amountSelected - twoOutputsFee, nil
	}
	return true, false, amountSelected - oneOutputFee, nil
}

// Check if the transaction has enough funds to cover the fee
// required for the txn.
func isFundedWithTargetOutputs(tx *wire.MsgTx, targetNumberOfOutputs uint64, amountSelected uint64, targetAmount uint64, walletUTXOSet utxoSet) (isTxFunded bool, fee uint64, err error) {
	reqFee, err := calcFee(tx, targetNumberOfOutputs, walletUTXOSet)
	if err != nil {
		return false, 0, err
	}
	return amountSelected > reqFee && amountSelected-reqFee >= targetAmount, reqFee, nil
}

func calcFee(msgTx *wire.MsgTx, numberOfOutputs uint64, walletUTXOSet utxoSet) (uint64, error) {
	txMass := calcTxMass(msgTx, walletUTXOSet)
	txMassWithOutputs := txMass + outputsTotalSize(numberOfOutputs)
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
