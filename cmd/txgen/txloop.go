package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/rpcclient"
	"github.com/daglabs/btcd/txscript"
	"github.com/daglabs/btcd/wire"
)

// utxo represents an unspent output spendable by the memWallet. The maturity
// height of the transaction is recorded in order to properly observe the
// maturity period of direct coinbase outputs.
type utxo struct {
	txOut    *wire.TxOut
	isLocked bool
}

var (
	random   = rand.New(rand.NewSource(time.Now().UnixNano()))
	utxos    map[wire.OutPoint]*utxo
	pkScript []byte
	spentTxs map[daghash.TxID]bool
)

const (
	// Those constants should be updated, when monetary policy changed
	minSpendableAmount uint64 = 10000
	minTxFee           uint64 = 3000
)

func isDust(value uint64) bool {
	return value < minSpendableAmount+minTxFee
}

// evalOutputs evaluates each of the passed outputs, creating a new matching
// utxo within the wallet if we're able to spend the output.
func evalOutputs(outputs []*wire.TxOut, txID *daghash.TxID) {
	for i, output := range outputs {
		if isDust(output.Value) {
			continue
		}
		op := wire.OutPoint{TxID: *txID, Index: uint32(i)}
		utxos[op] = &utxo{txOut: output}
	}
}

// evalInputs scans all the passed inputs, deleting any utxos within the
// wallet which are spent by an input.
func evalInputs(inputs []*wire.TxIn) {
	for _, txIn := range inputs {
		op := txIn.PreviousOutPoint
		if _, ok := utxos[op]; ok {
			delete(utxos, op)
		}
	}
}

func utxosFunds() uint64 {
	var funds uint64
	for _, utxo := range utxos {
		if utxo.isLocked {
			continue
		}
		funds += utxo.txOut.Value
	}
	return funds
}

func isTxMatured(tx *wire.MsgTx, confirmations uint64) bool {
	if !tx.IsBlockReward() {
		return confirmations >= 1
	}
	return confirmations >= uint64(float64(activeNetParams.BlockRewardMaturity)*1.5)
}

// DumpTx logs out transaction with given header
func DumpTx(header string, tx *wire.MsgTx) {
	logger.Info(header)
	logger.Infof("\tInputs:")
	for i, txIn := range tx.TxIn {
		asm, _ := txscript.DisasmString(txIn.SignatureScript)
		logger.Infof("\t\t%d: PreviousOutPoint: %v, SignatureScript: %s",
			i, txIn.PreviousOutPoint, asm)
	}
	logger.Infof("\tOutputs:")
	for i, txOut := range tx.TxOut {
		asm, _ := txscript.DisasmString(txOut.PkScript)
		logger.Infof("\t\t%d: Value: %d, PkScript: %s", i, txOut.Value, asm)
	}
}

func fetchAndPopulateUtxos(client *rpcclient.Client) (funds uint64, exit bool, err error) {
	skipCount := 0
	for atomic.LoadInt32(&isRunning) == 1 {
		arr, err := client.SearchRawTransactionsVerbose(p2pkhAddress, skipCount, 1000, true, false, nil)
		if err != nil {
			logger.Infof("No spandable transactions found and SearchRawTransactionsVerbose failed: %s", err)
			funds := utxosFunds()
			if !isDust(funds) {
				// we have something to spend
				logger.Infof("We have enough funds to generate transactions: %d", funds)
				return funds, false, nil
			}
			logger.Infof("Sleeping 30 sec...")
			for i := 0; i < 30; i++ {
				time.Sleep(time.Second)
				if atomic.LoadInt32(&isRunning) != 1 {
					return 0, true, nil
				}
			}
			skipCount = 0
			continue
		}
		receivedCount := len(arr)
		skipCount += receivedCount
		logger.Infof("Received %d transactions", receivedCount)
		for _, searchResult := range arr {
			txBytes, err := hex.DecodeString(searchResult.Hex)
			if err != nil {
				logger.Warnf("Failed to decode transactions bytes: %s", err)
				continue
			}
			txID, err := daghash.NewTxIDFromStr(searchResult.TxID)
			if err != nil {
				logger.Warnf("Failed to decode transaction ID: %s", err)
				continue
			}
			var tx wire.MsgTx
			rbuf := bytes.NewReader(txBytes)
			err = tx.Deserialize(rbuf)
			if err != nil {
				logger.Warnf("Failed to deserialize transaction: %s", err)
				continue
			}
			if spentTxs[*txID] {
				continue
			}
			if isTxMatured(&tx, searchResult.Confirmations) {
				spentTxs[*txID] = true
				evalOutputs(tx.TxOut, txID)
				evalInputs(tx.TxIn)
			}
		}
	}
	return 0, true, nil
}

// fundTx attempts to fund a transaction sending amount bitcoin. The coins are
// selected such that the final amount spent pays enough fees as dictated by
// the passed fee rate. The passed fee rate should be expressed in
// satoshis-per-byte.
func fundTx(tx *wire.MsgTx, amount uint64, feeRate uint64) (uint64, error) {
	const (
		// spendSize is the largest number of bytes of a sigScript
		// which spends a p2pkh output: OP_DATA_73 <sig> OP_DATA_33 <pubkey>
		spendSize = 1 + 73 + 1 + 33
	)

	var (
		amountSelected uint64
		txSize         int
	)

	for outPoint, utxo := range utxos {
		if utxo.isLocked {
			continue
		}

		amountSelected += utxo.txOut.Value

		// Add the selected output to the transaction, updating the
		// current tx size while accounting for the size of the future
		// sigScript.
		tx.AddTxIn(wire.NewTxIn(&outPoint, nil))
		txSize = tx.SerializeSize() + spendSize*len(tx.TxIn)

		// Calculate the fee required for the txn at this point
		// observing the specified fee rate. If we don't have enough
		// coins from he current amount selected to pay the fee, then
		// continue to grab more coins.
		reqFee := uint64(txSize) * feeRate
		if amountSelected-reqFee < amount {
			continue
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

	// If we've reached this point, then coin selection failed due to an
	// insufficient amount of coins.
	return 0, fmt.Errorf("not enough funds for coin selection")
}

// signTxAndLockSpentUtxo signs new transaction and locks spentutxo
func signTxAndLockSpentUtxo(tx *wire.MsgTx) error {
	// Populate all the selected inputs with valid sigScript for spending.
	// Along the way record all outputs being spent in order to avoid a
	// potential double spend.
	spentOutputs := make([]*utxo, 0, len(tx.TxIn))
	for i, txIn := range tx.TxIn {
		outPoint := txIn.PreviousOutPoint
		utxo := utxos[outPoint]
		txOut := utxo.txOut

		sigScript, err := txscript.SignatureScript(tx, i, txOut.PkScript,
			txscript.SigHashAll, privateKey, true)
		if err != nil {
			logger.Warnf("Failed to sign transaction: %s", err)
			return err
		}

		txIn.SignatureScript = sigScript

		spentOutputs = append(spentOutputs, utxo)
	}

	// As these outputs are now being spent by this newly created
	// transaction, mark the outputs are "locked". This action ensures
	// these outputs won't be double spent by any subsequent transactions.
	// These locked outputs can be freed via a call to UnlockOutputs.
	for _, utxo := range spentOutputs {
		utxo.isLocked = true
	}

	return nil
}

// createTransaction returns a fully signed transaction paying to the specified
// outputs while observing the desired fee rate. The passed fee rate should be
// expressed in satoshis-per-byte.
func createTransaction(outputs []*wire.TxOut, feeRate uint64) (*wire.MsgTx, uint64, error) {
	tx := wire.NewNativeMsgTx(wire.TxVersion, nil, nil)

	// Tally up the total amount to be sent in order to perform coin
	// selection shortly below.
	var outputAmount uint64
	for _, output := range outputs {
		outputAmount += output.Value
		tx.AddTxOut(output)
	}

	// Attempt to fund the transaction with spendable utxos.
	fees, err := fundTx(tx, outputAmount, feeRate)
	if err != nil {
		return nil, 0, err
	}

	err = signTxAndLockSpentUtxo(tx)
	if err != nil {
		return nil, 0, err
	}

	return tx, fees, nil
}

// txLoop performs main loop of transaction generation
func txLoop(clients []*rpcclient.Client) {
	clientsCount := int64(len(clients))

	utxos = make(map[wire.OutPoint]*utxo)
	spentTxs = make(map[daghash.TxID]bool)

	var err error
	pkScript, err = txscript.PayToAddrScript(p2pkhAddress)

	if err != nil {
		logger.Warnf("Failed to generate pkscript to address: %s", err)
		return
	}

	for atomic.LoadInt32(&isRunning) == 1 {
		funds, exit, err := fetchAndPopulateUtxos(clients[0])
		if exit {
			return
		}
		if err != nil {
			logger.Warnf("fetchAndPopulateUtxos failed: %s", err)
			continue
		}

		if isDust(funds) {
			logger.Warnf("fetchAndPopulateUtxos returned not enough funds")
			continue
		}

		logger.Infof("UTXO funds after population %d", funds)

		for !isDust(funds) {
			amount := minSpendableAmount + uint64(random.Int63n(int64(minSpendableAmount*4)))
			if amount > funds-minTxFee {
				amount = funds - minTxFee
			}
			output := wire.NewTxOut(amount, pkScript)

			tx, fees, err := createTransaction([]*wire.TxOut{output}, 10)

			if err != nil {
				logger.Warnf("Failed to create transaction (output value %d, funds %d): %s",
					amount, funds, err)
				continue
			}

			logger.Infof("Created transaction %s: amount %d, fees %d", tx.TxID(), amount, fees)

			funds = utxosFunds()
			logger.Infof("Remaining funds: %d", funds)

			var currentClient *rpcclient.Client
			if clientsCount == 1 {
				currentClient = clients[0]
			} else {
				currentClient = clients[random.Int63n(clientsCount)]
			}
			_, err = currentClient.SendRawTransaction(tx, true)
			if err != nil {
				logger.Warnf("Failed to send transaction: %s", err)
				continue
			}
		}
	}
}
