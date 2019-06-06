// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package bloom_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/util/bloom"
	"github.com/daglabs/btcd/util/daghash"
	"github.com/daglabs/btcd/wire"
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
	err = f.MsgFilterLoad().BtcEncode(got, wire.ProtocolVersion)
	if err != nil {
		t.Errorf("TestFilterInsert BtcDecode failed: %v\n", err)
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
		err = f.MsgFilterLoad().BtcEncode(got, wire.ProtocolVersion)
		if err != nil {
			t.Errorf("BtcDecode unexpected error: %v\n", err)
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
	err = f.MsgFilterLoad().BtcEncode(got, wire.ProtocolVersion)
	if err != nil {
		t.Errorf("TestFilterInsertWithTweak BtcDecode failed: %v\n", err)
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
	f.Add(wif.SerializePubKey())
	f.Add(util.Hash160(wif.SerializePubKey()))

	want, err := hex.DecodeString("038fc16b080000000000000001")
	if err != nil {
		t.Errorf("TestFilterInsertWithTweak DecodeString failed: %v\n", err)
		return
	}
	got := bytes.NewBuffer(nil)
	err = f.MsgFilterLoad().BtcEncode(got, wire.ProtocolVersion)
	if err != nil {
		t.Errorf("TestFilterInsertWithTweak BtcDecode failed: %v\n", err)
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

func TestFilterInsertP2PubKeyOnly(t *testing.T) {
	blockBytes := []byte{
		0x01, 0x00, 0x00, 0x00, // Version
		0x01,                                                             // NumParentBlocks
		0x82, 0xBB, 0x86, 0x9C, 0xF3, 0xA7, 0x93, 0x43, 0x2A, 0x66, 0xE8, // ParentHashes
		0x26, 0xE0, 0x5A, 0x6F, 0xC3, 0x74, 0x69, 0xF8, 0xEF, 0xB7, 0x42,
		0x1D, 0xC8, 0x80, 0x67, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x7F, 0x16, 0xC5, 0x96, 0x2E, 0x8B, 0xD9, 0x63, 0x65, 0x9C, 0x79, // HashMerkleRoot
		0x3C, 0xE3, 0x70, 0xD9, 0x5F, 0x09, 0x3B, 0xC7, 0xE3, 0x67, 0x11,
		0x7B, 0x3C, 0x30, 0xC1, 0xF8, 0xFD, 0xD0, 0xD9, 0x72, 0x87,
		0x3C, 0xE3, 0x70, 0xD9, 0x5F, 0x09, 0x3B, 0xC7, 0xE3, 0x67, 0x11, // AcceptedIDMerkleRoot
		0x7F, 0x16, 0xC5, 0x96, 0x2E, 0x8B, 0xD9, 0x63, 0x65, 0x9C, 0x79,
		0x7B, 0x3C, 0x30, 0xC1, 0xF8, 0xFD, 0xD0, 0xD9, 0x72, 0x87,
		0x10, 0x3B, 0xC7, 0xE3, 0x67, 0x11, 0x7B, 0x3C, // UTXOCommitment
		0x30, 0xC1, 0xF8, 0xFD, 0xD0, 0xD9, 0x72, 0x87,
		0x7F, 0x16, 0xC5, 0x96, 0x2E, 0x8B, 0xD9, 0x63,
		0x65, 0x9C, 0x79, 0x3C, 0xE3, 0x70, 0xD9, 0x5F,
		0x76, 0x38, 0x1B, 0x4D, 0x00, 0x00, 0x00, 0x00, // Time
		0x4C, 0x86, 0x04, 0x1B, // Bits
		0x55, 0x4B, 0x85, 0x29, 0x00, 0x00, 0x00, 0x00, // Fake Nonce. TODO: (Ori) Replace to a real nonce
		0x07, // NumTxns
		0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Txs[0]
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0xFF, 0xFF, 0x07, 0x04, 0x4C,
		0x86, 0x04, 0x1B, 0x01, 0x36, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0x01, 0x00, 0xF2, 0x05, 0x2A, 0x01, 0x00, 0x00, 0x00,
		0x43, 0x41, 0x04, 0xEA, 0xAF, 0xC2, 0x31, 0x4D, 0xEF, 0x4C, 0xA9,
		0x8A, 0xC9, 0x70, 0x24, 0x1B, 0xCA, 0xB0, 0x22, 0xB9, 0xC1, 0xE1,
		0xF4, 0xEA, 0x42, 0x3A, 0x20, 0xF1, 0x34, 0xC8, 0x76, 0xF2, 0xC0,
		0x1E, 0xC0, 0xF0, 0xDD, 0x5B, 0x2E, 0x86, 0xE7, 0x16, 0x8C, 0xEF,
		0xE0, 0xD8, 0x11, 0x13, 0xC3, 0x80, 0x74, 0x20, 0xCE, 0x13, 0xAD,
		0x13, 0x57, 0x23, 0x1A, 0x22, 0x52, 0x24, 0x7D, 0x97, 0xA4, 0x6A,
		0x91, 0xAC, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00,

		0x01, 0x00, 0x00, 0x00, 0x01, 0xBC, 0xAD, 0x20, 0xA6, 0xA2, 0x98, // Txs[1]
		0x27, 0xD1, 0x42, 0x4F, 0x08, 0x98, 0x92, 0x55, 0x12, 0x0B, 0xF7,
		0xF3, 0xE9, 0xE3, 0xCD, 0xAA, 0xA6, 0xBB, 0x31, 0xB0, 0x73, 0x7F,
		0xE0, 0x48, 0x72, 0x43, 0x00, 0x00, 0x00, 0x00, 0x49, 0x48, 0x30,
		0x45, 0x02, 0x20, 0x35, 0x6E, 0x83, 0x4B, 0x04, 0x6C, 0xAD, 0xC0,
		0xF8, 0xEB, 0xB5, 0xA8, 0xA0, 0x17, 0xB0, 0x2D, 0xE5, 0x9C, 0x86,
		0x30, 0x54, 0x03, 0xDA, 0xD5, 0x2C, 0xD7, 0x7B, 0x55, 0xAF, 0x06,
		0x2E, 0xA1, 0x02, 0x21, 0x00, 0x92, 0x53, 0xCD, 0x6C, 0x11, 0x9D,
		0x47, 0x29, 0xB7, 0x7C, 0x97, 0x8E, 0x1E, 0x2A, 0xA1, 0x9F, 0x5E,
		0xA6, 0xE0, 0xE5, 0x2B, 0x3F, 0x16, 0xE3, 0x2F, 0xA6, 0x08, 0xCD,
		0x5B, 0xAB, 0x75, 0x39, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0x02, 0x00, 0x8D, 0x38, 0x0C, 0x01, 0x00, 0x00, 0x00,
		0x19, 0x76, 0xA9, 0x14, 0x2B, 0x4B, 0x80, 0x72, 0xEC, 0xBB, 0xA1,
		0x29, 0xB6, 0x45, 0x3C, 0x63, 0xE1, 0x29, 0xE6, 0x43, 0x20, 0x72,
		0x49, 0xCA, 0x88, 0xAC, 0x00, 0x65, 0xCD, 0x1D, 0x00, 0x00, 0x00,
		0x00, 0x19, 0x76, 0xA9, 0x14, 0x1B, 0x8D, 0xD1, 0x3B, 0x99, 0x4B,
		0xCF, 0xC7, 0x87, 0xB3, 0x2A, 0xEA, 0xDF, 0x58, 0xCC, 0xB3, 0x61,
		0x5C, 0xBD, 0x54, 0x88, 0xAC, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x01, 0x00, 0x00, 0x00, 0x03, 0xFD, 0xAC, 0xF9, 0xB3, 0xEB, 0x07, // Txs[2]
		0x74, 0x12, 0xE7, 0xA9, 0x68, 0xD2, 0xE4, 0xF1, 0x1B, 0x9A, 0x9D,
		0xEE, 0x31, 0x2D, 0x66, 0x61, 0x87, 0xED, 0x77, 0xEE, 0x7D, 0x26,
		0xAF, 0x16, 0xCB, 0x0B, 0x00, 0x00, 0x00, 0x00, 0x8C, 0x49, 0x30,
		0x46, 0x02, 0x21, 0x00, 0xEA, 0x16, 0x08, 0xE7, 0x09, 0x11, 0xCA,
		0x0D, 0xE5, 0xAF, 0x51, 0xBA, 0x57, 0xAD, 0x23, 0xB9, 0xA5, 0x1D,
		0xB8, 0xD2, 0x8F, 0x82, 0xC5, 0x35, 0x63, 0xC5, 0x6A, 0x05, 0xC2,
		0x0F, 0x5A, 0x87, 0x02, 0x21, 0x00, 0xA8, 0xBD, 0xC8, 0xB4, 0xA8,
		0xAC, 0xC8, 0x63, 0x4C, 0x6B, 0x42, 0x04, 0x10, 0x15, 0x07, 0x75,
		0xEB, 0x7F, 0x24, 0x74, 0xF5, 0x61, 0x5F, 0x7F, 0xCC, 0xD6, 0x5A,
		0xF3, 0x0F, 0x31, 0x0F, 0xBF, 0x01, 0x41, 0x04, 0x65, 0xFD, 0xF4,
		0x9E, 0x29, 0xB0, 0x6B, 0x9A, 0x15, 0x82, 0x28, 0x7B, 0x62, 0x79,
		0x01, 0x4F, 0x83, 0x4E, 0xDC, 0x31, 0x76, 0x95, 0xD1, 0x25, 0xEF,
		0x62, 0x3C, 0x1C, 0xC3, 0xAA, 0xEC, 0xE2, 0x45, 0xBD, 0x69, 0xFC,
		0xAD, 0x75, 0x08, 0x66, 0x6E, 0x9C, 0x74, 0xA4, 0x9D, 0xC9, 0x05,
		0x6D, 0x5F, 0xC1, 0x43, 0x38, 0xEF, 0x38, 0x11, 0x8D, 0xC4, 0xAF,
		0xAE, 0x5F, 0xE2, 0xC5, 0x85, 0xCA, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0x30, 0x9E, 0x19, 0x13, 0x63, 0x4E, 0xCB, 0x50,
		0xF3, 0xC4, 0xF8, 0x3E, 0x96, 0xE7, 0x0B, 0x2D, 0xF0, 0x71, 0xB4,
		0x97, 0xB8, 0x97, 0x3A, 0x3E, 0x75, 0x42, 0x9D, 0xF3, 0x97, 0xB5,
		0xAF, 0x83, 0x00, 0x00, 0x00, 0x00, 0x49, 0x48, 0x30, 0x45, 0x02,
		0x20, 0x2B, 0xDB, 0x79, 0xC5, 0x96, 0xA9, 0xFF, 0xC2, 0x4E, 0x96,
		0xF4, 0x38, 0x61, 0x99, 0xAB, 0xA3, 0x86, 0xE9, 0xBC, 0x7B, 0x60,
		0x71, 0x51, 0x6E, 0x2B, 0x51, 0xDD, 0xA9, 0x42, 0xB3, 0xA1, 0xED,
		0x02, 0x21, 0x00, 0xC5, 0x3A, 0x85, 0x7E, 0x76, 0xB7, 0x24, 0xFC,
		0x14, 0xD4, 0x53, 0x11, 0xEA, 0xC5, 0x01, 0x96, 0x50, 0xD4, 0x15,
		0xC3, 0xAB, 0xB5, 0x42, 0x8F, 0x3A, 0xAE, 0x16, 0xD8, 0xE6, 0x9B,
		0xEC, 0x23, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0x20, 0x89, 0xE3, 0x34, 0x91, 0x69, 0x50, 0x80, 0xC9, 0xED, 0xC1,
		0x8A, 0x42, 0x8F, 0x7D, 0x83, 0x4D, 0xB5, 0xB6, 0xD3, 0x72, 0xDF,
		0x13, 0xCE, 0x2B, 0x1B, 0x0E, 0x0C, 0xBC, 0xB1, 0xE6, 0xC1, 0x00,
		0x00, 0x00, 0x00, 0x49, 0x48, 0x30, 0x45, 0x02, 0x21, 0x00, 0xD4,
		0xCE, 0x67, 0xC5, 0x89, 0x6E, 0xE2, 0x51, 0xC8, 0x10, 0xAC, 0x1F,
		0xF9, 0xCE, 0xCC, 0xD3, 0x28, 0xB4, 0x97, 0xC8, 0xF5, 0x53, 0xAB,
		0x6E, 0x08, 0x43, 0x1E, 0x7D, 0x40, 0xBA, 0xD6, 0xB5, 0x02, 0x20,
		0x33, 0x11, 0x9C, 0x0C, 0x2B, 0x7D, 0x79, 0x2D, 0x31, 0xF1, 0x18,
		0x77, 0x79, 0xC7, 0xBD, 0x95, 0xAE, 0xFD, 0x93, 0xD9, 0x0A, 0x71,
		0x55, 0x86, 0xD7, 0x38, 0x01, 0xD9, 0xB4, 0x74, 0x71, 0xC6, 0x01,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x01, 0x00, 0x71,
		0x44, 0x60, 0x03, 0x00, 0x00, 0x00, 0x19, 0x76, 0xA9, 0x14, 0xC7,
		0xB5, 0x51, 0x41, 0xD0, 0x97, 0xEA, 0x5D, 0xF7, 0xA0, 0xED, 0x33,
		0x0C, 0xF7, 0x94, 0x37, 0x6E, 0x53, 0xEC, 0x8D, 0x88, 0xAC, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00,
		0x01, 0x00, 0x00, 0x00, 0x04, 0x5B, 0xF0, 0xE2, 0x14, 0xAA, 0x40, // Txs[3]
		0x69, 0xA3, 0xE7, 0x92, 0xEC, 0xEE, 0x1E, 0x1B, 0xF0, 0xC1, 0xD3,
		0x97, 0xCD, 0xE8, 0xDD, 0x08, 0x13, 0x8F, 0x4B, 0x72, 0xA0, 0x06,
		0x81, 0x74, 0x34, 0x47, 0x00, 0x00, 0x00, 0x00, 0x8B, 0x48, 0x30,
		0x45, 0x02, 0x20, 0x0C, 0x45, 0xDE, 0x8C, 0x4F, 0x3E, 0x2C, 0x18,
		0x21, 0xF2, 0xFC, 0x87, 0x8C, 0xBA, 0x97, 0xB1, 0xE6, 0xF8, 0x80,
		0x7D, 0x94, 0x93, 0x07, 0x13, 0xAA, 0x1C, 0x86, 0xA6, 0x7B, 0x9B,
		0xF1, 0xE4, 0x02, 0x21, 0x00, 0x85, 0x81, 0xAB, 0xFE, 0xF2, 0xE3,
		0x0F, 0x95, 0x78, 0x15, 0xFC, 0x89, 0x97, 0x84, 0x23, 0x74, 0x6B,
		0x20, 0x86, 0x37, 0x5C, 0xA8, 0xEC, 0xF3, 0x59, 0xC8, 0x5C, 0x2A,
		0x5B, 0x7C, 0x88, 0xAD, 0x01, 0x41, 0x04, 0x62, 0xBB, 0x73, 0xF7,
		0x6C, 0xA0, 0x99, 0x4F, 0xCB, 0x8B, 0x42, 0x71, 0xE6, 0xFB, 0x75,
		0x61, 0xF5, 0xC0, 0xF9, 0xCA, 0x0C, 0xF6, 0x48, 0x52, 0x61, 0xC4,
		0xA0, 0xDC, 0x89, 0x4F, 0x4A, 0xB8, 0x44, 0xC6, 0xCD, 0xFB, 0x97,
		0xCD, 0x0B, 0x60, 0xFF, 0xB5, 0x01, 0x8F, 0xFD, 0x62, 0x38, 0xF4,
		0xD8, 0x72, 0x70, 0xEF, 0xB1, 0xD3, 0xAE, 0x37, 0x07, 0x9B, 0x79,
		0x4A, 0x92, 0xD7, 0xEC, 0x95, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xD6, 0x69, 0xF7, 0xD7, 0x95, 0x8D, 0x40, 0xFC, 0x59,
		0xD2, 0x25, 0x3D, 0x88, 0xE0, 0xF2, 0x48, 0xE2, 0x9B, 0x59, 0x9C,
		0x80, 0xBB, 0xCE, 0xC3, 0x44, 0xA8, 0x3D, 0xDA, 0x5F, 0x9A, 0xA7,
		0x2C, 0x00, 0x00, 0x00, 0x00, 0x8A, 0x47, 0x30, 0x44, 0x02, 0x20,
		0x78, 0x12, 0x4C, 0x8B, 0xEE, 0xAA, 0x82, 0x5F, 0x9E, 0x0B, 0x30,
		0xBF, 0xF9, 0x6E, 0x56, 0x4D, 0xD8, 0x59, 0x43, 0x2F, 0x2D, 0x0C,
		0xB3, 0xB7, 0x2D, 0x3D, 0x5D, 0x93, 0xD3, 0x8D, 0x7E, 0x93, 0x02,
		0x20, 0x69, 0x1D, 0x23, 0x3B, 0x6C, 0x0F, 0x99, 0x5B, 0xE5, 0xAC,
		0xB0, 0x3D, 0x70, 0xA7, 0xF7, 0xA6, 0x5B, 0x6B, 0xC9, 0xBD, 0xD4,
		0x26, 0x26, 0x0F, 0x38, 0xA1, 0x34, 0x66, 0x69, 0x50, 0x7A, 0x36,
		0x01, 0x41, 0x04, 0x62, 0xBB, 0x73, 0xF7, 0x6C, 0xA0, 0x99, 0x4F,
		0xCB, 0x8B, 0x42, 0x71, 0xE6, 0xFB, 0x75, 0x61, 0xF5, 0xC0, 0xF9,
		0xCA, 0x0C, 0xF6, 0x48, 0x52, 0x61, 0xC4, 0xA0, 0xDC, 0x89, 0x4F,
		0x4A, 0xB8, 0x44, 0xC6, 0xCD, 0xFB, 0x97, 0xCD, 0x0B, 0x60, 0xFF,
		0xB5, 0x01, 0x8F, 0xFD, 0x62, 0x38, 0xF4, 0xD8, 0x72, 0x70, 0xEF,
		0xB1, 0xD3, 0xAE, 0x37, 0x07, 0x9B, 0x79, 0x4A, 0x92, 0xD7, 0xEC,
		0x95, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xF8, 0x78,
		0xAF, 0x0D, 0x93, 0xF5, 0x22, 0x9A, 0x68, 0x16, 0x6C, 0xF0, 0x51,
		0xFD, 0x37, 0x2B, 0xB7, 0xA5, 0x37, 0x23, 0x29, 0x46, 0xE0, 0xA4,
		0x6F, 0x53, 0x63, 0x6B, 0x4D, 0xAF, 0xDA, 0xA4, 0x00, 0x00, 0x00,
		0x00, 0x8C, 0x49, 0x30, 0x46, 0x02, 0x21, 0x00,
		0xC7, 0x17, 0xD1, 0x71, 0x45, 0x51, 0x66, 0x3F, 0x69, 0xC3, 0xC5,
		0x75, 0x9B, 0xDB, 0xB3, 0xA0, 0xFC, 0xD3, 0xFA, 0xB0, 0x23, 0xAB,
		0xC0, 0xE5, 0x22, 0xFE, 0x64, 0x40, 0xDE, 0x35, 0xD8, 0x29, 0x02,
		0x21, 0x00, 0x8D, 0x9C, 0xBE, 0x25, 0xBF, 0xFC, 0x44, 0xAF, 0x2B,
		0x18, 0xE8, 0x1C, 0x58, 0xEB, 0x37, 0x29, 0x3F, 0xD7, 0xFE, 0x1C,
		0x2E, 0x7B, 0x46, 0xFC, 0x37, 0xEE, 0x8C, 0x96, 0xC5, 0x0A, 0xB1,
		0xE2, 0x01, 0x41, 0x04, 0x62, 0xBB, 0x73, 0xF7, 0x6C, 0xA0, 0x99,
		0x4F, 0xCB, 0x8B, 0x42, 0x71, 0xE6, 0xFB, 0x75, 0x61, 0xF5, 0xC0,
		0xF9, 0xCA, 0x0C, 0xF6, 0x48, 0x52, 0x61, 0xC4, 0xA0, 0xDC, 0x89,
		0x4F, 0x4A, 0xB8, 0x44, 0xC6, 0xCD, 0xFB, 0x97, 0xCD, 0x0B, 0x60,
		0xFF, 0xB5, 0x01, 0x8F, 0xFD, 0x62, 0x38, 0xF4, 0xD8, 0x72, 0x70,
		0xEF, 0xB1, 0xD3, 0xAE, 0x37, 0x07, 0x9B, 0x79, 0x4A, 0x92, 0xD7,
		0xEC, 0x95, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x27,
		0xF2, 0xB6, 0x68, 0x85, 0x9C, 0xD7, 0xF2, 0xF8, 0x94, 0xAA, 0x0F,
		0xD2, 0xD9, 0xE6, 0x09, 0x63, 0xBC, 0xD0, 0x7C, 0x88, 0x97, 0x3F,
		0x42, 0x5F, 0x99, 0x9B, 0x8C, 0xBF, 0xD7, 0xA1, 0xE2, 0x00, 0x00,
		0x00, 0x00, 0x8C, 0x49, 0x30, 0x46, 0x02, 0x21, 0x00, 0xE0, 0x08,
		0x47, 0x14, 0x7C, 0xBF, 0x51, 0x7B, 0xCC, 0x2F, 0x50, 0x2F, 0x3D,
		0xDC, 0x6D, 0x28, 0x43, 0x58, 0xD1, 0x02, 0xED, 0x20, 0xD4, 0x7A,
		0x8A, 0xA7, 0x88, 0xA6, 0x2F, 0x0D, 0xB7, 0x80, 0x02, 0x21, 0x00,
		0xD1, 0x7B, 0x2D, 0x6F, 0xA8, 0x4D, 0xCA, 0xF1, 0xC9, 0x5D, 0x88,
		0xD7, 0xE7, 0xC3, 0x03, 0x85, 0xAE, 0xCF, 0x41, 0x55, 0x88, 0xD7,
		0x49, 0xAF, 0xD3, 0xEC, 0x81, 0xF6, 0x02, 0x2C, 0xEC, 0xD7, 0x01,
		0x41, 0x04, 0x62, 0xBB, 0x73, 0xF7, 0x6C, 0xA0, 0x99, 0x4F, 0xCB,
		0x8B, 0x42, 0x71, 0xE6, 0xFB, 0x75, 0x61, 0xF5, 0xC0, 0xF9, 0xCA,
		0x0C, 0xF6, 0x48, 0x52, 0x61, 0xC4, 0xA0, 0xDC, 0x89, 0x4F, 0x4A,
		0xB8, 0x44, 0xC6, 0xCD, 0xFB, 0x97, 0xCD, 0x0B, 0x60, 0xFF, 0xB5,
		0x01, 0x8F, 0xFD, 0x62, 0x38, 0xF4, 0xD8, 0x72, 0x70, 0xEF, 0xB1,
		0xD3, 0xAE, 0x37, 0x07, 0x9B, 0x79, 0x4A, 0x92, 0xD7, 0xEC, 0x95,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x01, 0x00, 0xC8,
		0x17, 0xA8, 0x04, 0x00, 0x00, 0x00, 0x19, 0x76, 0xA9, 0x14, 0xB6,
		0xEF, 0xD8, 0x0D, 0x99, 0x17, 0x9F, 0x4F, 0x4F, 0xF6, 0xF4, 0xDD,
		0x0A, 0x00, 0x7D, 0x01, 0x8C, 0x38, 0x5D, 0x21,
		0x88, 0xAC, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x01, 0x00, 0x00, 0x00, 0x01, 0x83, 0x45, 0x37, 0xB2, 0xF1, 0xCE, // Txs[4]
		0x8E, 0xF9, 0x37, 0x3A, 0x25, 0x8E, 0x10, 0x54, 0x5C, 0xE5, 0xA5,
		0x0B, 0x75, 0x8D, 0xF6, 0x16, 0xCD, 0x43, 0x56, 0xE0, 0x03, 0x25,
		0x54, 0xEB, 0xD3, 0xC4, 0x00, 0x00, 0x00, 0x00, 0x8B, 0x48, 0x30,
		0x45, 0x02, 0x21, 0x00, 0xE6, 0x8F, 0x42, 0x2D, 0xD7, 0xC3, 0x4F,
		0xDC, 0xE1, 0x1E, 0xEB, 0x45, 0x09, 0xDD, 0xAE, 0x38, 0x20, 0x17,
		0x73, 0xDD, 0x62, 0xF2, 0x84, 0xE8, 0xAA, 0x9D, 0x96, 0xF8, 0x50,
		0x99, 0xD0, 0xB0, 0x02, 0x20, 0x22, 0x43, 0xBD, 0x39, 0x9F, 0xF9,
		0x6B, 0x64, 0x9A, 0x0F, 0xAD, 0x05, 0xFA, 0x75, 0x9D, 0x6A, 0x88,
		0x2F, 0x0A, 0xF8, 0xC9, 0x0C, 0xF7, 0x63, 0x2C, 0x28, 0x40, 0xC2,
		0x90, 0x70, 0xAE, 0xC2, 0x01, 0x41, 0x04, 0x5E, 0x58, 0x06, 0x7E,
		0x81, 0x5C, 0x2F, 0x46, 0x4C, 0x6A, 0x2A, 0x15, 0xF9, 0x87, 0x75,
		0x83, 0x74, 0x20, 0x38, 0x95, 0x71, 0x0C, 0x2D, 0x45, 0x24, 0x42,
		0xE2, 0x84, 0x96, 0xFF, 0x38, 0xBA, 0x8F, 0x5F, 0xD9, 0x01, 0xDC,
		0x20, 0xE2, 0x9E, 0x88, 0x47, 0x71, 0x67, 0xFE, 0x4F, 0xC2, 0x99,
		0xBF, 0x81, 0x8F, 0xD0, 0xD9, 0xE1, 0x63, 0x2D, 0x46, 0x7B, 0x2A,
		0x3D, 0x95, 0x03, 0xB1, 0xAA, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0x02, 0x80, 0xD7, 0xE6, 0x36, 0x03, 0x00, 0x00, 0x00,
		0x19, 0x76, 0xA9, 0x14, 0xF3, 0x4C, 0x3E, 0x10, 0xEB, 0x38, 0x7E,
		0xFE, 0x87, 0x2A, 0xCB, 0x61, 0x4C, 0x89, 0xE7, 0x8B, 0xFC, 0xA7,
		0x81, 0x5D, 0x88, 0xAC, 0x40, 0x4B, 0x4C, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x19, 0x76, 0xA9, 0x14, 0xA8, 0x4E, 0x27, 0x29, 0x33, 0xAA,
		0xF8, 0x7E, 0x17, 0x15, 0xD7, 0x78, 0x6C, 0x51, 0xDF, 0xAE, 0xB5,
		0xB6, 0x5A, 0x6F, 0x88, 0xAC, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x01, 0x00, 0x00, 0x00, 0x01, 0x43, 0xAC, 0x81, 0xC8, 0xE6, 0xF6, // Txs[5]
		0xEF, 0x30, 0x7D, 0xFE, 0x17, 0xF3, 0xD9, 0x06, 0xD9, 0x99, 0xE2,
		0x3E, 0x01, 0x89, 0xFD, 0xA8, 0x38, 0xC5, 0x51, 0x0D, 0x85, 0x09,
		0x27, 0xE0, 0x3A, 0xE7, 0x00, 0x00, 0x00, 0x00, 0x8C, 0x49, 0x30,
		0x46, 0x02, 0x21, 0x00, 0x9C, 0x87, 0xC3, 0x44, 0x76, 0x0A, 0x64,
		0xCB, 0x8A, 0xE6, 0x68, 0x5A, 0x3E, 0xEC, 0x2C, 0x1A, 0xC1, 0xBE,
		0xD5, 0xB8, 0x8C, 0x87, 0xDE, 0x51, 0xAC, 0xD0, 0xE1, 0x24, 0xF2,
		0x66, 0xC1, 0x66, 0x02, 0x21, 0x00, 0x82, 0xD0, 0x7C, 0x03, 0x73,
		0x59, 0xC3, 0xA2, 0x57, 0xB5, 0xC6, 0x3E, 0xBD, 0x90, 0xF5, 0xA5,
		0xED, 0xF9, 0x7B, 0x2A, 0xC1, 0xC4, 0x34, 0xB0, 0x8C, 0xA9, 0x98,
		0x83, 0x9F, 0x34, 0x6D, 0xD4, 0x01, 0x41, 0x04, 0x0B, 0xA7, 0xE5,
		0x21, 0xFA, 0x79, 0x46, 0xD1, 0x2E, 0xDB, 0xB1, 0xD1, 0xE9, 0x5A,
		0x15, 0xC3, 0x4B, 0xD4, 0x39, 0x81, 0x95, 0xE8, 0x64, 0x33, 0xC9,
		0x2B, 0x43, 0x1C, 0xD3, 0x15, 0xF4, 0x55, 0xFE, 0x30, 0x03, 0x2E,
		0xDE, 0x69, 0xCA, 0xD9, 0xD1, 0xE1, 0xED, 0x6C, 0x3C, 0x4E, 0xC0,
		0xDB, 0xFC, 0xED, 0x53, 0x43, 0x8C, 0x62, 0x54, 0x62, 0xAF, 0xB7,
		0x92, 0xDC, 0xB0, 0x98, 0x54, 0x4B, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0x02, 0x40, 0x42, 0x0F, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x19, 0x76, 0xA9, 0x14, 0x46, 0x76, 0xD1, 0xB8, 0x20, 0xD6,
		0x3E, 0xC2, 0x72, 0xF1, 0x90, 0x0D, 0x59, 0xD4, 0x3B, 0xC6, 0x46,
		0x3D, 0x96, 0xF8, 0x88, 0xAC, 0x40, 0x42, 0x0F, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x19, 0x76, 0xA9, 0x14, 0x64, 0x8D, 0x04, 0x34, 0x1D,
		0x00, 0xD7, 0x96, 0x8B, 0x34, 0x05, 0xC0, 0x34, 0xAD, 0xC3, 0x8D,
		0x4D, 0x8F, 0xB9, 0xBD, 0x88, 0xAC, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00,
		0x01, 0x00, 0x00, 0x00, 0x02, 0x48, 0xCC, 0x91, 0x75, 0x01, 0xEA, // Txs[6]
		0x5C, 0x55, 0xF4, 0xA8, 0xD2, 0x00, 0x9C, 0x05, 0x67, 0xC4, 0x0C,
		0xFE, 0x03, 0x7C, 0x2E, 0x71, 0xAF, 0x01, 0x7D, 0x0A, 0x45, 0x2F,
		0xF7, 0x05, 0xE3, 0xF1, 0x00, 0x00, 0x00, 0x00, 0x8B, 0x48, 0x30,
		0x45, 0x02, 0x21, 0x00, 0xBF, 0x5F, 0xDC, 0x86, 0xDC, 0x5F, 0x08,
		0xA5, 0xD5, 0xC8, 0xE4, 0x3A, 0x8C, 0x9D, 0x5B, 0x1E, 0xD8, 0xC6,
		0x55, 0x62, 0xE2, 0x80, 0x00, 0x7B, 0x52, 0xB1, 0x33, 0x02, 0x1A,
		0xCD, 0x9A, 0xCC, 0x02, 0x20, 0x5E, 0x32, 0x5D, 0x61, 0x3E, 0x55,
		0x5F, 0x77, 0x28, 0x02, 0xBF, 0x41, 0x3D, 0x36, 0xBA, 0x80, 0x78,
		0x92, 0xED, 0x1A, 0x69, 0x0A, 0x77, 0x81, 0x1D, 0x30, 0x33, 0xB3,
		0xDE, 0x22, 0x6E, 0x0A, 0x01, 0x41, 0x04, 0x29, 0xFA, 0x71, 0x3B,
		0x12, 0x44, 0x84, 0xCB, 0x2B, 0xD7, 0xB5, 0x55, 0x7B, 0x2C, 0x0B,
		0x9D, 0xF7, 0xB2, 0xB1, 0xFE, 0xE6, 0x18, 0x25, 0xEA, 0xDC, 0x5A,
		0xE6, 0xC3, 0x7A, 0x99, 0x20, 0xD3, 0x8B, 0xFC, 0xCD, 0xC7, 0xDC,
		0x3C, 0xB0, 0xC4, 0x7D, 0x7B, 0x17, 0x3D, 0xBC, 0x9D, 0xB8, 0xD3,
		0x7D, 0xB0, 0xA3, 0x3A, 0xE4, 0x87, 0x98, 0x2C, 0x59, 0xC6, 0xF8,
		0x60, 0x6E, 0x9D, 0x17, 0x91, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0x41, 0xED, 0x70, 0x55, 0x1D, 0xD7, 0xE8, 0x41, 0x88,
		0x3A, 0xB8, 0xF0, 0xB1, 0x6B, 0xF0, 0x41, 0x76, 0xB7, 0xD1, 0x48,
		0x0E, 0x4F, 0x0A, 0xF9, 0xF3, 0xD4, 0xC3, 0x59, 0x57, 0x68, 0xD0,
		0x68, 0x00, 0x00, 0x00, 0x00, 0x8B, 0x48, 0x30, 0x45, 0x02, 0x21,
		0x00, 0x85, 0x13, 0xAD, 0x65, 0x18, 0x7B, 0x90, 0x3A, 0xED, 0x11,
		0x02, 0xD1, 0xD0, 0xC4,
		0x76, 0x88, 0x12, 0x76, 0x58, 0xC5, 0x11, 0x06, 0x75, 0x3F, 0xED,
		0x01, 0x51, 0xCE, 0x9C, 0x16, 0xB8, 0x09, 0x02, 0x20, 0x14, 0x32,
		0xB9, 0xEB, 0xCB, 0x87, 0xBD, 0x04, 0xCE, 0xB2, 0xDE, 0x66, 0x03,
		0x5F, 0xBB, 0xAF, 0x4B, 0xF8, 0xB0, 0x0D, 0x1C, 0xFE, 0x41, 0xF1,
		0xA1, 0xF7, 0x33, 0x8F, 0x9A, 0xD7, 0x9D, 0x21, 0x01, 0x41, 0x04,
		0x9D, 0x4C, 0xF8, 0x01, 0x25, 0xBF, 0x50, 0xBE, 0x17, 0x09, 0xF7,
		0x18, 0xC0, 0x7A, 0xD1, 0x5D, 0x0F, 0xC6, 0x12, 0xB7, 0xDA, 0x1F,
		0x55, 0x70, 0xDD, 0xDC, 0x35, 0xF2, 0xA3, 0x52, 0xF0, 0xF2, 0x7C,
		0x97, 0x8B, 0x06, 0x82, 0x0E, 0xDC, 0xA9, 0xEF, 0x98, 0x2C, 0x35,
		0xFD, 0xA2, 0xD2, 0x55, 0xAF, 0xBA, 0x34, 0x00, 0x68, 0xC5, 0x03,
		0x55, 0x52, 0x36, 0x8B, 0xC7, 0x20, 0x0C, 0x14, 0x88, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x01, 0x00, 0x09, 0x3D, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x19, 0x76, 0xA9, 0x14, 0x8E, 0xDB, 0x68,
		0x82, 0x2F, 0x1A, 0xD5, 0x80, 0xB0, 0x43, 0xC7, 0xB3, 0xDF, 0x2E,
		0x40, 0x0F, 0x86, 0x99, 0xEB, 0x48, 0x88, 0xAC, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00,
	}
	block, err := util.NewBlockFromBytes(blockBytes)
	if err != nil {
		t.Errorf("TestFilterInsertP2PubKeyOnly NewBlockFromBytes failed: %v", err)
		return
	}

	f := bloom.NewFilter(10, 0, 0.000001, wire.BloomUpdateP2PubkeyOnly)

	// Generation pubkey
	inputStr := "04eaafc2314def4ca98ac970241bcab022b9c1e1f4ea423a20f134c" +
		"876f2c01ec0f0dd5b2e86e7168cefe0d81113c3807420ce13ad1357231a" +
		"2252247d97a46a91"
	inputBytes, err := hex.DecodeString(inputStr)
	if err != nil {
		t.Errorf("TestFilterInsertP2PubKeyOnly DecodeString failed: %v", err)
		return
	}
	f.Add(inputBytes)

	// Public key hash of 4th transaction
	inputStr = "b6efd80d99179f4f4ff6f4dd0a007d018c385d21"
	inputBytes, err = hex.DecodeString(inputStr)
	if err != nil {
		t.Errorf("TestFilterInsertP2PubKeyOnly DecodeString failed: %v", err)
		return
	}
	f.Add(inputBytes)

	// Ignore return value -- this is just used to update the filter.
	_, _ = bloom.NewMerkleBlock(block, f)

	// We should match the generation pubkey
	inputStr = "c2254e4d610867ee48decf60d8bd8e1d361eeeab5d1052ce3e98184a5b4d0923" //0st tx ID
	txID, err := daghash.NewTxIDFromStr(inputStr)
	if err != nil {
		t.Errorf("TestFilterInsertP2PubKeyOnly NewHashFromStr failed: %v", err)
		return
	}
	outpoint := wire.NewOutpoint(txID, 0)
	if !f.MatchesOutpoint(outpoint) {
		t.Errorf("TestFilterInsertP2PubKeyOnly didn't match the generation "+
			"outpoint %s", inputStr)
		return
	}

	// We should not match the 4th transaction, which is not p2pk
	inputStr = "f9a116ecc107b6b1b0bdcd0d727bfaa3355f27f8fed08347bf0004244949d9eb"
	txID, err = daghash.NewTxIDFromStr(inputStr)
	if err != nil {
		t.Errorf("TestFilterInsertP2PubKeyOnly NewHashFromStr failed: %v", err)
		return
	}
	outpoint = wire.NewOutpoint(txID, 0)
	if f.MatchesOutpoint(outpoint) {
		t.Errorf("TestFilterInsertP2PubKeyOnly matched outpoint %s", inputStr)
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
