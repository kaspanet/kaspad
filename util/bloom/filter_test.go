// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package bloom_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/bloom"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
)

// TestFilterLarge ensures a maximum sized filter can be created.
func TestFilterLarge(t *testing.T) {
	f := bloom.NewFilter(100000000, 0, 0.01, wire.BloomUpdateNone)
	if len(f.MsgFilterLoad().Filter) > wire.MaxFilterLoadFilterSize {
		t.Errorf("TestFilterLarge test failed: %d > %d",
			len(f.MsgFilterLoad().Filter), wire.MaxFilterLoadFilterSize)
	}
}

// TestFilterLoad ensures loading and unloading of a filter pass.
func TestFilterLoad(t *testing.T) {
	merkle := wire.MsgFilterLoad{}

	f := bloom.LoadFilter(&merkle)
	if !f.IsLoaded() {
		t.Errorf("TestFilterLoad IsLoaded test failed: want %v got %v",
			true, !f.IsLoaded())
		return
	}
	f.Unload()
	if f.IsLoaded() {
		t.Errorf("TestFilterLoad IsLoaded test failed: want %v got %v",
			f.IsLoaded(), false)
		return
	}
}

// TestFilterInsert ensures inserting data into the filter causes that data
// to be matched and the resulting serialized MsgFilterLoad is the expected
// value.
func TestFilterInsert(t *testing.T) {
	var tests = []struct {
		hex    string
		insert bool
	}{
		{"99108ad8ed9bb6274d3980bab5a85c048f0950c8", true},
		{"19108ad8ed9bb6274d3980bab5a85c048f0950c8", false},
		{"b5a2c786d9ef4658287ced5914b37a1b4aa32eee", true},
		{"b9300670b4c5366e95b2699e8b18bc75e5f729c5", true},
	}

	f := bloom.NewFilter(3, 0, 0.01, wire.BloomUpdateAll)

	for i, test := range tests {
		data, err := hex.DecodeString(test.hex)
		if err != nil {
			t.Errorf("TestFilterInsert DecodeString failed: %v\n", err)
			return
		}
		if test.insert {
			f.Add(data)
		}

		result := f.Matches(data)
		if test.insert != result {
			t.Errorf("TestFilterInsert Matches test #%d failure: got %v want %v\n",
				i, result, test.insert)
			return
		}
	}

	want, err := hex.DecodeString("03614e9b050000000000000001")
	if err != nil {
		t.Errorf("TestFilterInsert DecodeString failed: %v\n", err)
		return
	}

	got := bytes.NewBuffer(nil)
	err = f.MsgFilterLoad().KaspaEncode(got, wire.ProtocolVersion)
	if err != nil {
		t.Errorf("TestFilterInsert KaspaDecode failed: %v\n", err)
		return
	}

	if !bytes.Equal(got.Bytes(), want) {
		t.Errorf("TestFilterInsert failure: got %v want %v\n",
			got.Bytes(), want)
		return
	}
}

// TestFilterFPRange checks that new filters made with out of range
// false positive targets result in either max or min false positive rates.
func TestFilterFPRange(t *testing.T) {
	tests := []struct {
		name   string
		hash   string
		want   string
		filter *bloom.Filter
	}{
		{
			name:   "fprates > 1 should be clipped at 1",
			hash:   "02981fa052f0481dbc5868f4fc2166035a10f27a03cfd2de67326471df5bc041",
			want:   "00000000000000000001",
			filter: bloom.NewFilter(1, 0, 20.9999999769, wire.BloomUpdateAll),
		},
		{
			name:   "fprates less than 1e-9 should be clipped at min",
			hash:   "02981fa052f0481dbc5868f4fc2166035a10f27a03cfd2de67326471df5bc041",
			want:   "0566d97a91a91b0000000000000001",
			filter: bloom.NewFilter(1, 0, 0, wire.BloomUpdateAll),
		},
		{
			name:   "negative fprates should be clipped at min",
			hash:   "02981fa052f0481dbc5868f4fc2166035a10f27a03cfd2de67326471df5bc041",
			want:   "0566d97a91a91b0000000000000001",
			filter: bloom.NewFilter(1, 0, -1, wire.BloomUpdateAll),
		},
	}

	for _, test := range tests {
		// Convert test input to appropriate types.
		hash, err := daghash.NewHashFromStr(test.hash)
		if err != nil {
			t.Errorf("NewHashFromStr unexpected error: %v", err)
			continue
		}
		want, err := hex.DecodeString(test.want)
		if err != nil {
			t.Errorf("DecodeString unexpected error: %v\n", err)
			continue
		}

		// Add the test hash to the bloom filter and ensure the
		// filter serializes to the expected bytes.
		f := test.filter
		f.AddHash(hash)
		got := bytes.NewBuffer(nil)
		err = f.MsgFilterLoad().KaspaEncode(got, wire.ProtocolVersion)
		if err != nil {
			t.Errorf("KaspaDecode unexpected error: %v\n", err)
			continue
		}
		if !bytes.Equal(got.Bytes(), want) {
			t.Errorf("serialized filter mismatch: got %x want %x\n",
				got.Bytes(), want)
			continue
		}
	}
}

// TestFilterInsert ensures inserting data into the filter with a tweak causes
// that data to be matched and the resulting serialized MsgFilterLoad is the
// expected value.
func TestFilterInsertWithTweak(t *testing.T) {
	var tests = []struct {
		hex    string
		insert bool
	}{
		{"99108ad8ed9bb6274d3980bab5a85c048f0950c8", true},
		{"19108ad8ed9bb6274d3980bab5a85c048f0950c8", false},
		{"b5a2c786d9ef4658287ced5914b37a1b4aa32eee", true},
		{"b9300670b4c5366e95b2699e8b18bc75e5f729c5", true},
	}

	f := bloom.NewFilter(3, 2147483649, 0.01, wire.BloomUpdateAll)

	for i, test := range tests {
		data, err := hex.DecodeString(test.hex)
		if err != nil {
			t.Errorf("TestFilterInsertWithTweak DecodeString failed: %v\n", err)
			return
		}
		if test.insert {
			f.Add(data)
		}

		result := f.Matches(data)
		if test.insert != result {
			t.Errorf("TestFilterInsertWithTweak Matches test #%d failure: got %v want %v\n",
				i, result, test.insert)
			return
		}
	}

	want, err := hex.DecodeString("03ce4299050000000100008001")
	if err != nil {
		t.Errorf("TestFilterInsertWithTweak DecodeString failed: %v\n", err)
		return
	}
	got := bytes.NewBuffer(nil)
	err = f.MsgFilterLoad().KaspaEncode(got, wire.ProtocolVersion)
	if err != nil {
		t.Errorf("TestFilterInsertWithTweak KaspaDecode failed: %v\n", err)
		return
	}

	if !bytes.Equal(got.Bytes(), want) {
		t.Errorf("TestFilterInsertWithTweak failure: got %v want %v\n",
			got.Bytes(), want)
		return
	}
}

// TestFilterInsertKey ensures inserting public keys and addresses works as
// expected.
func TestFilterInsertKey(t *testing.T) {
	secret := "5Kg1gnAjaLfKiwhhPpGS3QfRg2m6awQvaj98JCZBZQ5SuS2F15C"

	wif, err := util.DecodeWIF(secret)
	if err != nil {
		t.Errorf("TestFilterInsertKey DecodeWIF failed: %v", err)
		return
	}

	f := bloom.NewFilter(2, 0, 0.001, wire.BloomUpdateAll)
	serializedPubKey, err := wif.SerializePubKey()
	if err != nil {
		t.Errorf("TestFilterInsertKey SerializePubKey failed: %v", err)
		return
	}
	f.Add(serializedPubKey)
	f.Add(util.Hash160(serializedPubKey))

	want, err := hex.DecodeString("038fc16b080000000000000001")
	if err != nil {
		t.Errorf("TestFilterInsertWithTweak DecodeString failed: %v\n", err)
		return
	}
	got := bytes.NewBuffer(nil)
	err = f.MsgFilterLoad().KaspaEncode(got, wire.ProtocolVersion)
	if err != nil {
		t.Errorf("TestFilterInsertWithTweak KaspaDecode failed: %v\n", err)
		return
	}

	if !bytes.Equal(got.Bytes(), want) {
		t.Errorf("TestFilterInsertWithTweak failure: got %v want %v\n",
			got.Bytes(), want)
		return
	}
}

func TestFilterBloomMatch(t *testing.T) {
	strBytes := []byte{
		0x01, 0x00, 0x00, 0x00, 0x01, 0x0b, 0x26, 0xe9,
		0xb7, 0x73, 0x5e, 0xb6, 0xaa, 0xbd, 0xf3, 0x58,
		0xba, 0xb6, 0x2f, 0x98, 0x16, 0xa2, 0x1b, 0xa9,
		0xeb, 0xdb, 0x71, 0x9d, 0x52, 0x99, 0xe8, 0x86,
		0x07, 0xd7, 0x22, 0xc1, 0x90, 0x00, 0x00, 0x00,
		0x00, 0x8b, 0x48, 0x30, 0x45, 0x02, 0x20, 0x07,
		0x0a, 0xca, 0x44, 0x50, 0x6c, 0x5c, 0xef, 0x3a,
		0x16, 0xed, 0x51, 0x9d, 0x7c, 0x3c, 0x39, 0xf8,
		0xaa, 0xb1, 0x92, 0xc4, 0xe1, 0xc9, 0x0d, 0x06,
		0x5f, 0x37, 0xb8, 0xa4, 0xaf, 0x61, 0x41, 0x02,
		0x21, 0x00, 0xa8, 0xe1, 0x60, 0xb8, 0x56, 0xc2,
		0xd4, 0x3d, 0x27, 0xd8, 0xfb, 0xa7, 0x1e, 0x5a,
		0xef, 0x64, 0x05, 0xb8, 0x64, 0x3a, 0xc4, 0xcb,
		0x7c, 0xb3, 0xc4, 0x62, 0xac, 0xed, 0x7f, 0x14,
		0x71, 0x1a, 0x01, 0x41, 0x04, 0x6d, 0x11, 0xfe,
		0xe5, 0x1b, 0x0e, 0x60, 0x66, 0x6d, 0x50, 0x49,
		0xa9, 0x10, 0x1a, 0x72, 0x74, 0x1d, 0xf4, 0x80,
		0xb9, 0x6e, 0xe2, 0x64, 0x88, 0xa4, 0xd3, 0x46,
		0x6b, 0x95, 0xc9, 0xa4, 0x0a, 0xc5, 0xee, 0xef,
		0x87, 0xe1, 0x0a, 0x5c, 0xd3, 0x36, 0xc1, 0x9a,
		0x84, 0x56, 0x5f, 0x80, 0xfa, 0x6c, 0x54, 0x79,
		0x57, 0xb7, 0x70, 0x0f, 0xf4, 0xdf, 0xbd, 0xef,
		0xe7, 0x60, 0x36, 0xc3, 0x39, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0x02, 0x1b, 0xff,
		0x3d, 0x11, 0x00, 0x00, 0x00, 0x00, 0x19, 0x76,
		0xa9, 0x14, 0x04, 0x94, 0x3f, 0xdd, 0x50, 0x80,
		0x53, 0xc7, 0x50, 0x00, 0x10, 0x6d, 0x3b, 0xc6,
		0xe2, 0x75, 0x4d, 0xbc, 0xff, 0x19, 0x88, 0xac,
		0x2f, 0x15, 0xde, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x19, 0x76, 0xa9, 0x14, 0xa2, 0x66, 0x43, 0x6d,
		0x29, 0x65, 0x54, 0x76, 0x08, 0xb9, 0xe1, 0x5d,
		0x90, 0x32, 0xa7, 0xb9, 0xd6, 0x4f, 0xa4, 0x31,
		0x88, 0xac, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	tx, err := util.NewTxFromBytes(strBytes)
	if err != nil {
		t.Errorf("TestFilterBloomMatch NewTxFromBytes failure: %v", err)
		return
	}
	spendingTxBytes := []byte{
		0x01, 0x00, 0x00, 0x00, 0x01, 0x95, 0x7c, 0x1d,
		0xfd, 0x07, 0x87, 0xc3, 0x2b, 0xb7, 0x67, 0xbb,
		0xa9, 0x4d, 0x29, 0x0e, 0x64, 0xdc, 0x3d, 0x12,
		0x19, 0xbf, 0x53, 0xe6, 0x15, 0x01, 0xef, 0xb3,
		0xfc, 0x5d, 0xc0, 0xf9, 0x81, 0x00, 0x00, 0x00,
		0x00, 0x8c, 0x49, 0x30, 0x46, 0x02, 0x21, 0x00,
		0xda, 0x0d, 0xc6, 0xae, 0xce, 0xfe, 0x1e, 0x06,
		0xef, 0xdf, 0x05, 0x77, 0x37, 0x57, 0xde, 0xb1,
		0x68, 0x82, 0x09, 0x30, 0xe3, 0xb0, 0xd0, 0x3f,
		0x46, 0xf5, 0xfc, 0xf1, 0x50, 0xbf, 0x99, 0x0c,
		0x02, 0x21, 0x00, 0xd2, 0x5b, 0x5c, 0x87, 0x04,
		0x00, 0x76, 0xe4, 0xf2, 0x53, 0xf8, 0x26, 0x2e,
		0x76, 0x3e, 0x2d, 0xd5, 0x1e, 0x7f, 0xf0, 0xbe,
		0x15, 0x77, 0x27, 0xc4, 0xbc, 0x42, 0x80, 0x7f,
		0x17, 0xbd, 0x39, 0x01, 0x41, 0x04, 0xe6, 0xc2,
		0x6e, 0xf6, 0x7d, 0xc6, 0x10, 0xd2, 0xcd, 0x19,
		0x24, 0x84, 0x78, 0x9a, 0x6c, 0xf9, 0xae, 0xa9,
		0x93, 0x0b, 0x94, 0x4b, 0x7e, 0x2d, 0xb5, 0x34,
		0x2b, 0x9d, 0x9e, 0x5b, 0x9f, 0xf7, 0x9a, 0xff,
		0x9a, 0x2e, 0xe1, 0x97, 0x8d, 0xd7, 0xfd, 0x01,
		0xdf, 0xc5, 0x22, 0xee, 0x02, 0x28, 0x3d, 0x3b,
		0x06, 0xa9, 0xd0, 0x3a, 0xcf, 0x80, 0x96, 0x96,
		0x8d, 0x7d, 0xbb, 0x0f, 0x91, 0x78, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x02, 0x8b,
		0xa7, 0x94, 0x0e, 0x00, 0x00, 0x00, 0x00, 0x19,
		0x76, 0xa9, 0x14, 0xba, 0xde, 0xec, 0xfd, 0xef,
		0x05, 0x07, 0x24, 0x7f, 0xc8, 0xf7, 0x42, 0x41,
		0xd7, 0x3b, 0xc0, 0x39, 0x97, 0x2d, 0x7b, 0x88,
		0xac, 0x40, 0x94, 0xa8, 0x02, 0x00, 0x00, 0x00,
		0x00, 0x19, 0x76, 0xa9, 0x14, 0xc1, 0x09, 0x32,
		0x48, 0x3f, 0xec, 0x93, 0xed, 0x51, 0xf5, 0xfe,
		0x95, 0xe7, 0x25, 0x59, 0xf2, 0xcc, 0x70, 0x43,
		0xf9, 0x88, 0xac, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}

	spendingTx, err := util.NewTxFromBytes(spendingTxBytes)
	if err != nil {
		t.Errorf("TestFilterBloomMatch NewTxFromBytes failure: %v", err)
		return
	}

	f := bloom.NewFilter(10, 0, 0.000001, wire.BloomUpdateAll)
	inputStr := "81f9c05dfcb3ef0115e653bf19123ddc640e294da9bb67b72bc38707fd1d7c95" // byte-reversed tx id
	hash, err := daghash.NewHashFromStr(inputStr)
	if err != nil {
		t.Errorf("TestFilterBloomMatch NewHashFromStr failed: %v\n", err)
		return
	}
	f.AddHash(hash)
	if !f.MatchTxAndUpdate(tx) {
		t.Errorf("TestFilterBloomMatch didn't match ID, want %s, got %s", inputStr, tx.ID())
	}

	f = bloom.NewFilter(10, 0, 0.000001, wire.BloomUpdateAll)
	inputStr = "957c1dfd0787c32bb767bba94d290e64dc3d1219bf53e61501efb3fc5dc0f981" // non-reversed tx id
	hashBytes, err := hex.DecodeString(inputStr)
	if err != nil {
		t.Errorf("TestFilterBloomMatch DecodeString failed: %v\n", err)
		return
	}
	f.Add(hashBytes)

	if !f.MatchTxAndUpdate(tx) {
		t.Errorf("TestFilterBloomMatch didn't match ID, want %s, got %s", inputStr,
			hex.EncodeToString(tx.ID()[:]))
	}

	f = bloom.NewFilter(10, 0, 0.000001, wire.BloomUpdateAll)
	inputStr = "30450220070aca44506c5cef3a16ed519d7c3c39f8aab192c4e1c90d065" +
		"f37b8a4af6141022100a8e160b856c2d43d27d8fba71e5aef6405b8643" +
		"ac4cb7cb3c462aced7f14711a01"
	hashBytes, err = hex.DecodeString(inputStr)
	if err != nil {
		t.Errorf("TestFilterBloomMatch DecodeString failed: %v\n", err)
		return
	}
	f.Add(hashBytes)
	if !f.MatchTxAndUpdate(tx) {
		t.Errorf("TestFilterBloomMatch didn't match input signature %s", inputStr)
	}

	f = bloom.NewFilter(10, 0, 0.000001, wire.BloomUpdateAll)
	inputStr = "046d11fee51b0e60666d5049a9101a72741df480b96ee26488a4d3466b95" +
		"c9a40ac5eeef87e10a5cd336c19a84565f80fa6c547957b7700ff4dfbdefe" +
		"76036c339"
	hashBytes, err = hex.DecodeString(inputStr)
	if err != nil {
		t.Errorf("TestFilterBloomMatch DecodeString failed: %v\n", err)
		return
	}
	f.Add(hashBytes)
	if !f.MatchTxAndUpdate(tx) {
		t.Errorf("TestFilterBloomMatch didn't match input pubkey %s", inputStr)
	}

	f = bloom.NewFilter(10, 0, 0.000001, wire.BloomUpdateAll)
	inputStr = "04943fdd508053c75000106d3bc6e2754dbcff19"
	hashBytes, err = hex.DecodeString(inputStr)
	if err != nil {
		t.Errorf("TestFilterBloomMatch DecodeString failed: %v\n", err)
		return
	}
	f.Add(hashBytes)
	if !f.MatchTxAndUpdate(tx) {
		t.Errorf("TestFilterBloomMatch didn't match output address %s", inputStr)
	}
	if !f.MatchTxAndUpdate(spendingTx) {
		t.Errorf("TestFilterBloomMatch spendingTx didn't match output address %s", inputStr)
	}

	f = bloom.NewFilter(10, 0, 0.000001, wire.BloomUpdateAll)
	inputStr = "a266436d2965547608b9e15d9032a7b9d64fa431"
	hashBytes, err = hex.DecodeString(inputStr)
	if err != nil {
		t.Errorf("TestFilterBloomMatch DecodeString failed: %v\n", err)
		return
	}
	f.Add(hashBytes)
	if !f.MatchTxAndUpdate(tx) {
		t.Errorf("TestFilterBloomMatch didn't match output address %s", inputStr)
	}

	f = bloom.NewFilter(10, 0, 0.000001, wire.BloomUpdateAll)
	inputStr = "90c122d70786e899529d71dbeba91ba216982fb6ba58f3bdaab65e73b7e9260b"
	txID, err := daghash.NewTxIDFromStr(inputStr)
	if err != nil {
		t.Errorf("TestFilterBloomMatch NewHashFromStr failed: %v\n", err)
		return
	}
	outpoint := wire.NewOutpoint(txID, 0)
	f.AddOutpoint(outpoint)
	if !f.MatchTxAndUpdate(tx) {
		t.Errorf("TestFilterBloomMatch didn't match outpoint %s", inputStr)
	}

	f = bloom.NewFilter(10, 0, 0.000001, wire.BloomUpdateAll)
	inputStr = "00000009e784f32f62ef849763d4f45b98e07ba658647343b915ff832b110436"
	hash, err = daghash.NewHashFromStr(inputStr)
	if err != nil {
		t.Errorf("TestFilterBloomMatch NewHashFromStr failed: %v\n", err)
		return
	}
	f.AddHash(hash)
	if f.MatchTxAndUpdate(tx) {
		t.Errorf("TestFilterBloomMatch matched hash %s", inputStr)
	}

	f = bloom.NewFilter(10, 0, 0.000001, wire.BloomUpdateAll)
	inputStr = "0000006d2965547608b9e15d9032a7b9d64fa431"
	hashBytes, err = hex.DecodeString(inputStr)
	if err != nil {
		t.Errorf("TestFilterBloomMatch DecodeString failed: %v\n", err)
		return
	}
	f.Add(hashBytes)
	if f.MatchTxAndUpdate(tx) {
		t.Errorf("TestFilterBloomMatch matched address %s", inputStr)
	}

	f = bloom.NewFilter(10, 0, 0.000001, wire.BloomUpdateAll)
	inputStr = "90c122d70786e899529d71dbeba91ba216982fb6ba58f3bdaab65e73b7e9260b"
	txID, err = daghash.NewTxIDFromStr(inputStr)
	if err != nil {
		t.Errorf("TestFilterBloomMatch NewHashFromStr failed: %v\n", err)
		return
	}
	outpoint = wire.NewOutpoint(txID, 1)
	f.AddOutpoint(outpoint)
	if f.MatchTxAndUpdate(tx) {
		t.Errorf("TestFilterBloomMatch matched outpoint %s", inputStr)
	}

	f = bloom.NewFilter(10, 0, 0.000001, wire.BloomUpdateAll)
	inputStr = "000000d70786e899529d71dbeba91ba216982fb6ba58f3bdaab65e73b7e9260b"
	txID, err = daghash.NewTxIDFromStr(inputStr)
	if err != nil {
		t.Errorf("TestFilterBloomMatch NewHashFromStr failed: %v\n", err)
		return
	}
	outpoint = wire.NewOutpoint(txID, 0)
	f.AddOutpoint(outpoint)
	if f.MatchTxAndUpdate(tx) {
		t.Errorf("TestFilterBloomMatch matched outpoint %s", inputStr)
	}
}

func TestFilterInsertUpdateNone(t *testing.T) {
	f := bloom.NewFilter(10, 0, 0.000001, wire.BloomUpdateNone)

	// Add the generation pubkey
	inputStr := "04eaafc2314def4ca98ac970241bcab022b9c1e1f4ea423a20f134c" +
		"876f2c01ec0f0dd5b2e86e7168cefe0d81113c3807420ce13ad1357231a" +
		"2252247d97a46a91"
	inputBytes, err := hex.DecodeString(inputStr)
	if err != nil {
		t.Errorf("TestFilterInsertUpdateNone DecodeString failed: %v", err)
		return
	}
	f.Add(inputBytes)

	// Add the output address for the 4th transaction
	inputStr = "b6efd80d99179f4f4ff6f4dd0a007d018c385d21"
	inputBytes, err = hex.DecodeString(inputStr)
	if err != nil {
		t.Errorf("TestFilterInsertUpdateNone DecodeString failed: %v", err)
		return
	}
	f.Add(inputBytes)

	inputStr = "147caa76786596590baa4e98f5d9f48b86c7765e489f7a6ff3360fe5c674360b"
	txID, err := daghash.NewTxIDFromStr(inputStr)
	if err != nil {
		t.Errorf("TestFilterInsertUpdateNone NewHashFromStr failed: %v", err)
		return
	}
	outpoint := wire.NewOutpoint(txID, 0)

	if f.MatchesOutpoint(outpoint) {
		t.Errorf("TestFilterInsertUpdateNone matched outpoint %s", inputStr)
		return
	}

	inputStr = "02981fa052f0481dbc5868f4fc2166035a10f27a03cfd2de67326471df5bc041"
	txID, err = daghash.NewTxIDFromStr(inputStr)
	if err != nil {
		t.Errorf("TestFilterInsertUpdateNone NewHashFromStr failed: %v", err)
		return
	}
	outpoint = wire.NewOutpoint(txID, 0)

	if f.MatchesOutpoint(outpoint) {
		t.Errorf("TestFilterInsertUpdateNone matched outpoint %s", inputStr)
		return
	}
}

func TestFilterReload(t *testing.T) {
	f := bloom.NewFilter(10, 0, 0.000001, wire.BloomUpdateAll)

	bFilter := bloom.LoadFilter(f.MsgFilterLoad())
	if bFilter.MsgFilterLoad() == nil {
		t.Errorf("TestFilterReload LoadFilter test failed")
		return
	}
	bFilter.Reload(nil)

	if bFilter.MsgFilterLoad() != nil {
		t.Errorf("TestFilterReload Reload test failed")
	}
}
