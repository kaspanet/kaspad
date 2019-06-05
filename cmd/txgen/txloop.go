package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"

	"github.com/daglabs/btcd/btcjson"
	"github.com/daglabs/btcd/txscript"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/util/daghash"
	"github.com/daglabs/btcd/wire"
)

const (
	// Those constants should be updated, when monetary policy changed
	minSpendableAmount uint64 = 10000
	maxSpendableAmount uint64 = 5 * minSpendableAmount
	minTxFee           uint64 = 3000

	// spendSize is the largest number of bytes of a sigScript
	// which spends a p2pkh output: OP_DATA_73 <sig> OP_DATA_33 <pubkey>
	spendSize = 1 + 73 + 1 + 33

	txLifeSpan                                     = 1000
	requiredConfirmations                          = 10
	approximateConfirmationsForBlockRewardMaturity = 150
	searchRawTransactionResultCount                = 1000
	searchRawTransactionMaxResults                 = 5000
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
	random   = rand.New(rand.NewSource(time.Now().UnixNano()))
	pkScript []byte
)

// txLoop performs main loop of transaction generation
func txLoop(client *txgenClient) error {
	var err error
	pkScript, err = txscript.PayToAddrScript(p2pkhAddress)

	if err != nil {
		return fmt.Errorf("failed to generate pkScript to address: %s", err)
	}

	err = client.LoadTxFilter(true, []util.Address{p2pkhAddress}, nil)
	if err != nil {
		return err
	}

	walletUTXOSet, walletTxs, err := getInitialUTXOSetAndWalletTxs(client)
	if err != nil {
		return err
	}

	for blockAdded := range client.onBlockAdded {
		log.Infof("Block %s Added with %d relevant transactions", blockAdded.header.BlockHash(), len(blockAdded.txs))
		updateWalletTxs(blockAdded, walletTxs)
		err := sendTransactions(client, blockAdded, walletUTXOSet, walletTxs)
		if err != nil {
			return err
		}
	}
	return nil
}

func getInitialUTXOSetAndWalletTxs(client *txgenClient) (utxoSet, map[daghash.TxID]*walletTransaction, error) {
	walletUTXOSet := make(utxoSet)
	walletTxs := make(map[daghash.TxID]*walletTransaction)

	initialTxs, err := collectTransactions(client)
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

func sendTransactions(client *txgenClient, blockAdded *blockAddedMsg, walletUTXOSet utxoSet, walletTxs map[daghash.TxID]*walletTransaction) error {
	if err := checkConfirmations(client, walletTxs, walletUTXOSet, blockAdded.chainHeight); err != nil {
		return err
	}

	for funds := calcUTXOSetFunds(walletUTXOSet); !isDust(funds); funds = calcUTXOSetFunds(walletUTXOSet) {
		amount := minSpendableAmount + uint64(random.Int63n(int64(maxSpendableAmount-minSpendableAmount)))
		if amount > funds-minTxFee {
			amount = funds - minTxFee
		}
		output := wire.NewTxOut(amount, pkScript)
		tx, _, err := createTx(walletUTXOSet, []*wire.TxOut{output}, 0)
		if err != nil {
			return err
		}

		_, err = client.SendRawTransaction(tx, true)
		log.Infof("Sending tx %s", tx.TxID())
		if err != nil {
			return err
		}
	}
	return nil
}

func createTx(walletUTXOSet utxoSet, outputs []*wire.TxOut, feeRate uint64) (*wire.MsgTx, uint64, error) {
	tx := wire.NewNativeMsgTx(wire.TxVersion, nil, nil)

	// Tally up the total amount to be sent in order to perform coin
	// selection shortly below.
	var outputAmount uint64
	for _, output := range outputs {
		outputAmount += output.Value
		tx.AddTxOut(output)
	}

	// Attempt to fund the transaction with spendable utxos.
	fees, err := fundTx(walletUTXOSet, tx, outputAmount, feeRate)
	if err != nil {
		return nil, 0, err
	}

	err = signTx(walletUTXOSet, tx)
	if err != nil {
		return nil, 0, err
	}

	removeTxInsFromUTXOSet(walletUTXOSet, tx)

	return tx, fees, nil
}

// signTx signs a transaction
func signTx(walletUTXOSet utxoSet, tx *wire.MsgTx) error {
	for i, txIn := range tx.TxIn {
		outpoint := txIn.PreviousOutpoint
		prevOut := walletUTXOSet[outpoint]

		sigScript, err := txscript.SignatureScript(tx, i, prevOut.PkScript,
			txscript.SigHashAll, privateKey, true)
		if err != nil {
			return fmt.Errorf("Failed to sign transaction: %s", err)
		}

		txIn.SignatureScript = sigScript
	}

	return nil
}

func fundTx(walletUTXOSet utxoSet, tx *wire.MsgTx, amount uint64, feeRate uint64) (uint64, error) {

	var (
		amountSelected uint64
		txSize         int
		reqFee         uint64
	)

	isFunded := false

	for outpoint, output := range walletUTXOSet {
		amountSelected += output.Value

		// Add the selected output to the transaction, updating the
		// current tx size while accounting for the size of the future
		// sigScript.
		tx.AddTxIn(wire.NewTxIn(&outpoint, nil))
		txSize = tx.SerializeSize() + spendSize*len(tx.TxIn)

		// Calculate the fee required for the txn at this point
		// observing the specified fee rate. If we don't have enough
		// coins from he current amount selected to pay the fee, then
		// continue to grab more coins.
		reqFee = uint64(txSize) * feeRate
		if reqFee < minTxFee {
			reqFee = minTxFee
		}
		if amountSelected > reqFee && amountSelected-reqFee >= amount {
			isFunded = true
			break
		}
	}

	if !isFunded {
		return 0, fmt.Errorf("not enough funds for coin selection")
	}

	// If we have any change left over, then add an additional
	// output to the transaction reserved for change.
	changeVal := amountSelected - amount - reqFee
	if changeVal > 0 {
		changeOutput := &wire.TxOut{
			Value:    changeVal,
			PkScript: pkScript,
		}
		tx.AddTxOut(changeOutput)
	}

	return reqFee, nil
}

func checkConfirmations(client *txgenClient, walletTxs map[daghash.TxID]*walletTransaction, walletUTXOSet utxoSet, blockChainHeight uint64) error {
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
			} else if *txResult.Confirmations == 0 && !txResult.IsInMempool && blockChainHeight-500 > walletTx.chainHeight {
				log.Infof("Transaction %s was not accepted in the DAG. Resending", txID)
				_, err := client.SendRawTransaction(msgTx, true)
				if err != nil {
					return err
				}
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
		outpoint := wire.Outpoint{TxID: *tx.TxID(), Index: uint32(i)}
		walletUTXOSet[outpoint] = txOut
	}
}

func isTxMatured(tx *wire.MsgTx, confirmations uint64) bool {
	if !tx.IsBlockReward() {
		return confirmations >= requiredConfirmations
	}
	return confirmations >= approximateConfirmationsForBlockRewardMaturity
}

func calcUTXOSetFunds(walletUTXOSet utxoSet) uint64 {
	var funds uint64
	for _, output := range walletUTXOSet {
		funds += output.Value
	}
	return funds
}

func collectTransactions(client *txgenClient) (map[daghash.TxID]*walletTransaction, error) {
	walletTxs := make(map[daghash.TxID]*walletTransaction)
	skip := 0
	for skip < searchRawTransactionMaxResults {
		results, err := client.SearchRawTransactionsVerbose(p2pkhAddress, skip, searchRawTransactionResultCount, true, true, nil)
		if err != nil {
			// Break when there are no further txs
			if rpcError, ok := err.(*btcjson.RPCError); ok && rpcError.Code == btcjson.ErrRPCNoTxInfo {
				break
			}

			return nil, err
		}

		for _, result := range results {
			// Mempool transactions and red block transactions bring about unnecessary complexity, so
			// simply don't bother processing them
			if *result.Confirmations == 0 {
				continue
			}

			tx, err := parseRawTransactionResult(result)
			if err != nil {
				return nil, fmt.Errorf("failed to process SearchRawTransactionResult: %s", err)
			}
			if tx == nil {
				continue
			}

			txID := tx.TxID()

			if existingTx, ok := walletTxs[*txID]; !ok || !existingTx.confirmed {
				walletTxs[*txID] = &walletTransaction{
					tx:                         util.NewTx(tx),
					checkConfirmationCountdown: requiredConfirmations,
					confirmed:                  isTxMatured(tx, *result.Confirmations),
				}
			}
		}

		skip += searchRawTransactionResultCount
	}
	return walletTxs, nil
}

func parseRawTransactionResult(result *btcjson.SearchRawTransactionsResult) (*wire.MsgTx, error) {
	txBytes, err := hex.DecodeString(result.Hex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode transaction bytes: %s", err)
	}
	var tx wire.MsgTx
	reader := bytes.NewReader(txBytes)
	err = tx.Deserialize(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize transaction: %s", err)
	}
	return &tx, nil
}
