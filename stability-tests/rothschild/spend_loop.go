package main

import (
	"encoding/hex"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionid"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"
	"math/rand"
	"strings"
	"sync/atomic"
	"time"

	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/app/appmessage"

	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
)

var pendingOutpoints = map[appmessage.RPCOutpoint]time.Time{}

func spendLoop(client *rpcclient.RPCClient, addresses *addressesList,
	utxosChangedNotificationChan <-chan *appmessage.UTXOsChangedNotificationMessage) <-chan struct{} {

	doneChan := make(chan struct{})

	spawn("spendLoop", func() {
		log.Infof("Fetching the initial UTXO set")
		utxos, err := fetchSpendableUTXOs(client, addresses.myAddress.EncodeAddress())
		if err != nil {
			panic(err)
		}

		cfg := activeConfig()
		ticker := time.NewTicker(time.Duration(cfg.TransactionInterval) * time.Millisecond)
		for range ticker.C {
			shuffleUTXOs(utxos)

			hasFunds, err := maybeSendTransaction(client, addresses, utxos)
			if err != nil {
				panic(err)
			}

			checkTransactions(utxosChangedNotificationChan)

			if !hasFunds {
				log.Infof("No funds. Refetching UTXO set.")
				utxos, err = fetchSpendableUTXOs(client, addresses.myAddress.EncodeAddress())
				if err != nil {
					panic(err)
				}
			}

			if atomic.LoadInt32(&shutdown) != 0 {
				close(doneChan)
				return
			}
		}
	})

	return doneChan
}

func checkTransactions(utxosChangedNotificationChan <-chan *appmessage.UTXOsChangedNotificationMessage) {
	isDone := false
	for !isDone {
		select {
		case notification := <-utxosChangedNotificationChan:
			for _, removed := range notification.Removed {
				sendTime, ok := pendingOutpoints[*removed.Outpoint]
				if !ok {
					continue // this is coinbase transaction paying to our address or some transaction from an old run
				}

				log.Infof("Output %s:%d accepted. Time since send: %s",
					removed.Outpoint.TransactionID, removed.Outpoint.Index, time.Now().Sub(sendTime))

				delete(pendingOutpoints, *removed.Outpoint)
			}
		default:
			isDone = true
		}
	}

	for pendingOutpoint, txTime := range pendingOutpoints {
		timeSince := time.Now().Sub(txTime)
		if timeSince > 10*time.Minute {
			log.Tracef("Outpoint %s:%d is pending for %s",
				pendingOutpoint.TransactionID, pendingOutpoint.Index, timeSince)
		}
	}
}

const balanceEpsilon = 10_000         // 10,000 sompi = 0.0001 kaspa
const feeAmount = balanceEpsilon * 10 // use high fee amount, because can have a large number of inputs

func maybeSendTransaction(client *rpcclient.RPCClient, addresses *addressesList,
	availableUTXOs []*appmessage.UTXOsByAddressesEntry) (hasFunds bool, err error) {

	sendAmount := randomizeSpendAmount()
	totalSendAmount := sendAmount + feeAmount

	selectedUTXOs, selectedValue, err := selectUTXOs(availableUTXOs, totalSendAmount)
	if err != nil {
		return false, err
	}

	if len(selectedUTXOs) == 0 {
		return false, nil
	}

	if selectedValue < totalSendAmount {
		sendAmount = selectedValue - feeAmount
	}

	change := selectedValue - sendAmount - feeAmount

	spendAddress := randomizeSpendAddress(addresses)

	rpcTransaction, err := generateTransaction(
		addresses.myPrivateKey, selectedUTXOs, sendAmount, change, spendAddress, addresses.myAddress)
	if err != nil {
		return false, err
	}

	spawn("sendTransaction", func() {
		transactionID, err := sendTransaction(client, rpcTransaction)
		if err != nil {
			if !strings.Contains(err.Error(), "orphan transaction") {
				panic(errors.Wrapf(err, "error sending transaction: %s", err))
			}
			log.Warnf("Double spend error: %s", err)
		} else {
			log.Infof("Sent transaction %s worth %f kaspa with %d inputs and %d outputs", transactionID,
				float64(sendAmount)/util.SompiPerKaspa, len(rpcTransaction.Inputs), len(rpcTransaction.Outputs))
		}
	})

	updateState(selectedUTXOs)

	return true, nil
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
func isUTXOSpendable(entry *appmessage.UTXOsByAddressesEntry, virtualSelectedParentBlueScore uint64) bool {
	blockBlueScore := entry.UTXOEntry.BlockBlueScore
	if !entry.UTXOEntry.IsCoinbase {
		const minConfirmations = 10
		return blockBlueScore+minConfirmations < virtualSelectedParentBlueScore
	}
	coinbaseMaturity := activeConfig().ActiveNetParams.BlockCoinbaseMaturity
	return blockBlueScore+coinbaseMaturity < virtualSelectedParentBlueScore
}

func shuffleUTXOs(utxos []*appmessage.UTXOsByAddressesEntry) {
	rand.Shuffle(len(utxos), func(i, j int) { utxos[i], utxos[j] = utxos[j], utxos[i] })
}

func updateState(selectedUTXOs []*appmessage.UTXOsByAddressesEntry) {
	for _, utxo := range selectedUTXOs {
		pendingOutpoints[*utxo.Outpoint] = time.Now()
	}
}

func filterSpentUTXOsAndCalculateBalance(utxos []*appmessage.UTXOsByAddressesEntry) (
	filteredUTXOs []*appmessage.UTXOsByAddressesEntry, balance uint64) {

	balance = 0
	for _, utxo := range utxos {
		if _, ok := pendingOutpoints[*utxo.Outpoint]; ok {
			continue
		}
		balance += utxo.UTXOEntry.Amount
		filteredUTXOs = append(filteredUTXOs, utxo)
	}
	return filteredUTXOs, balance
}

func randomizeSpendAddress(addresses *addressesList) util.Address {
	spendAddressIndex := rand.Intn(len(addresses.spendAddresses))

	return addresses.spendAddresses[spendAddressIndex]
}

func randomizeSpendAmount() uint64 {
	const maxAmountToSent = 10 * feeAmount
	amountToSend := rand.Int63n(int64(maxAmountToSent))

	// round to balanceEpsilon
	amountToSend = amountToSend / balanceEpsilon * balanceEpsilon
	if amountToSend < balanceEpsilon {
		amountToSend = balanceEpsilon
	}

	return uint64(amountToSend)
}

func selectUTXOs(utxos []*appmessage.UTXOsByAddressesEntry, amountToSend uint64) (
	selectedUTXOs []*appmessage.UTXOsByAddressesEntry, selectedValue uint64, err error) {

	selectedUTXOs = []*appmessage.UTXOsByAddressesEntry{}
	selectedValue = uint64(0)

	for _, utxo := range utxos {
		if _, ok := pendingOutpoints[*utxo.Outpoint]; ok {
			continue
		}

		selectedUTXOs = append(selectedUTXOs, utxo)
		selectedValue += utxo.UTXOEntry.Amount

		if selectedValue >= amountToSend {
			break
		}

		const maxInputs = 100
		if len(selectedUTXOs) == maxInputs {
			log.Infof("Selected %d UTXOs so sending the transaction with %d sompis instead "+
				"of %d", maxInputs, selectedValue, amountToSend)
			break
		}
	}

	return selectedUTXOs, selectedValue, nil
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
		Version:      constants.MaxTransactionVersion,
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

func init() {
	rand.Seed(time.Now().UnixNano())
}
