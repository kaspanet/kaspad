package main

import (
	"bytes"
	"encoding/hex"
	"github.com/daglabs/kaspad/btcjson"
	"github.com/daglabs/kaspad/rpcclient"
	"github.com/daglabs/kaspad/util"
	"github.com/daglabs/kaspad/wire"
	"github.com/pkg/errors"
)

const (
	resultsCount     = 1000
	minConfirmations = 10
)

func findUnspentTXO(cfg *ConfigFlags, client *rpcclient.Client, addrPubKeyHash *util.AddressPubKeyHash) (*wire.Outpoint, *wire.MsgTx, error) {
	txs, err := collectTransactions(client, addrPubKeyHash)
	if err != nil {
		return nil, nil, err
	}

	utxos := buildUTXOs(txs)
	for outpoint, tx := range utxos {
		// Skip TXOs that can't pay for registration
		if tx.TxOut[outpoint.Index].Value < cfg.RegistryTxFee {
			continue
		}

		return &outpoint, tx, nil
	}

	return nil, nil, nil
}

func collectTransactions(client *rpcclient.Client, addrPubKeyHash *util.AddressPubKeyHash) ([]*wire.MsgTx, error) {
	txs := make([]*wire.MsgTx, 0)
	skip := 0
	for {
		results, err := client.SearchRawTransactionsVerbose(addrPubKeyHash, skip, resultsCount, true, false, nil)
		if err != nil {
			// Break when there are no further txs
			if rpcError, ok := err.(*btcjson.RPCError); ok && rpcError.Code == btcjson.ErrRPCNoTxInfo {
				break
			}

			return nil, err
		}

		for _, result := range results {
			// Mempool transactions bring about unnecessary complexity, so
			// simply don't bother processing them
			if result.IsInMempool {
				continue
			}

			tx, err := parseRawTransactionResult(result)
			if err != nil {
				return nil, errors.Errorf("failed to process SearchRawTransactionResult: %s", err)
			}
			if tx == nil {
				continue
			}
			if !isTxMatured(tx, *result.Confirmations) {
				continue
			}

			txs = append(txs, tx)
		}

		skip += resultsCount
	}
	return txs, nil
}

func parseRawTransactionResult(result *btcjson.SearchRawTransactionsResult) (*wire.MsgTx, error) {
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

func isTxMatured(tx *wire.MsgTx, confirmations uint64) bool {
	if !tx.IsCoinBase() {
		return confirmations >= minConfirmations
	}
	return confirmations >= ActiveConfig().NetParams().BlockCoinbaseMaturity
}

func buildUTXOs(txs []*wire.MsgTx) map[wire.Outpoint]*wire.MsgTx {
	utxos := make(map[wire.Outpoint]*wire.MsgTx)
	for _, tx := range txs {
		for i := range tx.TxOut {
			outpoint := wire.NewOutpoint(tx.TxID(), uint32(i))
			utxos[*outpoint] = tx
		}
	}
	for _, tx := range txs {
		for _, input := range tx.TxIn {
			delete(utxos, input.PreviousOutpoint)
		}
	}
	return utxos
}
