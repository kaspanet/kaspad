package main

import (
	"encoding/hex"
	"fmt"
	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionid"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
)

const feeSompis uint64 = 1000

func send(conf *sendConfig) error {
	toAddress, err := util.DecodeAddress(conf.ToAddress, util.Bech32PrefixUnknown)
	if err != nil {
		return err
	}

	keyPair, publicKey, err := parsePrivateKey(conf.PrivateKey)
	if err != nil {
		return err
	}

	serializedPublicKey, err := publicKey.Serialize()
	if err != nil {
		return err
	}
	fromAddress, err := util.NewAddressPubKeyHashFromPublicKey(serializedPublicKey[:], toAddress.Prefix())
	if err != nil {
		return err
	}

	client, err := rpcclient.NewRPCClient(conf.RPCServer)
	if err != nil {
		return err
	}
	utxos, err := fetchSpendableUTXOs(client, fromAddress.String())
	if err != nil {
		return err
	}

	sendAmountSompi := uint64(conf.SendAmount * util.SompiPerKaspa)
	totalToSend := sendAmountSompi + feeSompis

	selectedUTXOs, changeSompi, err := selectUTXOs(utxos, totalToSend)
	if err != nil {
		return err
	}

	rpcTransaction, err := generateTransaction(keyPair, selectedUTXOs, sendAmountSompi, changeSompi, toAddress, fromAddress)
	if err != nil {
		return err
	}

	transactionID, err := sendTransaction(client, rpcTransaction)
	if err != nil {
		return err
	}

	fmt.Println("Transaction was sent successfully")
	fmt.Printf("Transaction ID: \t%s\n", transactionID)

	return nil
}

func parsePrivateKey(privateKeyHex string) (*secp256k1.SchnorrKeyPair, *secp256k1.SchnorrPublicKey, error) {
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Error parsing private key hex")
	}
	keyPair, err := secp256k1.DeserializePrivateKeyFromSlice(privateKeyBytes)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Error deserializing private key")
	}
	publicKey, err := keyPair.SchnorrPublicKey()
	if err != nil {
		return nil, nil, errors.Wrap(err, "Error generating public key")
	}
	return keyPair, publicKey, nil
}

func fetchSpendableUTXOs(client *rpcclient.RPCClient, address string) ([]*appmessage.UTXOsByAddressesEntry, error) {
	getUTXOsByAddressesResponse, err := client.GetUTXOsByAddresses([]string{address})
	if err != nil {
		return nil, err
	}
	virtualSelectedParentBlueScoreResponse, err := client.GetVirtualSelectedParentBlueScore()
	if err != nil {
		return nil, err
	}
	virtualSelectedParentBlueScore := virtualSelectedParentBlueScoreResponse.BlueScore

	spendableUTXOs := make([]*appmessage.UTXOsByAddressesEntry, 0)
	for _, entry := range getUTXOsByAddressesResponse.Entries {
		if !isUTXOSpendable(entry, virtualSelectedParentBlueScore) {
			continue
		}
		spendableUTXOs = append(spendableUTXOs, entry)
	}
	return spendableUTXOs, nil
}

func selectUTXOs(utxos []*appmessage.UTXOsByAddressesEntry, totalToSpend uint64) (
	selectedUTXOs []*appmessage.UTXOsByAddressesEntry, changeSompi uint64, err error) {

	selectedUTXOs = []*appmessage.UTXOsByAddressesEntry{}
	totalValue := uint64(0)

	for _, utxo := range utxos {
		selectedUTXOs = append(selectedUTXOs, utxo)
		totalValue += utxo.UTXOEntry.Amount

		if totalValue >= totalToSpend {
			break
		}
	}

	if totalValue < totalToSpend {
		return nil, 0, errors.Errorf("Insufficient funds for send: %f required, while only %f available",
			float64(totalToSpend)/util.SompiPerKaspa, float64(totalValue)/util.SompiPerKaspa)
	}

	return selectedUTXOs, totalValue - totalToSpend, nil
}

func generateTransaction(keyPair *secp256k1.SchnorrKeyPair, selectedUTXOs []*appmessage.UTXOsByAddressesEntry, sompisToSend uint64, change uint64,
	toAddress util.Address, fromAddress util.Address) (*appmessage.RPCTransaction, error) {

	txIns := make([]*appmessage.TxIn, len(selectedUTXOs))
	for i, utxo := range selectedUTXOs {
		outpointTransactionIDBytes, err := hex.DecodeString(utxo.Outpoint.TransactionID)
		if err != nil {
			return nil, err
		}
		outpointTransactionID, err := transactionid.FromBytes(outpointTransactionIDBytes)
		if err != nil {
			return nil, err
		}
		txIns[i] = appmessage.NewTxIn(appmessage.NewOutpoint(outpointTransactionID, utxo.Outpoint.Index), []byte{})
	}

	toScript, err := txscript.PayToAddrScript(toAddress)
	if err != nil {
		return nil, err
	}
	mainTxOut := appmessage.NewTxOut(sompisToSend, toScript)

	fromScript, err := txscript.PayToAddrScript(fromAddress)
	if err != nil {
		return nil, err
	}
	changeTxOut := appmessage.NewTxOut(change, fromScript)

	txOuts := []*appmessage.TxOut{mainTxOut, changeTxOut}

	msgTx := appmessage.NewNativeMsgTx(constants.TransactionVersion, txIns, txOuts)
	domainTransaction := appmessage.MsgTxToDomainTransaction(msgTx)

	for i, txIn := range domainTransaction.Inputs {
		signatureScript, err := txscript.SignatureScript(domainTransaction, i, fromScript, txscript.SigHashAll, keyPair)
		if err != nil {
			return nil, err
		}
		txIn.SignatureScript = signatureScript
	}

	rpcTransaction := appmessage.DomainTransactionToRPCTransaction(domainTransaction)
	return rpcTransaction, nil
}

func sendTransaction(client *rpcclient.RPCClient, rpcTransaction *appmessage.RPCTransaction) (string, error) {
	submitTransactionResponse, err := client.SubmitTransaction(rpcTransaction)
	if err != nil {
		return "", errors.Wrapf(err, "error submitting transaction")
	}
	return submitTransactionResponse.TransactionID, nil
}
