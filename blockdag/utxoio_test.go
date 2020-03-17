package blockdag

import (
	"bytes"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
	"reflect"
	"testing"
)

func TestSerializeUTXO(t *testing.T) {
	txID0, _ := daghash.NewTxIDFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	outpoint0 := wire.NewOutpoint(txID0, 55555555)
	utxoEntry0 := NewUTXOEntry(&wire.TxOut{ScriptPubKey: []byte{0x00}, Value: 10}, true, 0)

	w := &bytes.Buffer{}
	err := serializeUTXO(w, utxoEntry0, outpoint0)
	if err != nil {
		t.Fatalf("serializeUTXO: %s", err)
	}

	r := bytes.NewReader(w.Bytes())
	dentry, dout, err := deserializeUTXO(r)
	if err != nil {
		t.Fatalf("deserializeUTXO: %s", err)
	}

	if *dout != *outpoint0 {
		t.Fatalf("ddddd")
	}

	if !reflect.DeepEqual(dentry, utxoEntry0) {
		t.Fatalf("aaaaaaaaa")

	}
}
