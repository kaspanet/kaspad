package rpc

import (
	"encoding/hex"
	"github.com/kaspanet/kaspad/btcjson"
	"github.com/kaspanet/kaspad/txscript"
	"github.com/kaspanet/kaspad/util"
)

// handleDecodeScript handles decodeScript commands.
func handleDecodeScript(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*btcjson.DecodeScriptCmd)

	// Convert the hex script to bytes.
	hexStr := c.HexScript
	if len(hexStr)%2 != 0 {
		hexStr = "0" + hexStr
	}
	script, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, rpcDecodeHexError(hexStr)
	}

	// The disassembled string will contain [error] inline if the script
	// doesn't fully parse, so ignore the error here.
	disbuf, _ := txscript.DisasmString(script)

	// Get information about the script.
	// Ignore the error here since an error means the script couldn't parse
	// and there is no additinal information about it anyways.
	scriptClass, addr, _ := txscript.ExtractScriptPubKeyAddress(script,
		s.cfg.DAGParams)
	var address *string
	if addr != nil {
		address = btcjson.String(addr.EncodeAddress())
	}

	// Convert the script itself to a pay-to-script-hash address.
	p2sh, err := util.NewAddressScriptHash(script, s.cfg.DAGParams.Prefix)
	if err != nil {
		context := "Failed to convert script to pay-to-script-hash"
		return nil, internalRPCError(err.Error(), context)
	}

	// Generate and return the reply.
	reply := btcjson.DecodeScriptResult{
		Asm:     disbuf,
		Type:    scriptClass.String(),
		Address: address,
	}
	if scriptClass != txscript.ScriptHashTy {
		reply.P2sh = p2sh.EncodeAddress()
	}
	return reply, nil
}
