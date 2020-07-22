package rpc

import (
	"encoding/hex"
	"fmt"
	"github.com/kaspanet/kaspad/rpc/model"
	"github.com/kaspanet/kaspad/txscript"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/pointers"
	"github.com/kaspanet/kaspad/wire"
)

// handleGetTxOut handles getTxOut commands.
func handleGetTxOut(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*model.GetTxOutCmd)

	// Convert the provided transaction hash hex to a Hash.
	txID, err := daghash.NewTxIDFromStr(c.TxID)
	if err != nil {
		return nil, rpcDecodeHexError(c.TxID)
	}

	// If requested and the tx is available in the mempool try to fetch it
	// from there, otherwise attempt to fetch from the block database.
	var selectedTipHash string
	var confirmations *uint64
	var value uint64
	var scriptPubKey []byte
	var isCoinbase bool
	isInMempool := false
	includeMempool := true
	if c.IncludeMempool != nil {
		includeMempool = *c.IncludeMempool
	}
	// TODO: This is racy. It should attempt to fetch it directly and check
	// the error.
	if includeMempool && s.txMempool.HaveTransaction(txID) {
		tx, err := s.txMempool.FetchTransaction(txID)
		if err != nil {
			return nil, rpcNoTxInfoError(txID)
		}

		mtx := tx.MsgTx()
		if c.Vout > uint32(len(mtx.TxOut)-1) {
			return nil, &model.RPCError{
				Code: model.ErrRPCInvalidTxVout,
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

		selectedTipHash = s.dag.SelectedTipHash().String()
		value = txOut.Value
		scriptPubKey = txOut.ScriptPubKey
		isCoinbase = mtx.IsCoinBase()
		isInMempool = true
	} else {
		out := wire.Outpoint{TxID: *txID, Index: c.Vout}
		entry, ok := s.dag.GetUTXOEntry(out)
		if !ok {
			return nil, rpcNoTxInfoError(txID)
		}

		// To match the behavior of the reference client, return nil
		// (JSON null) if the transaction output is spent by another
		// transaction already in the DAG. Mined transactions
		// that are spent by a mempool transaction are not affected by
		// this.
		if entry == nil {
			return nil, nil
		}

		utxoConfirmations, ok := s.dag.UTXOConfirmations(&out)
		if !ok {
			errStr := fmt.Sprintf("Cannot get confirmations for tx id %s, index %d",
				out.TxID, out.Index)
			return nil, internalRPCError(errStr, "")
		}
		confirmations = &utxoConfirmations

		selectedTipHash = s.dag.SelectedTipHash().String()
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
		s.dag.Params)
	var address *string
	if addr != nil {
		address = pointers.String(addr.EncodeAddress())
	}

	txOutReply := &model.GetTxOutResult{
		SelectedTip:   selectedTipHash,
		Confirmations: confirmations,
		IsInMempool:   isInMempool,
		Value:         util.Amount(value).ToKAS(),
		ScriptPubKey: model.ScriptPubKeyResult{
			Asm:     disbuf,
			Hex:     hex.EncodeToString(scriptPubKey),
			Type:    scriptClass.String(),
			Address: address,
		},
		Coinbase: isCoinbase,
	}
	return txOutReply, nil
}
