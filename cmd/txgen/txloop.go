package main

import (
	"bytes"
	"encoding/hex"
	"github.com/pkg/errors"
	"math"
	"math/rand"
	"time"

	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/kaspajson"
	"github.com/kaspanet/kaspad/txscript"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/subnetworkid"
	"github.com/kaspanet/kaspad/wire"
)

const (
	// Those constants should be updated, when monetary policy changed
	minSpendableAmount uint64 = 10000
	maxSpendableAmount uint64 = 5 * minSpendableAmount
	minTxFee           uint64 = 3000

	// spendSize is the largest number of bytes of a sigScript
	// which spends a p2pkh output: OP_DATA_73 <sig> OP_DATA_33 <pubkey>
	spendSize uint64 = 1 + 73 + 1 + 33
	// Value 8 bytes + serialized varint size for the length of ScriptPubKey +
	// ScriptPubKey bytes.
	outputSize uint64 = 8 + 1 + 25

	txLifeSpan                                  = 1000
	requiredConfirmations                       = 10
	approximateConfirmationsForCoinbaseMaturity = 150
	searchRawTransactionResultCount             = 1000
	searchRawTransactionMaxResults              = 5000
	txMaxQueueLength                            = 10000
	maxResendDepth                              = 500
	minSecondaryTxAmount                        = 100000000
)

type walletTransaction struct {
	tx                         *util.Tx
	chainHeight                uint64
	checkConfirmationCountdown uint64
	confirmed                  bool
}

type utxoSet map[wire.Outpoint]*wire.TxOut

func isDust(value uint64) bool {
	return value < minSpendableAmount+minTxFee
}

var (
	random                 = rand.New(rand.NewSource(time.Now().UnixNano()))
	primaryScriptPubKey    []byte
	secondaryScriptPubKey  []byte
	sentToSecondaryAddress bool
)

// txLoop performs main loop of transaction generation
func txLoop(client *txgenClient, cfg *configFlags) error {
	filterAddresses := []util.Address{p2pkhAddress}
	var err error
	primaryScriptPubKey, err = txscript.PayToAddrScript(p2pkhAddress)
	if err != nil {
		return errors.Errorf("failed to generate primaryScriptPubKey to address: %s", err)
	}

	if secondaryAddress != nil {
		secondaryScriptPubKey, err = txscript.PayToAddrScript(secondaryAddress)
		if err != nil {
			return errors.Errorf("failed to generate primaryScriptPubKey to address: %s", err)
		}

		filterAddresses = append(filterAddresses, secondaryAddress)
	}

	err = client.LoadTxFilter(true, filterAddresses, nil)
	if err != nil {
		return err
	}

	gasLimitMap := make(map[subnetworkid.SubnetworkID]uint64)
	gasLimitMap[*subnetworkid.SubnetworkIDNative] = 0

	walletUTXOSet, walletTxs, err := getInitialUTXOSetAndWalletTxs(client, gasLimitMap)
	if err != nil {
		return err
	}

	txChan := make(chan *wire.MsgTx, txMaxQueueLength)
	spawn(func() {
		err := sendTransactionLoop(client, cfg.TxInterval, txChan)
		if err != nil {
			panic(err)
		}
	})

	for blockAdded := range client.onBlockAdded {
		log.Infof("Block %s Added with %d relevant transactions", blockAdded.header.BlockHash(), len(blockAdded.txs))
		err := updateSubnetworks(blockAdded.txs, gasLimitMap)
		if err != nil {
			return err
		}
		updateWalletTxs(blockAdded, walletTxs)
		err = enqueueTransactions(client, blockAdded, walletUTXOSet, walletTxs, txChan, cfg, gasLimitMap)
		if err != nil {
			return err
		}
	}
	return nil
}

func updateSubnetworks(txs []*util.Tx, gasLimitMap map[subnetworkid.SubnetworkID]uint64) error {
	for _, tx := range txs {
		msgTx := tx.MsgTx()
		if msgTx.SubnetworkID.IsEqual(subnetworkid.SubnetworkIDRegistry) {
			subnetworkID, err := blockdag.TxToSubnetworkID(msgTx)
			if err != nil {
				return errors.Errorf("could not build subnetwork ID: %s", err)
			}
			gasLimit := blockdag.ExtractGasLimit(msgTx)
			log.Infof("Found subnetwork %s with gas limit %d", subnetworkID, gasLimit)
			gasLimitMap[*subnetworkID] = gasLimit
		}
	}
	return nil
}

func sendTransactionLoop(client *txgenClient, interval uint64, txChan chan *wire.MsgTx) error {
	var ticker *time.Ticker
	if interval != 0 {
		ticker = time.NewTicker(time.Duration(interval) * time.Millisecond)
	}
	for tx := range txChan {
		_, err := client.SendRawTransaction(tx, true)
		log.Infof("Sending tx %s to subnetwork %s with %d inputs, %d outputs, %d payload size and %d gas", tx.TxID(), tx.SubnetworkID, len(tx.TxIn), len(tx.TxOut), len(tx.Payload), tx.Gas)
		if err != nil {
			return err
		}
		if ticker != nil {
			<-ticker.C
		}
	}
	return nil
}

func getInitialUTXOSetAndWalletTxs(client *txgenClient, gasLimitMap map[subnetworkid.SubnetworkID]uint64) (utxoSet, map[daghash.TxID]*walletTransaction, error) {
	walletUTXOSet := make(utxoSet)
	walletTxs := make(map[daghash.TxID]*walletTransaction)

	initialTxs, err := collectTransactions(client, gasLimitMap)
	if err != nil {
		return nil, nil, err
	}

	// Add all of the confirmed transaction outputs to the UTXO.
	for _, initialTx := range initialTxs {
		if initialTx.confirmed {
			addTxOutsToUTXOSet(walletUTXOSet, initialTx.tx.MsgTx())
		}
	}

	for _, initialTx := range initialTxs {
		// Remove all of the previous outpoints from the UTXO.
		// The previous outpoints are removed for unconfirmed
		// transactions as well, to avoid potential
		// double spends.
		removeTxInsFromUTXOSet(walletUTXOSet, initialTx.tx.MsgTx())

		// Add unconfirmed transactions to walletTxs, so we can
		// add their outputs to the UTXO when they are confirmed.
		if !initialTx.confirmed {
			walletTxs[*initialTx.tx.ID()] = initialTx
		}
	}

	return walletUTXOSet, walletTxs, nil
}

func updateWalletTxs(blockAdded *blockAddedMsg, walletTxs map[daghash.TxID]*walletTransaction) {
	for txID, walletTx := range walletTxs {
		if walletTx.checkConfirmationCountdown > 0 && walletTx.chainHeight < blockAdded.chainHeight {
			walletTx.checkConfirmationCountdown--
		}

		// Delete old confirmed transactions to save memory
		if walletTx.confirmed && walletTx.chainHeight+txLifeSpan < blockAdded.chainHeight {
			delete(walletTxs, txID)
		}
	}

	for _, tx := range blockAdded.txs {
		if _, ok := walletTxs[*tx.ID()]; !ok {
			walletTxs[*tx.ID()] = &walletTransaction{
				tx:                         tx,
				chainHeight:                blockAdded.chainHeight,
				checkConfirmationCountdown: requiredConfirmations,
			}
		}
	}
}

func randomWithAverageTarget(target float64) float64 {
	randomFraction := random.Float64()
	return randomFraction * target * 2
}

func randomIntegerWithAverageTarget(target uint64, allowZero bool) uint64 {
	randomNum := randomWithAverageTarget(float64(target))
	if !allowZero && randomNum < 1 {
		randomNum = 1
	}
	return uint64(math.Round(randomNum))
}

func createRandomTxFromFunds(walletUTXOSet utxoSet, cfg *configFlags, gasLimitMap map[subnetworkid.SubnetworkID]uint64, funds uint64) (tx *wire.MsgTx, isSecondaryAddress bool, err error) {
	if secondaryScriptPubKey != nil && !sentToSecondaryAddress && funds > minSecondaryTxAmount {
		tx, err = createTx(walletUTXOSet, minSecondaryTxAmount, cfg.AverageFeeRate, 1, 1, subnetworkid.SubnetworkIDNative, 0, 0, secondaryScriptPubKey)
		if err != nil {
			return nil, false, err
		}
		return tx, true, nil
	}

	payloadSize := uint64(0)
	gas := uint64(0)

	// In Go map iteration is randomized, so if we want
	// to choose a random element from a map we can
	// just take the first iterated element.
	chosenSubnetwork := subnetworkid.SubnetworkIDNative
	chosenGasLimit := uint64(0)
	for subnetworkID, gasLimit := range gasLimitMap {
		chosenSubnetwork = &subnetworkID
		chosenGasLimit = gasLimit
		break
	}

	if !chosenSubnetwork.IsEqual(subnetworkid.SubnetworkIDNative) {
		payloadSize = randomIntegerWithAverageTarget(cfg.AveragePayloadSize, true)
		gas = randomIntegerWithAverageTarget(uint64(float64(chosenGasLimit)*cfg.AverageGasFraction), true)
		if gas > chosenGasLimit {
			gas = chosenGasLimit
		}
	}

	targetNumberOfOutputs := randomIntegerWithAverageTarget(cfg.TargetNumberOfOutputs, false)
	targetNumberOfInputs := randomIntegerWithAverageTarget(cfg.TargetNumberOfInputs, false)

	feeRate := randomWithAverageTarget(cfg.AverageFeeRate)

	amount := minSpendableAmount + uint64(random.Int63n(int64(maxSpendableAmount-minSpendableAmount)))
	amount *= targetNumberOfOutputs
	if amount > funds-minTxFee {
		amount = funds - minTxFee
	}
	tx, err = createTx(walletUTXOSet, amount, feeRate, targetNumberOfOutputs, targetNumberOfInputs, chosenSubnetwork, payloadSize, gas, primaryScriptPubKey)
	if err != nil {
		return nil, false, err
	}
	return tx, true, nil
}

func enqueueTransactions(client *txgenClient, blockAdded *blockAddedMsg, walletUTXOSet utxoSet, walletTxs map[daghash.TxID]*walletTransaction,
	txChan chan *wire.MsgTx, cfg *configFlags, gasLimitMap map[subnetworkid.SubnetworkID]uint64) error {
	if err := applyConfirmedTransactionsAndResendNonAccepted(client, walletTxs, walletUTXOSet, blockAdded.chainHeight, txChan); err != nil {
		return err
	}

	for funds := calcUTXOSetFunds(walletUTXOSet); !isDust(funds); funds = calcUTXOSetFunds(walletUTXOSet) {
		tx, isSecondaryAddress, err := createRandomTxFromFunds(walletUTXOSet, cfg, gasLimitMap, funds)
		if err != nil {
			return err
		}
		txChan <- tx
		if isSecondaryAddress {
			sentToSecondaryAddress = true
		}
	}
	return nil
}

func createTx(walletUTXOSet utxoSet, minAmount uint64, feeRate float64, targetNumberOfOutputs uint64, targetNumberOfInputs uint64,
	subnetworkdID *subnetworkid.SubnetworkID, payloadSize uint64, gas uint64, scriptPubKey []byte) (*wire.MsgTx, error) {
	var tx *wire.MsgTx
	if subnetworkdID.IsEqual(subnetworkid.SubnetworkIDNative) {
		tx = wire.NewNativeMsgTx(wire.TxVersion, nil, nil)
	} else {
		payload := make([]byte, payloadSize)
		tx = wire.NewSubnetworkMsgTx(wire.TxVersion, nil, nil, subnetworkdID, gas, payload)
	}

	// Attempt to fund the transaction with spendable utxos.
	funds, err := fundTx(walletUTXOSet, tx, minAmount, feeRate, targetNumberOfOutputs, targetNumberOfInputs)
	if err != nil {
		return nil, err
	}

	maxNumOuts := funds / minSpendableAmount
	numOuts := targetNumberOfOutputs
	if numOuts > maxNumOuts {
		numOuts = maxNumOuts
	}

	fee := calcFee(tx, feeRate, numOuts, walletUTXOSet)
	funds -= fee

	for i := uint64(0); i < numOuts; i++ {
		tx.AddTxOut(&wire.TxOut{
			Value:        funds / numOuts,
			ScriptPubKey: scriptPubKey,
		})
	}

	err = signTx(walletUTXOSet, tx)
	if err != nil {
		return nil, err
	}

	removeTxInsFromUTXOSet(walletUTXOSet, tx)

	return tx, nil
}

// signTx signs a transaction
func signTx(walletUTXOSet utxoSet, tx *wire.MsgTx) error {
	for i, txIn := range tx.TxIn {
		outpoint := txIn.PreviousOutpoint
		prevOut := walletUTXOSet[outpoint]

		sigScript, err := txscript.SignatureScript(tx, i, prevOut.ScriptPubKey,
			txscript.SigHashAll, privateKey, true)
		if err != nil {
			return errors.Errorf("Failed to sign transaction: %s", err)
		}

		txIn.SignatureScript = sigScript
	}

	return nil
}

func fundTx(walletUTXOSet utxoSet, tx *wire.MsgTx, amount uint64, feeRate float64, targetNumberOfOutputs uint64, targetNumberOfInputs uint64) (uint64, error) {

	amountSelected := uint64(0)
	isTxFunded := false

	for outpoint, output := range walletUTXOSet {
		amountSelected += output.Value

		// Add the selected output to the transaction
		tx.AddTxIn(wire.NewTxIn(&outpoint, nil))

		// Check if transaction has enough funds. If we don't have enough
		// coins from he current amount selected to pay the fee, or we have
		// less inputs then the targeted amount, continue to grab more coins.
		isTxFunded = isFunded(tx, feeRate, targetNumberOfOutputs, amountSelected, amount, walletUTXOSet)
		if uint64(len(tx.TxIn)) >= targetNumberOfInputs && isTxFunded {
			break
		}
	}

	if !isTxFunded {
		return 0, errors.Errorf("not enough funds for coin selection")
	}

	return amountSelected, nil
}

// isFunded checks if the transaction has enough funds to cover the fee
// required for the txn.
func isFunded(tx *wire.MsgTx, feeRate float64, targetNumberOfOutputs uint64, amountSelected uint64, targetAmount uint64, walletUTXOSet utxoSet) bool {
	reqFee := calcFee(tx, feeRate, targetNumberOfOutputs, walletUTXOSet)
	return amountSelected > reqFee && amountSelected-reqFee >= targetAmount
}

func calcFee(msgTx *wire.MsgTx, feeRate float64, numberOfOutputs uint64, walletUTXOSet utxoSet) uint64 {
	txMass := calcTxMass(msgTx, walletUTXOSet)
	txMassWithOutputs := txMass + outputsTotalSize(numberOfOutputs)
	reqFee := uint64(float64(txMassWithOutputs) * feeRate)
	if reqFee < minTxFee {
		return minTxFee
	}
	return reqFee
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
		prevOut := walletUTXOSet[outpoint]
		previousScriptPubKeys[i] = prevOut.ScriptPubKey
	}
	return previousScriptPubKeys
}

func applyConfirmedTransactionsAndResendNonAccepted(client *txgenClient, walletTxs map[daghash.TxID]*walletTransaction, walletUTXOSet utxoSet,
	blockChainHeight uint64, txChan chan *wire.MsgTx) error {
	for txID, walletTx := range walletTxs {
		if !walletTx.confirmed && walletTx.checkConfirmationCountdown == 0 {
			txResult, err := client.GetRawTransactionVerbose(&txID)
			if err != nil {
				return err
			}
			msgTx := walletTx.tx.MsgTx()
			if isTxMatured(msgTx, *txResult.Confirmations) {
				walletTx.confirmed = true
				addTxOutsToUTXOSet(walletUTXOSet, msgTx)
			} else if !msgTx.IsCoinBase() && *txResult.Confirmations == 0 && !txResult.IsInMempool && blockChainHeight > walletTx.chainHeight+maxResendDepth {
				log.Infof("Transaction %s was not accepted in the DAG. Resending", txID)
				txChan <- msgTx
			}
		}
	}
	return nil
}

func removeTxInsFromUTXOSet(walletUTXOSet utxoSet, tx *wire.MsgTx) {
	for _, txIn := range tx.TxIn {
		delete(walletUTXOSet, txIn.PreviousOutpoint)
	}
}

func addTxOutsToUTXOSet(walletUTXOSet utxoSet, tx *wire.MsgTx) {
	for i, txOut := range tx.TxOut {
		if bytes.Equal(txOut.ScriptPubKey, primaryScriptPubKey) {
			outpoint := wire.Outpoint{TxID: *tx.TxID(), Index: uint32(i)}
			walletUTXOSet[outpoint] = txOut
		}
	}
}

func isTxMatured(tx *wire.MsgTx, confirmations uint64) bool {
	if !tx.IsCoinBase() {
		return confirmations >= requiredConfirmations
	}
	return confirmations >= approximateConfirmationsForCoinbaseMaturity
}

func calcUTXOSetFunds(walletUTXOSet utxoSet) uint64 {
	var funds uint64
	for _, output := range walletUTXOSet {
		funds += output.Value
	}
	return funds
}

func collectTransactions(client *txgenClient, gasLimitMap map[subnetworkid.SubnetworkID]uint64) (map[daghash.TxID]*walletTransaction, error) {
	registryTxs := make([]*util.Tx, 0)
	walletTxs := make(map[daghash.TxID]*walletTransaction)
	skip := 0
	for skip < searchRawTransactionMaxResults {
		results, err := client.SearchRawTransactionsVerbose(p2pkhAddress, skip, searchRawTransactionResultCount, true, true, nil)
		if err != nil {
			// Break when there are no further txs
			if rpcError, ok := err.(*kaspajson.RPCError); ok && rpcError.Code == kaspajson.ErrRPCNoTxInfo {
				break
			}

			return nil, err
		}

		for _, result := range results {
			// Mempool transactions and red block transactions bring about unnecessary complexity, so
			// simply don't bother processing them
			if result.IsInMempool || *result.Confirmations == 0 {
				continue
			}

			tx, err := parseRawTransactionResult(result)
			if err != nil {
				return nil, errors.Errorf("failed to process SearchRawTransactionResult: %s", err)
			}
			if tx == nil {
				continue
			}

			txID := tx.TxID()
			utilTx := util.NewTx(tx)

			if existingTx, ok := walletTxs[*txID]; !ok || !existingTx.confirmed {
				walletTxs[*txID] = &walletTransaction{
					tx:                         utilTx,
					checkConfirmationCountdown: requiredConfirmations,
					confirmed:                  isTxMatured(tx, *result.Confirmations),
				}
			}

			if tx.SubnetworkID.IsEqual(subnetworkid.SubnetworkIDRegistry) {
				registryTxs = append(registryTxs, utilTx)
			}
		}

		skip += searchRawTransactionResultCount
	}
	err := updateSubnetworks(registryTxs, gasLimitMap)
	if err != nil {
		return nil, err
	}
	return walletTxs, nil
}

func parseRawTransactionResult(result *kaspajson.SearchRawTransactionsResult) (*wire.MsgTx, error) {
	txBytes, err := hex.DecodeString(result.Hex)
	if err != nil {
		return nil, errors.Errorf("failed to decode transaction bytes: %s", err)
	}
	var tx wire.MsgTx
	reader := bytes.NewReader(txBytes)
	err = tx.Deserialize(reader)
	if err != nil {
		return nil, errors.Errorf("failed to deserialize transaction: %s", err)
	}
	return &tx, nil
}
