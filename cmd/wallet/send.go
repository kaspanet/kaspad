package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/daglabs/btcd/apiserver/apimodels"
	"github.com/daglabs/btcd/btcec"
	"github.com/daglabs/btcd/txscript"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/util/daghash"
	"github.com/daglabs/btcd/wire"
	"github.com/pkg/errors"
)

const feeSatoshis uint64 = 1000

func send(conf *sendConfig) error {
	toAddress, err := util.DecodeAddress(conf.ToAddress, util.Bech32PrefixUnknown)
	if err != nil {
		return err
	}

	privateKey, publicKey, err := parsePrivateKey(conf.PrivateKey)
	if err != nil {
		return err
	}

	fromAddress, err := util.NewAddressPubKeyHashFromPublicKey(publicKey.SerializeCompressed(), toAddress.Prefix())
	if err != nil {
		return err
	}

	utxos, err := getUTXOs(conf.APIAddress, fromAddress.String())
	if err != nil {
		return err
	}

	satoshisToSend := uint64(conf.SendAmount * util.SatoshiPerBitcoin)
	totalToSpend := satoshisToSend + feeSatoshis

	selectedUTXOs, totalValue, err := selectUTXOs(utxos, totalToSpend)
	if err != nil {
		return err
	}

	msgTx, err := generateTx(privateKey, selectedUTXOs, satoshisToSend, totalValue, toAddress, fromAddress)
	if err != nil {
		return err
	}

	txBuffer := bytes.NewBuffer(make([]byte, 0, msgTx.SerializeSize()))
	if err := msgTx.BtcEncode(txBuffer, 0); err != nil {
		return err
	}
	rawTx := &apimodels.RawTransaction{
		RawTransaction: txBuffer.String(),
	}
	txBytes, err := json.Marshal(rawTx)
	if err != nil {
		return err
	}

	http.Post(fmt.Sprintf("%s/transaction", conf.APIAddress), "application/json", bytes.NewBuffer(txBytes))

	return nil
}

func generateTx(privateKey *btcec.PrivateKey, selectedUTXOs []*apimodels.TransactionOutputResponse, satoshisToSend uint64, totalValue uint64,
	toAddress util.Address, fromAddress util.Address) (*wire.MsgTx, error) {

	txIns := make([]*wire.TxIn, len(selectedUTXOs))
	for i, utxo := range selectedUTXOs {
		txID, err := daghash.NewTxIDFromStr(utxo.TransactionID)
		if err != nil {
			return nil, err
		}

		txIns[i] = wire.NewTxIn(wire.NewOutpoint(txID, utxo.Index), []byte{})
	}

	toScript, err := txscript.PayToAddrScript(toAddress)
	if err != nil {
		return nil, err
	}
	mainTxOut := wire.NewTxOut(satoshisToSend, toScript)

	fromScript, err := txscript.PayToAddrScript(fromAddress)
	if err != nil {
		return nil, err
	}
	changeTxOut := wire.NewTxOut(totalValue-satoshisToSend-feeSatoshis, fromScript)

	txOuts := []*wire.TxOut{mainTxOut, changeTxOut}

	tx := wire.NewNativeMsgTx(wire.TxVersion, txIns, txOuts)

	for i, txIn := range tx.TxIn {
		signatureScript, err := txscript.SignatureScript(tx, i, fromScript, txscript.SigHashAll, privateKey, true)
		if err != nil {
			return nil, err
		}
		txIn.SignatureScript = signatureScript
	}

	return tx, nil
}

func selectUTXOs(utxos []*apimodels.TransactionOutputResponse, totalToSpend uint64) (
	[]*apimodels.TransactionOutputResponse, uint64, error) {

	selectedUTXOs := []*apimodels.TransactionOutputResponse{}
	totalValue := uint64(0)

	for _, utxo := range utxos {
		if utxo.IsSpendable == nil || !*utxo.IsSpendable {
			continue
		}

		selectedUTXOs = append(selectedUTXOs, utxo)
		totalValue += utxo.Value

		if totalValue >= totalToSpend {
			break
		}
	}

	if totalValue < totalToSpend {
		return nil, 0, errors.Errorf("Insufficient funds for send: %f required, while only %f available",
			float64(totalToSpend)/util.SatoshiPerBitcoin, float64(totalValue)/util.SatoshiPerBitcoin)
	}

	return selectedUTXOs, totalValue, nil
}

func parsePrivateKey(privateKeyHex string) (*btcec.PrivateKey, *btcec.PublicKey, error) {
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Error parsing private key hex")
	}
	privateKey, publicKey := btcec.PrivKeyFromBytes(btcec.S256(), privateKeyBytes)
	return privateKey, publicKey, nil
}
