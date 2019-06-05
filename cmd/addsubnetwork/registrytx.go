package main

import (
	"fmt"
	"github.com/daglabs/btcd/btcec"
	"github.com/daglabs/btcd/txscript"
	"github.com/daglabs/btcd/wire"
)

func buildSubnetworkRegistryTx(cfg *config, fundingOutPoint *wire.OutPoint, fundingTx *wire.MsgTx, privateKey *btcec.PrivateKey) (*wire.MsgTx, error) {
	txIn := &wire.TxIn{
		PreviousOutPoint: *fundingOutPoint,
		Sequence:         wire.MaxTxInSequenceNum,
	}
	txOut := &wire.TxOut{
		PkScript: fundingTx.TxOut[fundingOutPoint.Index].PkScript,
		Value:    fundingTx.TxOut[fundingOutPoint.Index].Value - cfg.RegistryTxFee,
	}
	registryTx := wire.NewRegistryMsgTx(1, []*wire.TxIn{txIn}, []*wire.TxOut{txOut}, cfg.GasLimit)

	SignatureScript, err := txscript.SignatureScript(registryTx, 0, fundingTx.TxOut[fundingOutPoint.Index].PkScript,
		txscript.SigHashAll, privateKey, true)
	if err != nil {
		return nil, fmt.Errorf("failed to build signature script: %s", err)
	}
	txIn.SignatureScript = SignatureScript

	return registryTx, nil
}
