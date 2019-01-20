// This is a small tool to generate testdata blocks file

package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/wire"
)

func main() {
	targetFile, numBlocks := parseArgs()

	out, err := os.Create(targetFile)
	if err != nil {
		panic(fmt.Errorf("error reading target file: %s", err))
	}
	defer func() {
		err := out.Close()
		if err != nil {
			panic(fmt.Errorf("error closing target file: %s", err))
		}
	}()

	generateBlocks(out, numBlocks)
}

func generateBlocks(out *os.File, numBlocks int) {
	lastBlock := dagconfig.MainNetParams.GenesisBlock

	for i := 0; i < numBlocks; i++ {
		lastBlock = generateBlock(lastBlock)
		writeBlock(out, lastBlock)
	}
}

func generateBlock(parent *wire.MsgBlock) *wire.MsgBlock {
	return &wire.MsgBlock{
		Header: wire.BlockHeader{
			Version:      1,
			ParentHashes: []daghash.Hash{parent.BlockHash()},
			MerkleRoot:   genesisMerkleRoot,
			Timestamp:    time.Unix(0x5b28c4c8, 0), // 2018-06-19 08:54:32 +0000 UTC
			Bits:         0x2e00ffff,               // 503382015 [000000ffff000000000000000000000000000000000000000000000000000000]
			Nonce:        0xc0192550,               // 2148484547
		},
		Transactions: []*wire.MsgTx{&genesisCoinbaseTx},
	}
}

func writeBlock(out *os.File, block *wire.MsgBlock) {
	writeNet(out)

	blockLen := uint32(block.SerializeSize())
	buf := bytes.NewBuffer(make([]byte, 0, blockLen))

	err := block.Serialize(buf)
	if err != nil {
		panic(fmt.Errorf("error serializing block: %s", err))
	}

	err = binary.Write(out, binary.LittleEndian, blockLen)
	if err != nil {
		panic(fmt.Errorf("error writing blockLen: %s", err))
	}

	_, err = out.Write(buf.Bytes())
	if err != nil {
		panic(fmt.Errorf("error writing block: %s", err))
	}
}

func writeNet(out *os.File) {
	err := binary.Write(out, binary.LittleEndian, wire.MainNet)
	if err != nil {
		panic(fmt.Errorf("error writing net to file: %s", err))
	}
}

func parseArgs() (targetFile string, numBlocks int) {
	if len(os.Args) != 3 {
		printUsage()
	}

	targetFile = os.Args[1]
	numBlocks, err := strconv.Atoi(os.Args[2])
	if err != nil {
		printUsage()
	}

	return
}

func printUsage() {
	fmt.Println("Usage: generator [targetFile] [numBlocks]")
	os.Exit(1)
}

var genesisCoinbaseTx = wire.MsgTx{
	Version: 1,
	TxIn: []*wire.TxIn{
		{
			PreviousOutPoint: wire.OutPoint{
				Hash:  daghash.Hash{},
				Index: 0xffffffff,
			},
			SignatureScript: []byte{
				0x04, 0xff, 0xff, 0x00, 0x1d, 0x01, 0x04, 0x45, /* |.......E| */
				0x54, 0x68, 0x65, 0x20, 0x54, 0x69, 0x6d, 0x65, /* |The Time| */
				0x73, 0x20, 0x30, 0x33, 0x2f, 0x4a, 0x61, 0x6e, /* |s 03/Jan| */
				0x2f, 0x32, 0x30, 0x30, 0x39, 0x20, 0x43, 0x68, /* |/2009 Ch| */
				0x61, 0x6e, 0x63, 0x65, 0x6c, 0x6c, 0x6f, 0x72, /* |ancellor| */
				0x20, 0x6f, 0x6e, 0x20, 0x62, 0x72, 0x69, 0x6e, /* | on brin| */
				0x6b, 0x20, 0x6f, 0x66, 0x20, 0x73, 0x65, 0x63, /* |k of sec|*/
				0x6f, 0x6e, 0x64, 0x20, 0x62, 0x61, 0x69, 0x6c, /* |ond bail| */
				0x6f, 0x75, 0x74, 0x20, 0x66, 0x6f, 0x72, 0x20, /* |out for |*/
				0x62, 0x61, 0x6e, 0x6b, 0x73, /* |banks| */
			},
			Sequence: 0xffffffff,
		},
	},
	TxOut: []*wire.TxOut{
		{
			Value: 0x12a05f200,
			PkScript: []byte{
				0x41, 0x04, 0x67, 0x8a, 0xfd, 0xb0, 0xfe, 0x55, /* |A.g....U| */
				0x48, 0x27, 0x19, 0x67, 0xf1, 0xa6, 0x71, 0x30, /* |H'.g..q0| */
				0xb7, 0x10, 0x5c, 0xd6, 0xa8, 0x28, 0xe0, 0x39, /* |..\..(.9| */
				0x09, 0xa6, 0x79, 0x62, 0xe0, 0xea, 0x1f, 0x61, /* |..yb...a| */
				0xde, 0xb6, 0x49, 0xf6, 0xbc, 0x3f, 0x4c, 0xef, /* |..I..?L.| */
				0x38, 0xc4, 0xf3, 0x55, 0x04, 0xe5, 0x1e, 0xc1, /* |8..U....| */
				0x12, 0xde, 0x5c, 0x38, 0x4d, 0xf7, 0xba, 0x0b, /* |..\8M...| */
				0x8d, 0x57, 0x8a, 0x4c, 0x70, 0x2b, 0x6b, 0xf1, /* |.W.Lp+k.| */
				0x1d, 0x5f, 0xac, /* |._.| */
			},
		},
	},
	LockTime:     0,
	SubnetworkID: wire.SubnetworkIDNative,
}

var genesisMerkleRoot = daghash.Hash([daghash.HashSize]byte{ // Make go vet happy.
	0x3b, 0xa3, 0xed, 0xfd, 0x7a, 0x7b, 0x12, 0xb2,
	0x7a, 0xc7, 0x2c, 0x3e, 0x67, 0x76, 0x8f, 0x61,
	0x7f, 0xc8, 0x1b, 0xc3, 0x88, 0x8a, 0x51, 0x32,
	0x3a, 0x9f, 0xb8, 0xaa, 0x4b, 0x1e, 0x5e, 0x4a,
})
