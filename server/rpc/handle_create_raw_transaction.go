package rpc

import (
	"github.com/kaspanet/kaspad/rpcmodel"
	"github.com/kaspanet/kaspad/txscript"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
)

// handleCreateRawTransaction handles createRawTransaction commands.
func handleCreateRawTransaction(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*rpcmodel.CreateRawTransactionCmd)

	// Validate the locktime, if given.
	if c.LockTime != nil &&
		(*c.LockTime < 0 || *c.LockTime > wire.MaxTxInSequenceNum) {
		return nil, &rpcmodel.RPCError{
			Code:    rpcmodel.ErrRPCInvalidParameter,
			Message: "Locktime out of range",
		}
	}

	txIns := []*wire.TxIn{}
	// Add all transaction inputs to a new transaction after performing
	// some validity checks.
	for _, input := range c.Inputs {
		txID, err := daghash.NewTxIDFromStr(input.TxID)
		if err != nil {
			return nil, rpcDecodeHexError(input.TxID)
		}

		prevOut := wire.NewOutpoint(txID, input.Vout)
		txIn := wire.NewTxIn(prevOut, []byte{})
		if c.LockTime != nil && *c.LockTime != 0 {
			txIn.Sequence = wire.MaxTxInSequenceNum - 1
		}
		txIns = append(txIns, txIn)
	}
	mtx := wire.NewNativeMsgTx(wire.TxVersion, txIns, nil)

	// Add all transaction outputs to the transaction after performing
	// some validity checks.
	params := s.cfg.DAGParams
	for encodedAddr, amount := range c.Amounts {
		// Ensure amount is in the valid range for monetary amounts.
		if amount <= 0 || amount > util.MaxSompi {
			return nil, &rpcmodel.RPCError{
				Code:    rpcmodel.ErrRPCType,
				Message: "Invalid amount",
			}
		}

		// Decode the provided address.
		addr, err := util.DecodeAddress(encodedAddr, params.Prefix)
		if err != nil {
			return nil, &rpcmodel.RPCError{
				Code:    rpcmodel.ErrRPCInvalidAddressOrKey,
				Message: "Invalid address or key: " + err.Error(),
			}
		}

		// Ensure the address is one of the supported types and that
		// the network encoded with the address matches the network the
		// server is currently on.
		switch addr.(type) {
		case *util.AddressPubKeyHash:
		case *util.AddressScriptHash:
		default:
			return nil, &rpcmodel.RPCError{
				Code:    rpcmodel.ErrRPCInvalidAddressOrKey,
				Message: "Invalid address or key",
			}
		}
		if !addr.IsForPrefix(params.Prefix) {
			return nil, &rpcmodel.RPCError{
				Code: rpcmodel.ErrRPCInvalidAddressOrKey,
				Message: "Invalid address: " + encodedAddr +
					" is for the wrong network",
			}
		}

		// Create a new script which pays to the provided address.
		scriptPubKey, err := txscript.PayToAddrScript(addr)
		if err != nil {
			context := "Failed to generate pay-to-address script"
			return nil, internalRPCError(err.Error(), context)
		}

		// Convert the amount to satoshi.
		satoshi, err := util.NewAmount(amount)
		if err != nil {
			context := "Failed to convert amount"
			return nil, internalRPCError(err.Error(), context)
		}

		txOut := wire.NewTxOut(uint64(satoshi), scriptPubKey)
		mtx.AddTxOut(txOut)
	}

	// Set the Locktime, if given.
	if c.LockTime != nil {
		mtx.LockTime = *c.LockTime
	}

	// Return the serialized and hex-encoded transaction. Note that this
	// is intentionally not directly returning because the first return
	// value is a string and it would result in returning an empty string to
	// the client instead of nothing (nil) in the case of an error.
	mtxHex, err := messageToHex(mtx)
	if err != nil {
		return nil, err
	}
	return mtxHex, nil
}
