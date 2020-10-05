package blockdag

import (
	"bytes"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/util/daghash"
	"testing"
)

func Benchmark_serializeUTXO(b *testing.B) {
	entry := &UTXOEntry{
		amount:         5000000000,
		scriptPubKey:   hexToBytes("76a914ad06dd6ddee55cbca9a9e3713bd7587509a3056488ac"), // p2pkh
		blockBlueScore: 1432432,
		packedFlags:    0,
	}
	outpoint := &appmessage.Outpoint{
		TxID: daghash.TxID{
			0x16, 0x5e, 0x38, 0xe8, 0xb3, 0x91, 0x45, 0x95,
			0xd9, 0xc6, 0x41, 0xf3, 0xb8, 0xee, 0xc2, 0xf3,
			0x46, 0x11, 0x89, 0x6b, 0x82, 0x1a, 0x68, 0x3b,
			0x7a, 0x4e, 0xde, 0xfe, 0x2c, 0x00, 0x00, 0x00,
		},
		Index: 0xffffffff,
	}

	buf := bytes.NewBuffer(make([]byte, 8+1+8+9+len(entry.scriptPubKey)+len(outpoint.TxID)+4))

	for i := 0; i < b.N; i++ {
		buf.Reset()
		err := serializeUTXO(buf, entry, outpoint)
		if err != nil {
			b.Fatal(err)
		}
	}
}
