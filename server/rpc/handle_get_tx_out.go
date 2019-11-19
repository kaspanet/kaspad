package rpc

import (
	"encoding/hex"
	"fmt"
	"github.com/daglabs/btcd/btcjson"
	"github.com/daglabs/btcd/txscript"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/util/daghash"
	"github.com/daglabs/btcd/wire"
)

// handleGetTxOut handles getTxOut commands.
func handleGetTxOut(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*btcjson.GetTxOutCmd)

	// Convert the provided transaction hash hex to a Hash.
	txID, err := daghash.NewTxIDFromStr(c.TxID)
	if err != nil {
		return nil, rpcDecodeHexError(c.TxID)
	}

	// If requested and the tx is available in the mempool try to fetch it
	// from there, otherwise attempt to fetch from the block database.
	var bestBlockHash string
	var confirmations *uint64
	var value uint64
	var scriptPubKey []byte
	var isCoinbase bool
	isInMempool := false
	includeMempool := true
	if c.IncludeMempool != nil {
		includeMempool = *c.IncludeMempool
	}
	// TODO: This is racy.  It should attempt to fetch it directly and check
	// the error.
	if includeMempool && s.cfg.TxMemPool.HaveTransaction(txID) {
		tx, err := s.cfg.TxMemPool.FetchTransaction(txID)
		if err != nil {
			return nil, rpcNoTxInfoError(txID)
		}

		mtx := tx.MsgTx()
		if c.Vout > uint32(len(mtx.TxOut)-1) {
			return nil, &btcjson.RPCError{
				Code: btcjson.ErrRPCInvalidTxVout,
				Message: "Output index number (vout) does not " +
					"exist for transaction.",
			}
		}

		txOut := mtx.TxOut[c.Vout]
		if txOut == nil {
			errStr := fmt.Sprintf("Output index: %d for txid: %s "+
				"does not exist", c.Vout, txID)
			return nil, internalRPCError(errStr, "")
		}

		bestBlockHash = s.cfg.DAG.SelectedTipHash().String()
		value = txOut.Value
		scriptPubKey = txOut.ScriptPubKey
		isCoinbase = mtx.IsCoinBase()
		isInMempool = true
	} else {
		out := wire.Outpoint{TxID: *txID, Index: c.Vout}
		entry, ok := s.cfg.DAG.GetUTXOEntry(out)
		if !ok {
			return nil, rpcNoTxInfoError(txID)
		}

		// To match the behavior of the reference client, return nil
		// (JSON null) if the transaction output is spent by another
		// transaction already in the main chain.  Mined transactions
		// that are spent by a mempool transaction are not affected by
		// this.
		if entry == nil {
			return nil, nil
		}

		if s.cfg.TxIndex != nil {
			txConfirmations, err := txConfirmationsWithLock(s, txID)
			if err != nil {
				return nil, internalRPCError("Output index number (vout) does not "+
					"exist for transaction.", "")
			}
			confirmations = &txConfirmations
		}

		bestBlockHash = s.cfg.DAG.SelectedTipHash().String()
		value = entry.Amount()
		scriptPubKey = entry.ScriptPubKey()
		isCoinbase = entry.IsCoinbase()
	}

	// Disassemble script into single line printable format.
	// The disassembled string will contain [error] inline if the script
	// doesn't fully parse, so ignore the error here.
	disbuf, _ := txscript.DisasmString(scriptPubKey)

	// Get further info about the script.
	// Ignore the error here since an error means the script couldn't parse
	// and there is no additional information about it anyways.
	scriptClass, addr, _ := txscript.ExtractScriptPubKeyAddress(scriptPubKey,
		s.cfg.DAGParams)
	var address *string
	if addr != nil {
		address = btcjson.String(addr.EncodeAddress())
	}

	txOutReply := &btcjson.GetTxOutResult{
		BestBlock:     bestBlockHash,
		Confirmations: confirmations,
		IsInMempool:   isInMempool,
		Value:         util.Amount(value).ToBTC(),
		ScriptPubKey: btcjson.ScriptPubKeyResult{
			Asm:     disbuf,
			Hex:     hex.EncodeToString(scriptPubKey),
			Type:    scriptClass.String(),
			Address: address,
		},
		Coinbase: isCoinbase,
	}
	return txOutReply, nil
}
