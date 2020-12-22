package main

import (
	"encoding/hex"
	"fmt"
	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionid"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
)

const feeSompis uint64 = 1000

func send(conf *sendConfig) error {
	toAddress, err := util.DecodeAddress(conf.ToAddress, conf.ActiveNetParams.Prefix)
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
	fromAddress, err := util.NewAddressPubKeyHashFromPublicKey(serializedPublicKey[:], conf.ActiveNetParams.Prefix)
	if err != nil {
		return err
	}

	client, err := rpcclient.NewRPCClient(conf.RPCServer)
	if err != nil {
		return err
	}
	utxos, err := fetchSpendableUTXOs(conf, client, fromAddress.String())
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

func fetchSpendableUTXOs(conf *sendConfig, client *rpcclient.RPCClient, address string) ([]*appmessage.UTXOsByAddressesEntry, error) {
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
		if !isUTXOSpendable(entry, virtualSelectedParentBlueScore, conf.ActiveNetParams.BlockCoinbaseMaturity) {
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

func generateTransaction(keyPair *secp256k1.SchnorrKeyPair, selectedUTXOs []*appmessage.UTXOsByAddressesEntry,
	sompisToSend uint64, change uint64, toAddress util.Address,
	fromAddress util.Address) (*appmessage.RPCTransaction, error) {

	inputs := make([]*externalapi.DomainTransactionInput, len(selectedUTXOs))
	for i, utxo := range selectedUTXOs {
		outpointTransactionIDBytes, err := hex.DecodeString(utxo.Outpoint.TransactionID)
		if err != nil {
			return nil, err
		}
		outpointTransactionID, err := transactionid.FromBytes(outpointTransactionIDBytes)
		if err != nil {
			return nil, err
		}
		outpoint := externalapi.DomainOutpoint{
			TransactionID: *outpointTransactionID,
			Index:         utxo.Outpoint.Index,
		}
		inputs[i] = &externalapi.DomainTransactionInput{PreviousOutpoint: outpoint}
	}

	toScript, err := txscript.PayToAddrScript(toAddress)
	if err != nil {
		return nil, err
	}
	mainOutput := &externalapi.DomainTransactionOutput{
		Value:           sompisToSend,
		ScriptPublicKey: toScript,
	}
	fromScript, err := txscript.PayToAddrScript(fromAddress)
	if err != nil {
		return nil, err
	}
	changeOutput := &externalapi.DomainTransactionOutput{
		Value:           change,
		ScriptPublicKey: fromScript,
	}
	outputs := []*externalapi.DomainTransactionOutput{mainOutput, changeOutput}

	domainTransaction := &externalapi.DomainTransaction{
		Version:      constants.TransactionVersion,
		Inputs:       inputs,
		Outputs:      outputs,
		LockTime:     0,
		SubnetworkID: subnetworks.SubnetworkIDNative,
		Gas:          0,
		Payload:      nil,
		PayloadHash:  externalapi.DomainHash{},
	}

	for i, input := range domainTransaction.Inputs {
		signatureScript, err := txscript.SignatureScript(domainTransaction, i, fromScript, txscript.SigHashAll, keyPair)
		if err != nil {
			return nil, err
		}
		input.SignatureScript = signatureScript
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
