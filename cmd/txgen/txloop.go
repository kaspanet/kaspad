package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/mempool"
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
	skipTxCount int
	random      = rand.New(rand.NewSource(time.Now().UnixNano()))
	utxos       map[wire.OutPoint]*utxo
	pkScript    []byte
)

const (
	minSpendableAmount uint64 = 10000
	minRelayTxFee      uint64 = uint64(mempool.DefaultMinRelayTxFee)
)

func isDust(value uint64) bool {
	return value < minSpendableAmount+minRelayTxFee
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

// evalInputs scans all the passed inputs, destroying any utxos within the
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

func populateUtxos(client *rpcclient.Client) error {
	for atomic.LoadInt32(&isRunning) == 1 {
		arr, err := client.SearchRawTransactionsVerbose(pkHash, skipTxCount, 1, true, false, nil)
		if err != nil {
			log.Printf("SearchRawTransactionsVerbose failed: %s", err)
			log.Printf("Sleeping 30 sec...")
			for i := 0; i < 30; i++ {
				time.Sleep(time.Second)
				if atomic.LoadInt32(&isRunning) != 1 {
					return nil
				}
			}
			continue
		}
		searchResult := arr[0]
		txBytes, err := hex.DecodeString(searchResult.Hex)
		if err != nil {
			log.Printf("Failed to decode transactions bytes: %s", err)
			return err
		}
		txID, err := daghash.NewTxIDFromStr(searchResult.TxID)
		if err != nil {
			log.Printf("Failed to decode transaction ID: %s", err)
			return err
		}
		var tx wire.MsgTx
		rbuf := bytes.NewReader(txBytes)
		err = tx.Deserialize(rbuf)
		if err != nil {
			log.Printf("Failed to deserialize transaction: %s", err)
			return err
		}
		if !tx.IsBlockReward() {
			if searchResult.Confirmations < 1 {
				log.Printf("Got non-mined transaction, sleeping 1 sec")
				time.Sleep(time.Second)
				continue
			}
			evalOutputs(tx.TxOut, txID)
			evalInputs(tx.TxIn)
			skipTxCount++
			if utxosFunds() < minSpendableAmount+minRelayTxFee {
				continue
			}
			return nil
		}
		if searchResult.Confirmations < uint64(activeNetParams.BlockRewardMaturity) {
			loops := int(activeNetParams.BlockRewardMaturity) - int(searchResult.Confirmations)
			log.Printf("Got block reward transaction, which is not enough mature, sleeping %d sec", loops)
			for i := 0; i < loops; i++ {
				time.Sleep(time.Second)
				if atomic.LoadInt32(&isRunning) != 1 {
					return nil
				}
			}
			continue
		}
		evalOutputs(tx.TxOut, txID)
		evalInputs(tx.TxIn)
		skipTxCount++
		if utxosFunds() < minSpendableAmount+minRelayTxFee {
			continue
		}
		return nil
	}
	return nil
}

// fundTx attempts to fund a transaction sending amt bitcoin. The coins are
// selected such that the final amount spent pays enough fees as dictated by
// the passed fee rate. The passed fee rate should be expressed in
// satoshis-per-byte.
func fundTx(tx *wire.MsgTx, amt uint64, feeRate uint64) (uint64, error) {
	const (
		// spendSize is the largest number of bytes of a sigScript
		// which spends a p2pkh output: OP_DATA_73 <sig> OP_DATA_33 <pubkey>
		spendSize = 1 + 73 + 1 + 33
	)

	var (
		amtSelected uint64
		txSize      int
	)

	for outPoint, utxo := range utxos {
		if utxo.isLocked {
			continue
		}

		amtSelected += utxo.txOut.Value

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
		if amtSelected-reqFee < amt {
			continue
		}

		// If we have any change left over, then add an additional
		// output to the transaction reserved for change.
		changeVal := amtSelected - amt - reqFee
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

// createTransaction returns a fully signed transaction paying to the specified
// outputs while observing the desired fee rate. The passed fee rate should be
// expressed in satoshis-per-byte.
func createTransaction(outputs []*wire.TxOut, feeRate uint64) (*wire.MsgTx, uint64, error) {
	tx := wire.NewNativeMsgTx(wire.TxVersion, nil, nil)

	// Tally up the total amount to be sent in order to perform coin
	// selection shortly below.
	var outputAmt uint64
	for _, output := range outputs {
		outputAmt += output.Value
		tx.AddTxOut(output)
	}

	// Attempt to fund the transaction with spendable utxos.
	fees, err := fundTx(tx, outputAmt, feeRate)
	if err != nil {
		return nil, 0, err
	}

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
			log.Printf("Failed to sign transaction: %s", err)
			return nil, 0, err
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

	return tx, fees, nil
}

// txLoop performs main loop of transaction generation
func txLoop(clients []*rpcclient.Client) error {
	clientsCount := int64(len(clients))

	utxos = make(map[wire.OutPoint]*utxo)
	pkScript, err := txscript.PayToAddrScript(pkHash)
	if err != nil {
		log.Printf("Failed to generate pkscript to address: %s", err)
		return err
	}

	for atomic.LoadInt32(&isRunning) == 1 {
		err := populateUtxos(clients[0])
		if err != nil {
			return err
		}

		funds := utxosFunds()
		if funds < minSpendableAmount+minRelayTxFee {
			return nil
		}

		log.Printf("UTXO funds after population %d", funds)

		for funds > minSpendableAmount+minRelayTxFee {
			amount := minSpendableAmount + uint64(random.Int63n(int64(funds-minSpendableAmount)))
			output := wire.NewTxOut(amount, pkScript)

			tx, fees, err := createTransaction([]*wire.TxOut{output}, 10)

			if err != nil {
				log.Printf("Failed to create transaction (output value %d, funds %d): %s",
					amount, funds, err)
				break
			}

			log.Printf("Created transaction: amount %d, fees %d", amount, fees)

			funds = utxosFunds()

			var currentClient *rpcclient.Client
			if clientsCount == 1 {
				currentClient = clients[0]
			} else {
				currentClient = clients[random.Int63n(clientsCount)]
			}
			_, err = currentClient.SendRawTransaction(tx, true)
			if err != nil {
				log.Printf("Failed to send transaction: %s", err)
				return err
			}
		}
	}

	return nil
}
