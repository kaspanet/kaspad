// Copyright (c) 2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package mining

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/daglabs/btcd/blockdag"
	"github.com/daglabs/btcd/btcec"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/txscript"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/util/daghash"
	"github.com/daglabs/btcd/wire"
)

// newHashFromStr converts the passed big-endian hex string into a
// daghash.Hash.  It only differs from the one available in daghash in that
// it panics on an error since it will only (and must only) be called with
// hard-coded, and therefore known good, hashes.
func newHashFromStr(hexStr string) *daghash.Hash {
	hash, err := daghash.NewHashFromStr(hexStr)
	if err != nil {
		panic("invalid hash in source file: " + hexStr)
	}
	return hash
}

// hexToBytes converts the passed hex string into bytes and will panic if there
// is an error.  This is only provided for the hard-coded constants so errors in
// the source code can be detected. It will only (and must only) be called with
// hard-coded values.
func hexToBytes(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic("invalid hex in source file: " + s)
	}
	return b
}

// newUTXOSet returns a new utxo view populated with outputs of the
// provided source transactions as if there were available at the respective
// block height specified in the heights slice.  The length of the source txns
// and source tx heights must match or it will panic.
func newUTXOSet(sourceTxns []*wire.MsgTx, sourceTxHeights []uint64) blockdag.UTXOSet {
	if len(sourceTxns) != len(sourceTxHeights) {
		panic("each transaction must have its block height specified")
	}

	utxoSet := blockdag.NewFullUTXOSet()
	for i, tx := range sourceTxns {
		if isAccepted, err := utxoSet.AddTx(tx, sourceTxHeights[i]); err != nil {
			panic(fmt.Sprintf("AddTx unexpectedly failed. Error: %s", err))
		} else if !isAccepted {
			panic(fmt.Sprintf("AddTx unexpectedly didn't add tx %s", tx.TxID()))
		}
	}
	return utxoSet
}

func createTxIn(originTx *wire.MsgTx, outputIndex uint32) *wire.TxIn {
	var prevOut *wire.Outpoint
	if originTx != nil {
		prevOut = wire.NewOutpoint(originTx.TxID(), 0)
	} else {
		prevOut = &wire.Outpoint{
			TxID:  daghash.TxID{},
			Index: 0xFFFFFFFF,
		}
	}
	return wire.NewTxIn(prevOut, nil)
}

func createTransaction(value uint64, originTx *wire.MsgTx, originTxoutputIndex uint32, sigScript []byte) (*wire.MsgTx, error) {
	lookupKey := func(a util.Address) (*btcec.PrivateKey, bool, error) {
		// Ordinarily this function would involve looking up the private
		// key for the provided address, but since the only thing being
		// signed in this example uses the address associated with the
		// private key from above, simply return it with the compressed
		// flag set since the address is using the associated compressed
		// public key.
		return privKey, true, nil
	}

	txIn := createTxIn(originTx, originTxoutputIndex)
	pkScript, err := txscript.PayToAddrScript(addr)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	txOut := wire.NewTxOut(value, pkScript)

	tx := wire.NewNativeMsgTx(wire.TxVersion, []*wire.TxIn{txIn}, []*wire.TxOut{txOut})

	if sigScript == nil {
		sigScript, err = txscript.SignTxOutput(&dagconfig.MainNetParams,
			tx, 0, originTx.TxOut[0].PkScript, txscript.SigHashAll,
			txscript.KeyClosure(lookupKey), nil, nil)
	}
	tx.TxIn[0].SignatureScript = sigScript
	return tx, nil
}

// TestCalcPriority ensures the priority calculations work as intended.
func TestCalcPriority(t *testing.T) {
	// commonSourceTx1 is a valid transaction used in the tests below as an
	// input to transactions that are having their priority calculated.
	//
	// From block 7 in main blockchain.
	// tx 0437cd7f8525ceed2324359c2d0ba26006d92d856a9c20fa0241106ee5a597c9
	commonSourceTx1, err := createTransaction(5000000000, nil, 0, hexToBytes("04ffff001d0134"))

	if err != nil {
		t.Errorf("Error with creating source tx: %v", err)
	}

	// commonRedeemTx1 is a valid transaction used in the tests below as the
	// transaction to calculate the priority for.
	//
	// It originally came from block 170 in main blockchain.
	commonRedeemTx1, err := createTransaction(5000000000, commonSourceTx1, 0, nil)

	if err != nil {
		t.Errorf("Error with creating redeem tx: %v", err)
	}

	tests := []struct {
		name       string           // test description
		tx         *wire.MsgTx      // tx to calc priority for
		utxoSet    blockdag.UTXOSet // inputs to tx
		nextHeight uint64           // height for priority calc
		want       float64          // expected priority
	}{
		{
			name: "one height 7 input, prio tx height 169",
			tx:   commonRedeemTx1,
			utxoSet: newUTXOSet([]*wire.MsgTx{commonSourceTx1},
				[]uint64{7}),
			nextHeight: 169,
			want:       1.125e+10,
		},
		{
			name: "one height 100 input, prio tx height 169",
			tx:   commonRedeemTx1,
			utxoSet: newUTXOSet([]*wire.MsgTx{commonSourceTx1},
				[]uint64{100}),
			nextHeight: 169,
			want:       4.791666666666667e+09,
		},
		{
			name: "one height 7 input, prio tx height 100000",
			tx:   commonRedeemTx1,
			utxoSet: newUTXOSet([]*wire.MsgTx{commonSourceTx1},
				[]uint64{7}),
			nextHeight: 100000,
			want:       6.943958333333333e+12,
		},
		{
			name: "one height 100 input, prio tx height 100000",
			tx:   commonRedeemTx1,
			utxoSet: newUTXOSet([]*wire.MsgTx{commonSourceTx1},
				[]uint64{100}),
			nextHeight: 100000,
			want:       6.9375e+12,
		},
	}

	for i, test := range tests {
		got := CalcPriority(test.tx, test.utxoSet, test.nextHeight)
		if got != test.want {
			t.Errorf("CalcPriority #%d (%q): unexpected priority "+
				"got %v want %v", i, test.name, got, test.want)
			continue
		}
	}
}

var privKeyBytes, _ = hex.DecodeString("22a47fa09a223f2aa079edf85a7c2" +
	"d4f8720ee63e502ee2869afab7de234b80c")

var privKey, pubKey = btcec.PrivKeyFromBytes(btcec.S256(), privKeyBytes)
var pubKeyHash = util.Hash160(pubKey.SerializeCompressed())
var addr, _ = util.NewAddressPubKeyHash(pubKeyHash, util.Bech32PrefixDAGCoin)
