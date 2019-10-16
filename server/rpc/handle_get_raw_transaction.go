package rpc

import (
	"bytes"
	"encoding/hex"
	"github.com/daglabs/btcd/blockdag"
	"github.com/daglabs/btcd/btcjson"
	"github.com/daglabs/btcd/database"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/util/daghash"
	"github.com/daglabs/btcd/wire"
)

// handleGetRawTransaction implements the getRawTransaction command.
func handleGetRawTransaction(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*btcjson.GetRawTransactionCmd)

	// Convert the provided transaction hash hex to a Hash.
	txID, err := daghash.NewTxIDFromStr(c.TxID)
	if err != nil {
		return nil, rpcDecodeHexError(c.TxID)
	}

	verbose := false
	if c.Verbose != nil {
		verbose = *c.Verbose != 0
	}

	// Try to fetch the transaction from the memory pool and if that fails,
	// try the block database.
	var tx *util.Tx
	var blkHash *daghash.Hash
	var txMass uint64
	isInMempool := false
	mempoolTx, err := s.cfg.TxMemPool.FetchTransaction(txID)
	if err != nil {
		if s.cfg.TxIndex == nil {
			return nil, &btcjson.RPCError{
				Code: btcjson.ErrRPCNoTxInfo,
				Message: "The transaction index must be " +
					"enabled to query the blockchain " +
					"(specify --txindex)",
			}
		}

		txBytes, txBlockHash, err := fetchTxBytesAndBlockHashFromTxIndex(s, txID)
		if err != nil {
			return nil, err
		}

		// When the verbose flag isn't set, simply return the serialized
		// transaction as a hex-encoded string.  This is done here to
		// avoid deserializing it only to reserialize it again later.
		if !verbose {
			return hex.EncodeToString(txBytes), nil
		}

		// Grab the block hash.
		blkHash = txBlockHash

		// Deserialize the transaction
		var mtx wire.MsgTx
		err = mtx.Deserialize(bytes.NewReader(txBytes))
		if err != nil {
			context := "Failed to deserialize transaction"
			return nil, internalRPCError(err.Error(), context)
		}

		tx = util.NewTx(&mtx)
		txMass, err = calcTxMassFromTxIndex(s, tx)
		if err != nil {
			return nil, err
		}
	} else {
		// When the verbose flag isn't set, simply return the
		// network-serialized transaction as a hex-encoded string.
		if !verbose {
			// Note that this is intentionally not directly
			// returning because the first return value is a
			// string and it would result in returning an empty
			// string to the client instead of nothing (nil) in the
			// case of an error.
			mtxHex, err := messageToHex(mempoolTx.MsgTx())
			if err != nil {
				return nil, err
			}
			return mtxHex, nil
		}

		var err error
		txMass, err = blockdag.CalcTxMassFromUTXOSet(mempoolTx, s.cfg.DAG.UTXOSet())
		if err != nil {
			return nil, err
		}

		tx = mempoolTx
		isInMempool = true
	}

	// The verbose flag is set, so generate the JSON object and return it.
	var blkHeader *wire.BlockHeader
	var blkHashStr string
	if blkHash != nil {
		// Fetch the header from chain.
		header, err := s.cfg.DAG.HeaderByHash(blkHash)
		if err != nil {
			context := "Failed to fetch block header"
			return nil, internalRPCError(err.Error(), context)
		}

		blkHeader = header
		blkHashStr = blkHash.String()
	}

	confirmations, err := txConfirmations(s, tx.ID())
	if err != nil {
		return nil, err
	}
	rawTxn, err := createTxRawResult(s.cfg.DAGParams, tx.MsgTx(), txID.String(),
		blkHeader, blkHashStr, nil, &confirmations, isInMempool, txMass)
	if err != nil {
		return nil, err
	}
	return *rawTxn, nil
}

func calcTxMassFromTxIndex(s *Server, tx *util.Tx) (uint64, error) {
	if tx.IsCoinBase() {
		return blockdag.CalcTxMass(tx, nil), nil
	}
	previousScriptPubKeys := make([][]byte, len(tx.MsgTx().TxIn))
	for i, txIn := range tx.MsgTx().TxIn {
		txBytes, _, err := fetchTxBytesAndBlockHashFromTxIndex(s, &txIn.PreviousOutpoint.TxID)
		if err != nil {
			return 0, err
		}
		var msgTx wire.MsgTx
		err = msgTx.Deserialize(bytes.NewReader(txBytes))
		if err != nil {
			context := "Failed to deserialize transaction"
			return 0, internalRPCError(err.Error(), context)
		}
		previousScriptPubKeys[i] = msgTx.TxOut[txIn.PreviousOutpoint.Index].ScriptPubKey
	}
	return blockdag.CalcTxMass(tx, previousScriptPubKeys), nil
}

func fetchTxBytesAndBlockHashFromTxIndex(s *Server, txID *daghash.TxID) ([]byte, *daghash.Hash, error) {
	blockRegion, err := s.cfg.TxIndex.TxFirstBlockRegion(txID)
	if err != nil {
		context := "Failed to retrieve transaction location"
		return nil, nil, internalRPCError(err.Error(), context)
	}
	if blockRegion == nil {
		return nil, nil, rpcNoTxInfoError(txID)
	}

	// Load the raw transaction bytes from the database.
	var txBytes []byte
	err = s.cfg.DB.View(func(dbTx database.Tx) error {
		var err error
		txBytes, err = dbTx.FetchBlockRegion(blockRegion)
		return err
	})
	if err != nil {
		return nil, nil, rpcNoTxInfoError(txID)
	}
	return txBytes, blockRegion.Hash, nil
}
