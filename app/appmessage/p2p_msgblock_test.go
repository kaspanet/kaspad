// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package appmessage

import (
	"bytes"
	"io"
	"math"
	"reflect"
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"

	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/pkg/errors"

	"github.com/davecgh/go-spew/spew"
)

// TestBlock tests the MsgBlock API.
func TestBlock(t *testing.T) {
	pver := ProtocolVersion

	// Block 1 header.
	parentHashes := blockOne.Header.ParentHashes
	hashMerkleRoot := blockOne.Header.HashMerkleRoot
	acceptedIDMerkleRoot := blockOne.Header.AcceptedIDMerkleRoot
	utxoCommitment := blockOne.Header.UTXOCommitment
	bits := blockOne.Header.Bits
	nonce := blockOne.Header.Nonce
	bh := NewBlockHeader(1, parentHashes, hashMerkleRoot, acceptedIDMerkleRoot, utxoCommitment, bits, nonce)

	// Ensure the command is expected value.
	wantCmd := MessageCommand(5)
	msg := NewMsgBlock(bh)
	if cmd := msg.Command(); cmd != wantCmd {
		t.Errorf("NewMsgBlock: wrong command - got %v want %v",
			cmd, wantCmd)
	}

	// Ensure max payload is expected value for latest protocol version.
	wantPayload := uint32(1024 * 1024 * 32)
	maxPayload := msg.MaxPayloadLength(pver)
	if maxPayload != wantPayload {
		t.Errorf("MaxPayloadLength: wrong max payload length for "+
			"protocol version %d - got %v, want %v", pver,
			maxPayload, wantPayload)
	}

	// Ensure we get the same block header data back out.
	if !reflect.DeepEqual(&msg.Header, bh) {
		t.Errorf("NewMsgBlock: wrong block header - got %v, want %v",
			spew.Sdump(&msg.Header), spew.Sdump(bh))
	}

	// Ensure transactions are added properly.
	tx := blockOne.Transactions[0].Copy()
	msg.AddTransaction(tx)
	if !reflect.DeepEqual(msg.Transactions, blockOne.Transactions) {
		t.Errorf("AddTransaction: wrong transactions - got %v, want %v",
			spew.Sdump(msg.Transactions),
			spew.Sdump(blockOne.Transactions))
	}

	// Ensure transactions are properly cleared.
	msg.ClearTransactions()
	if len(msg.Transactions) != 0 {
		t.Errorf("ClearTransactions: wrong transactions - got %v, want %v",
			len(msg.Transactions), 0)
	}
}

// TestBlockHash tests the ability to generate the hash of a block accurately.
func TestBlockHash(t *testing.T) {
	// Block 1 hash.
	hashStr := "5150f62f89507cc64dd7c9efbd536996a96efc77f759b513e978af82665448a6"
	wantHash, err := hashes.FromString(hashStr)
	if err != nil {
		t.Errorf("NewHashFromStr: %v", err)
	}

	// Ensure the hash produced is expected.
	blockHash := blockOne.BlockHash()
	if *blockHash != *wantHash {
		t.Errorf("BlockHash: wrong hash - got %v, want %v",
			spew.Sprint(blockHash), spew.Sprint(wantHash))
	}
}

func TestConvertToPartial(t *testing.T) {
	localSubnetworkID := &externalapi.DomainSubnetworkID{0x12}

	transactions := []struct {
		subnetworkID          *externalapi.DomainSubnetworkID
		payload               []byte
		expectedPayloadLength int
	}{
		{
			subnetworkID:          &subnetworks.SubnetworkIDNative,
			payload:               []byte{},
			expectedPayloadLength: 0,
		},
		{
			subnetworkID:          &subnetworks.SubnetworkIDRegistry,
			payload:               []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
			expectedPayloadLength: 0,
		},
		{
			subnetworkID:          localSubnetworkID,
			payload:               []byte{0x01},
			expectedPayloadLength: 1,
		},
		{
			subnetworkID:          &externalapi.DomainSubnetworkID{0x34},
			payload:               []byte{0x02},
			expectedPayloadLength: 0,
		},
	}

	block := MsgBlock{}
	payload := []byte{1}
	for _, transaction := range transactions {
		block.Transactions = append(block.Transactions, NewSubnetworkMsgTx(1, nil, nil, transaction.subnetworkID, 0, payload))
	}

	block.ConvertToPartial(localSubnetworkID)

	for _, testTransaction := range transactions {
		var subnetworkTx *MsgTx
		for _, blockTransaction := range block.Transactions {
			if blockTransaction.SubnetworkID == *testTransaction.subnetworkID {
				subnetworkTx = blockTransaction
			}
		}
		if subnetworkTx == nil {
			t.Errorf("ConvertToPartial: subnetworkID '%s' not found in block!", testTransaction.subnetworkID)
			continue
		}

		payloadLength := len(subnetworkTx.Payload)
		if payloadLength != testTransaction.expectedPayloadLength {
			t.Errorf("ConvertToPartial: unexpected payload length for subnetwork '%s': expected: %d, got: %d",
				testTransaction.subnetworkID, testTransaction.expectedPayloadLength, payloadLength)
		}
	}
}

// TestBlockEncoding tests the MsgBlock appmessage encode and decode for various numbers
// of transaction inputs and outputs and protocol versions.
func TestBlockEncoding(t *testing.T) {
	tests := []struct {
		in     *MsgBlock // Message to encode
		out    *MsgBlock // Expected decoded message
		buf    []byte    // Encoded value
		txLocs []TxLoc   // Expected transaction locations
		pver   uint32    // Protocol version for appmessage encoding
	}{
		// Latest protocol version.
		{
			&blockOne,
			&blockOne,
			blockOneBytes,
			blockOneTxLocs,
			ProtocolVersion,
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Encode the message to appmessage format.
		var buf bytes.Buffer
		err := test.in.KaspaEncode(&buf, test.pver)
		if err != nil {
			t.Errorf("KaspaEncode #%d error %v", i, err)
			continue
		}
		if !bytes.Equal(buf.Bytes(), test.buf) {
			t.Errorf("KaspaEncode #%d\n got: %s want: %s", i,
				spew.Sdump(buf.Bytes()), spew.Sdump(test.buf))
			continue
		}

		// Decode the message from appmessage format.
		var msg MsgBlock
		rbuf := bytes.NewReader(test.buf)
		err = msg.KaspaDecode(rbuf, test.pver)
		if err != nil {
			t.Errorf("KaspaDecode #%d error %v", i, err)
			continue
		}
		if !reflect.DeepEqual(&msg, test.out) {
			t.Errorf("KaspaDecode #%d\n got: %s want: %s", i,
				spew.Sdump(&msg), spew.Sdump(test.out))
			continue
		}
	}
}

// TestBlockEncodingErrors performs negative tests against appmessage encode and decode
// of MsgBlock to confirm error paths work correctly.
func TestBlockEncodingErrors(t *testing.T) {
	pver := ProtocolVersion

	tests := []struct {
		in       *MsgBlock // Value to encode
		buf      []byte    // Encoded value
		pver     uint32    // Protocol version for appmessage encoding
		max      int       // Max size of fixed buffer to induce errors
		writeErr error     // Expected write error
		readErr  error     // Expected read error
	}{
		// Force error in version.
		{&blockOne, blockOneBytes, pver, 0, io.ErrShortWrite, io.EOF},
		// Force error in num block hashes.
		{&blockOne, blockOneBytes, pver, 4, io.ErrShortWrite, io.EOF},
		// Force error in prev block hash #1.
		{&blockOne, blockOneBytes, pver, 5, io.ErrShortWrite, io.EOF},
		// Force error in prev block hash #2.
		{&blockOne, blockOneBytes, pver, 37, io.ErrShortWrite, io.EOF},
		// Force error in hash merkle root.
		{&blockOne, blockOneBytes, pver, 69, io.ErrShortWrite, io.EOF},
		// Force error in accepted ID merkle root.
		{&blockOne, blockOneBytes, pver, 101, io.ErrShortWrite, io.EOF},
		// Force error in utxo commitment.
		{&blockOne, blockOneBytes, pver, 133, io.ErrShortWrite, io.EOF},
		// Force error in timestamp.
		{&blockOne, blockOneBytes, pver, 165, io.ErrShortWrite, io.EOF},
		// Force error in difficulty bits.
		{&blockOne, blockOneBytes, pver, 173, io.ErrShortWrite, io.EOF},
		// Force error in header nonce.
		{&blockOne, blockOneBytes, pver, 177, io.ErrShortWrite, io.EOF},
		// Force error in transaction count.
		{&blockOne, blockOneBytes, pver, 185, io.ErrShortWrite, io.EOF},
		// Force error in transactions.
		{&blockOne, blockOneBytes, pver, 186, io.ErrShortWrite, io.EOF},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Encode to appmessage format.
		w := newFixedWriter(test.max)
		err := test.in.KaspaEncode(w, test.pver)
		if !errors.Is(err, test.writeErr) {
			t.Errorf("KaspaEncode #%d wrong error got: %v, want: %v",
				i, err, test.writeErr)
			continue
		}

		// Decode from appmessage format.
		var msg MsgBlock
		r := newFixedReader(test.max, test.buf)
		err = msg.KaspaDecode(r, test.pver)
		if !errors.Is(err, test.readErr) {
			t.Errorf("KaspaDecode #%d wrong error got: %v, want: %v",
				i, err, test.readErr)
			continue
		}
	}
}

// TestBlockSerialize tests MsgBlock serialize and deserialize.
func TestBlockSerialize(t *testing.T) {
	tests := []struct {
		in     *MsgBlock // Message to encode
		out    *MsgBlock // Expected decoded message
		buf    []byte    // Serialized data
		txLocs []TxLoc   // Expected transaction locations
	}{
		{
			&blockOne,
			&blockOne,
			blockOneBytes,
			blockOneTxLocs,
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Serialize the block.
		var buf bytes.Buffer
		err := test.in.Serialize(&buf)
		if err != nil {
			t.Errorf("Serialize #%d error %v", i, err)
			continue
		}
		if !bytes.Equal(buf.Bytes(), test.buf) {
			t.Errorf("Serialize #%d\n got: %s want: %s", i,
				spew.Sdump(buf.Bytes()), spew.Sdump(test.buf))
			continue
		}

		// Deserialize the block.
		var block MsgBlock
		rbuf := bytes.NewReader(test.buf)
		err = block.Deserialize(rbuf)
		if err != nil {
			t.Errorf("Deserialize #%d error %v", i, err)
			continue
		}
		if !reflect.DeepEqual(&block, test.out) {
			t.Errorf("Deserialize #%d\n got: %s want: %s", i,
				spew.Sdump(&block), spew.Sdump(test.out))
			continue
		}

		// Deserialize the block while gathering transaction location
		// information.
		var txLocBlock MsgBlock
		br := bytes.NewBuffer(test.buf)
		txLocs, err := txLocBlock.DeserializeTxLoc(br)
		if err != nil {
			t.Errorf("DeserializeTxLoc #%d error %v", i, err)
			continue
		}
		if !reflect.DeepEqual(&txLocBlock, test.out) {
			t.Errorf("DeserializeTxLoc #%d\n got: %s want: %s", i,
				spew.Sdump(&txLocBlock), spew.Sdump(test.out))
			continue
		}
		if !reflect.DeepEqual(txLocs, test.txLocs) {
			t.Errorf("DeserializeTxLoc #%d\n got: %s want: %s", i,
				spew.Sdump(txLocs), spew.Sdump(test.txLocs))
			continue
		}
	}
}

// TestBlockSerializeErrors performs negative tests against appmessage encode and
// decode of MsgBlock to confirm error paths work correctly.
func TestBlockSerializeErrors(t *testing.T) {
	tests := []struct {
		in       *MsgBlock // Value to encode
		buf      []byte    // Serialized data
		max      int       // Max size of fixed buffer to induce errors
		writeErr error     // Expected write error
		readErr  error     // Expected read error
	}{
		// Force error in version.
		{&blockOne, blockOneBytes, 0, io.ErrShortWrite, io.EOF},
		// Force error in numParentBlocks.
		{&blockOne, blockOneBytes, 4, io.ErrShortWrite, io.EOF},
		// Force error in prev block hash #1.
		{&blockOne, blockOneBytes, 5, io.ErrShortWrite, io.EOF},
		// Force error in prev block hash #2.
		{&blockOne, blockOneBytes, 37, io.ErrShortWrite, io.EOF},
		// Force error in hash merkle root.
		{&blockOne, blockOneBytes, 69, io.ErrShortWrite, io.EOF},
		// Force error in accepted ID merkle root.
		{&blockOne, blockOneBytes, 101, io.ErrShortWrite, io.EOF},
		// Force error in utxo commitment.
		{&blockOne, blockOneBytes, 133, io.ErrShortWrite, io.EOF},
		// Force error in timestamp.
		{&blockOne, blockOneBytes, 165, io.ErrShortWrite, io.EOF},
		// Force error in difficulty bits.
		{&blockOne, blockOneBytes, 173, io.ErrShortWrite, io.EOF},
		// Force error in header nonce.
		{&blockOne, blockOneBytes, 177, io.ErrShortWrite, io.EOF},
		// Force error in transaction count.
		{&blockOne, blockOneBytes, 185, io.ErrShortWrite, io.EOF},
		// Force error in transactions.
		{&blockOne, blockOneBytes, 186, io.ErrShortWrite, io.EOF},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Serialize the block.
		w := newFixedWriter(test.max)
		err := test.in.Serialize(w)
		if !errors.Is(err, test.writeErr) {
			t.Errorf("Serialize #%d wrong error got: %v, want: %v",
				i, err, test.writeErr)
			continue
		}

		// Deserialize the block.
		var block MsgBlock
		r := newFixedReader(test.max, test.buf)
		err = block.Deserialize(r)
		if !errors.Is(err, test.readErr) {
			t.Errorf("Deserialize #%d wrong error got: %v, want: %v",
				i, err, test.readErr)
			continue
		}

		var txLocBlock MsgBlock
		br := bytes.NewBuffer(test.buf[0:test.max])
		_, err = txLocBlock.DeserializeTxLoc(br)
		if !errors.Is(err, test.readErr) {
			t.Errorf("DeserializeTxLoc #%d wrong error got: %v, want: %v",
				i, err, test.readErr)
			continue
		}
	}
}

// TestBlockOverflowErrors  performs tests to ensure deserializing blocks which
// are intentionally crafted to use large values for the number of transactions
// are handled properly. This could otherwise potentially be used as an attack
// vector.
func TestBlockOverflowErrors(t *testing.T) {
	pver := ProtocolVersion

	tests := []struct {
		buf  []byte // Encoded value
		pver uint32 // Protocol version for appmessage encoding
		err  error  // Expected error
	}{
		// Block that claims to have ~uint64(0) transactions.
		{
			[]byte{
				0x01, 0x00, 0x00, 0x00, // Version 1
				0x02,                                           // NumParentBlocks
				0x6f, 0xe2, 0x8c, 0x0a, 0xb6, 0xf1, 0xb3, 0x72, // mainnetGenesisHash
				0xc1, 0xa6, 0xa2, 0x46, 0xae, 0x63, 0xf7, 0x4f,
				0x93, 0x1e, 0x83, 0x65, 0xe1, 0x5a, 0x08, 0x9c,
				0x68, 0xd6, 0x19, 0x00, 0x00, 0x00, 0x00, 0x00,
				0xf6, 0x7a, 0xd7, 0x69, 0x5d, 0x9b, 0x66, 0x2a, // simnetGenesisHash
				0x72, 0xff, 0x3d, 0x8e, 0xdb, 0xbb, 0x2d, 0xe0,
				0xbf, 0xa6, 0x7b, 0x13, 0x97, 0x4b, 0xb9, 0x91,
				0x0d, 0x11, 0x6d, 0x5c, 0xbd, 0x86, 0x3e, 0x68,
				0x3b, 0xa3, 0xed, 0xfd, 0x7a, 0x7b, 0x12, 0xb2, // HashMerkleRoot
				0x7a, 0xc7, 0x2c, 0x3e, 0x67, 0x76, 0x8f, 0x61,
				0x7f, 0xc8, 0x1b, 0xc3, 0x88, 0x8a, 0x51, 0x32,
				0x3a, 0x9f, 0xb8, 0xaa, 0x4b, 0x1e, 0x5e, 0x4a,
				0x09, 0x3B, 0xC7, 0xE3, 0x67, 0x11, 0x7B, 0x3C, // AcceptedIDMerkleRoot
				0x30, 0xC1, 0xF8, 0xFD, 0xD0, 0xD9, 0x72, 0x87,
				0x7F, 0x16, 0xC5, 0x96, 0x2E, 0x8B, 0xD9, 0x63,
				0x65, 0x9C, 0x79, 0x3C, 0xE3, 0x70, 0xD9, 0x5F,
				0x10, 0x3B, 0xC7, 0xE3, 0x67, 0x11, 0x7B, 0x3C, // UTXOCommitment
				0x30, 0xC1, 0xF8, 0xFD, 0xD0, 0xD9, 0x72, 0x87,
				0x7F, 0x16, 0xC5, 0x96, 0x2E, 0x8B, 0xD9, 0x63,
				0x65, 0x9C, 0x79, 0x3C, 0xE3, 0x70, 0xD9, 0x5F,
				0x61, 0xbc, 0x66, 0x49, 0x00, 0x00, 0x00, 0x00, // Timestamp
				0xff, 0xff, 0x00, 0x1d, // Bits
				0x01, 0xe3, 0x62, 0x99, 0x00, 0x00, 0x00, 0x00, // Fake Nonce
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, // TxnCount
			}, pver, &MessageError{},
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Decode from appmessage format.
		var msg MsgBlock
		r := bytes.NewReader(test.buf)
		err := msg.KaspaDecode(r, test.pver)
		if reflect.TypeOf(err) != reflect.TypeOf(test.err) {
			t.Errorf("KaspaDecode #%d wrong error got: %v, want: %v",
				i, err, reflect.TypeOf(test.err))
			continue
		}

		// Deserialize from appmessage format.
		r = bytes.NewReader(test.buf)
		err = msg.Deserialize(r)
		if reflect.TypeOf(err) != reflect.TypeOf(test.err) {
			t.Errorf("Deserialize #%d wrong error got: %v, want: %v",
				i, err, reflect.TypeOf(test.err))
			continue
		}

		// Deserialize with transaction location info from appmessage format.
		br := bytes.NewBuffer(test.buf)
		_, err = msg.DeserializeTxLoc(br)
		if reflect.TypeOf(err) != reflect.TypeOf(test.err) {
			t.Errorf("DeserializeTxLoc #%d wrong error got: %v, "+
				"want: %v", i, err, reflect.TypeOf(test.err))
			continue
		}
	}
}

// TestBlockSerializeSize performs tests to ensure the serialize size for
// various blocks is accurate.
func TestBlockSerializeSize(t *testing.T) {
	// Block with no transactions.
	noTxBlock := NewMsgBlock(&blockOne.Header)

	tests := []struct {
		in   *MsgBlock // Block to encode
		size int       // Expected serialized size
	}{
		// Block with no transactions.
		{noTxBlock, 186},

		// First block in the mainnet block DAG.
		{&blockOne, len(blockOneBytes)},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		serializedSize := test.in.SerializeSize()
		if serializedSize != test.size {
			t.Errorf("MsgBlock.SerializeSize: #%d got: %d, want: "+
				"%d", i, serializedSize, test.size)

			continue
		}
	}
}

// blockOne is the first block in the mainnet block DAG.
var blockOne = MsgBlock{
	Header: BlockHeader{
		Version:              1,
		ParentHashes:         []*externalapi.DomainHash{mainnetGenesisHash, simnetGenesisHash},
		HashMerkleRoot:       mainnetGenesisMerkleRoot,
		AcceptedIDMerkleRoot: exampleAcceptedIDMerkleRoot,
		UTXOCommitment:       exampleUTXOCommitment,
		Timestamp:            mstime.UnixMilliseconds(0x17315ed0f99),
		Bits:                 0x1d00ffff, // 486604799
		Nonce:                0x9962e301, // 2573394689
	},
	Transactions: []*MsgTx{
		NewNativeMsgTx(1,
			[]*TxIn{
				{
					PreviousOutpoint: Outpoint{
						TxID:  externalapi.DomainTransactionID{},
						Index: 0xffffffff,
					},
					SignatureScript: []byte{
						0x04, 0xff, 0xff, 0x00, 0x1d, 0x01, 0x04,
					},
					Sequence: math.MaxUint64,
				},
			},
			[]*TxOut{
				{
					Value: 0x12a05f200,
					ScriptPubKey: []byte{
						0x41, // OP_DATA_65
						0x04, 0x96, 0xb5, 0x38, 0xe8, 0x53, 0x51, 0x9c,
						0x72, 0x6a, 0x2c, 0x91, 0xe6, 0x1e, 0xc1, 0x16,
						0x00, 0xae, 0x13, 0x90, 0x81, 0x3a, 0x62, 0x7c,
						0x66, 0xfb, 0x8b, 0xe7, 0x94, 0x7b, 0xe6, 0x3c,
						0x52, 0xda, 0x75, 0x89, 0x37, 0x95, 0x15, 0xd4,
						0xe0, 0xa6, 0x04, 0xf8, 0x14, 0x17, 0x81, 0xe6,
						0x22, 0x94, 0x72, 0x11, 0x66, 0xbf, 0x62, 0x1e,
						0x73, 0xa8, 0x2c, 0xbf, 0x23, 0x42, 0xc8, 0x58,
						0xee, // 65-byte signature
						0xac, // OP_CHECKSIG
					},
				},
			}),
	},
}

// Block one serialized bytes.
var blockOneBytes = []byte{
	0x01, 0x00, 0x00, 0x00, // Version 1
	0x02,                                           // NumParentBlocks
	0xdc, 0x5f, 0x5b, 0x5b, 0x1d, 0xc2, 0xa7, 0x25, // mainnetGenesisHash
	0x49, 0xd5, 0x1d, 0x4d, 0xee, 0xd7, 0xa4, 0x8b,
	0xaf, 0xd3, 0x14, 0x4b, 0x56, 0x78, 0x98, 0xb1,
	0x8c, 0xfd, 0x9f, 0x69, 0xdd, 0xcf, 0xbb, 0x63,
	0xf6, 0x7a, 0xd7, 0x69, 0x5d, 0x9b, 0x66, 0x2a, // simnetGenesisHash
	0x72, 0xff, 0x3d, 0x8e, 0xdb, 0xbb, 0x2d, 0xe0,
	0xbf, 0xa6, 0x7b, 0x13, 0x97, 0x4b, 0xb9, 0x91,
	0x0d, 0x11, 0x6d, 0x5c, 0xbd, 0x86, 0x3e, 0x68,
	0x4a, 0x5e, 0x1e, 0x4b, 0xaa, 0xb8, 0x9f, 0x3a, // HashMerkleRoot
	0x32, 0x51, 0x8a, 0x88, 0xc3, 0x1b, 0xc8, 0x7f,
	0x61, 0x8f, 0x76, 0x67, 0x3e, 0x2c, 0xc7, 0x7a,
	0xb2, 0x12, 0x7b, 0x7a, 0xfd, 0xed, 0xa3, 0x3b,
	0x09, 0x3B, 0xC7, 0xE3, 0x67, 0x11, 0x7B, 0x3C, // AcceptedIDMerkleRoot
	0x30, 0xC1, 0xF8, 0xFD, 0xD0, 0xD9, 0x72, 0x87,
	0x7F, 0x16, 0xC5, 0x96, 0x2E, 0x8B, 0xD9, 0x63,
	0x65, 0x9C, 0x79, 0x3C, 0xE3, 0x70, 0xD9, 0x5F,
	0x10, 0x3B, 0xC7, 0xE3, 0x67, 0x11, 0x7B, 0x3C, // UTXOCommitment
	0x30, 0xC1, 0xF8, 0xFD, 0xD0, 0xD9, 0x72, 0x87,
	0x7F, 0x16, 0xC5, 0x96, 0x2E, 0x8B, 0xD9, 0x63,
	0x65, 0x9C, 0x79, 0x3C, 0xE3, 0x70, 0xD9, 0x5F,
	0x99, 0x0f, 0xed, 0x15, 0x73, 0x01, 0x00, 0x00, // Timestamp
	0xff, 0xff, 0x00, 0x1d, // Bits
	0x01, 0xe3, 0x62, 0x99, 0x00, 0x00, 0x00, 0x00, // Fake Nonce
	0x01,                   // TxnCount
	0x01, 0x00, 0x00, 0x00, // Version
	0x01, // Varint for number of transaction inputs
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Previous output hash
	0xff, 0xff, 0xff, 0xff, // Prevous output index
	0x07,                                     // Varint for length of signature script
	0x04, 0xff, 0xff, 0x00, 0x1d, 0x01, 0x04, // Signature script (coinbase)
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // Sequence
	0x01,                                           // Varint for number of transaction outputs
	0x00, 0xf2, 0x05, 0x2a, 0x01, 0x00, 0x00, 0x00, // Transaction amount
	0x43, // Varint for length of scriptPubKey
	0x41, // OP_DATA_65
	0x04, 0x96, 0xb5, 0x38, 0xe8, 0x53, 0x51, 0x9c,
	0x72, 0x6a, 0x2c, 0x91, 0xe6, 0x1e, 0xc1, 0x16,
	0x00, 0xae, 0x13, 0x90, 0x81, 0x3a, 0x62, 0x7c,
	0x66, 0xfb, 0x8b, 0xe7, 0x94, 0x7b, 0xe6, 0x3c,
	0x52, 0xda, 0x75, 0x89, 0x37, 0x95, 0x15, 0xd4,
	0xe0, 0xa6, 0x04, 0xf8, 0x14, 0x17, 0x81, 0xe6,
	0x22, 0x94, 0x72, 0x11, 0x66, 0xbf, 0x62, 0x1e,
	0x73, 0xa8, 0x2c, 0xbf, 0x23, 0x42, 0xc8, 0x58,
	0xee,                                           // 65-byte uncompressed public key
	0xac,                                           // OP_CHECKSIG
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Lock time
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, // SubnetworkID
}

// Transaction location information for block one transactions.
var blockOneTxLocs = []TxLoc{
	{TxStart: 186, TxLen: 162},
}
