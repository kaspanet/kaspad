// Copyright (c) 2017 The btcsuite developers
// Copyright (c) 2017 The Lightning Network Developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package builder_test

import (
	"encoding/hex"
	"testing"

	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/txscript"
	"github.com/daglabs/btcd/wire"
	"github.com/daglabs/btcd/btcutil"
	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/btcutil/gcs/builder"
	"github.com/daglabs/btcd/btcutil/gcs"
)

var (
	// No need to allocate an err variable in every test
	err error

	// List of values for building a filter
	contents = [][]byte{
		[]byte("Alex"),
		[]byte("Bob"),
		[]byte("Charlie"),
		[]byte("Dick"),
		[]byte("Ed"),
		[]byte("Frank"),
		[]byte("George"),
		[]byte("Harry"),
		[]byte("Ilya"),
		[]byte("John"),
		[]byte("Kevin"),
		[]byte("Larry"),
		[]byte("Michael"),
		[]byte("Nate"),
		[]byte("Owen"),
		[]byte("Paul"),
		[]byte("Quentin"),
	}

	testKey = [16]byte{0x4c, 0xb1, 0xab, 0x12, 0x57, 0x62, 0x1e, 0x41,
		0x3b, 0x8b, 0x0e, 0x26, 0x64, 0x8d, 0x4a, 0x15}

	testHash = "000000000000000000496d7ff9bd2c96154a8d64260e8b3b411e625712abb14c"

	testAddr = "dagcoin:pr54mwe99q7vxh22dth6wmqrstnrec86xctaeyyphz"
)

// TestUseBlockHash tests using a block hash as a filter key.
func TestUseBlockHash(t *testing.T) {
	// Block hash #448710, pretty high difficulty.
	hash, err := daghash.NewHashFromStr(testHash)
	if err != nil {
		t.Fatalf("Hash from string failed: %s", err.Error())
	}

	// wire.OutPoint
	outPoint := wire.OutPoint{
		Hash:  *hash,
		Index: 4321,
	}

	// btcutil.Address
	addr, err := btcutil.DecodeAddress(testAddr, &dagconfig.MainNetParams)
	if err != nil {
		t.Fatalf("Address decode failed: %s", err.Error())
	}
	addrBytes, err := txscript.PayToAddrScript(addr)
	if err != nil {
		t.Fatalf("Address script build failed: %s", err.Error())
	}

	// Create a GCSBuilder with a key hash and check that the key is derived
	// correctly, then test it.
	b := builder.WithKeyHash(hash)
	key, err := b.Key()
	if err != nil {
		t.Fatalf("Builder instantiation with key hash failed: %s",
			err.Error())
	}
	if key != testKey {
		t.Fatalf("Key not derived correctly from key hash:\n%s\n%s",
			hex.EncodeToString(key[:]),
			hex.EncodeToString(testKey[:]))
	}
	BuilderTest(b, hash, builder.DefaultP, outPoint, addrBytes, t)

	// Create a GCSBuilder with a key hash and non-default P and test it.
	b = builder.WithKeyHashP(hash, 30)
	BuilderTest(b, hash, 30, outPoint, addrBytes, t)

	// Create a GCSBuilder with a random key, set the key from a hash
	// manually, check that the key is correct, and test it.
	b = builder.WithRandomKey()
	b.SetKeyFromHash(hash)
	key, err = b.Key()
	if err != nil {
		t.Fatalf("Builder instantiation with known key failed: %s",
			err.Error())
	}
	if key != testKey {
		t.Fatalf("Key not copied correctly from known key:\n%s\n%s",
			hex.EncodeToString(key[:]),
			hex.EncodeToString(testKey[:]))
	}
	BuilderTest(b, hash, builder.DefaultP, outPoint, addrBytes, t)

	// Create a GCSBuilder with a random key and test it.
	b = builder.WithRandomKey()
	key1, err := b.Key()
	if err != nil {
		t.Fatalf("Builder instantiation with random key failed: %s",
			err.Error())
	}
	t.Logf("Random Key 1: %s", hex.EncodeToString(key1[:]))
	BuilderTest(b, hash, builder.DefaultP, outPoint, addrBytes, t)

	// Create a GCSBuilder with a random key and non-default P and test it.
	b = builder.WithRandomKeyP(30)
	key2, err := b.Key()
	if err != nil {
		t.Fatalf("Builder instantiation with random key failed: %s",
			err.Error())
	}
	t.Logf("Random Key 2: %s", hex.EncodeToString(key2[:]))
	if key2 == key1 {
		t.Fatalf("Random keys are the same!")
	}
	BuilderTest(b, hash, 30, outPoint, addrBytes, t)

	// Create a GCSBuilder with a known key and test it.
	b = builder.WithKey(testKey)
	key, err = b.Key()
	if err != nil {
		t.Fatalf("Builder instantiation with known key failed: %s",
			err.Error())
	}
	if key != testKey {
		t.Fatalf("Key not copied correctly from known key:\n%s\n%s",
			hex.EncodeToString(key[:]),
			hex.EncodeToString(testKey[:]))
	}
	BuilderTest(b, hash, builder.DefaultP, outPoint, addrBytes, t)

	// Create a GCSBuilder with a known key and non-default P and test it.
	b = builder.WithKeyP(testKey, 30)
	key, err = b.Key()
	if err != nil {
		t.Fatalf("Builder instantiation with known key failed: %s",
			err.Error())
	}
	if key != testKey {
		t.Fatalf("Key not copied correctly from known key:\n%s\n%s",
			hex.EncodeToString(key[:]),
			hex.EncodeToString(testKey[:]))
	}
	BuilderTest(b, hash, 30, outPoint, addrBytes, t)

	// Create a GCSBuilder with a known key and too-high P and ensure error
	// works throughout all functions that use it.
	b = builder.WithRandomKeyP(33).SetKeyFromHash(hash).SetKey(testKey)
	b.SetP(30).AddEntry(hash.CloneBytes()).AddEntries(contents)
	b.AddOutPoint(outPoint).AddHash(hash).AddScript(addrBytes)
	_, err = b.Key()
	if err != gcs.ErrPTooBig {
		t.Fatalf("No error on P too big!")
	}
	_, err = b.Build()
	if err != gcs.ErrPTooBig {
		t.Fatalf("No error on P too big!")
	}
}

func BuilderTest(b *builder.GCSBuilder, hash *daghash.Hash, p uint8,
	outPoint wire.OutPoint, addrBytes []byte, t *testing.T) {

	key, err := b.Key()
	if err != nil {
		t.Fatalf("Builder instantiation with key hash failed: %s",
			err.Error())
	}

	// Build a filter and test matches.
	b.AddEntries(contents)
	f, err := b.Build()
	if err != nil {
		t.Fatalf("Filter build failed: %s", err.Error())
	}
	if f.P() != p {
		t.Fatalf("Filter built with wrong probability")
	}
	match, err := f.Match(key, []byte("Nate"))
	if err != nil {
		t.Fatalf("Filter match failed: %s", err)
	}
	if !match {
		t.Fatal("Filter didn't match when it should have!")
	}
	match, err = f.Match(key, []byte("weks"))
	if err != nil {
		t.Fatalf("Filter match failed: %s", err)
	}
	if match {
		t.Logf("False positive match, should be 1 in 2**%d!",
			builder.DefaultP)
	}

	// Add a hash, build a filter, and test matches
	b.AddHash(hash)
	f, err = b.Build()
	if err != nil {
		t.Fatalf("Filter build failed: %s", err.Error())
	}
	match, err = f.Match(key, hash.CloneBytes())
	if err != nil {
		t.Fatalf("Filter match failed: %s", err)
	}
	if !match {
		t.Fatal("Filter didn't match when it should have!")
	}

	// Add a wire.OutPoint, build a filter, and test matches
	b.AddOutPoint(outPoint)
	f, err = b.Build()
	if err != nil {
		t.Fatalf("Filter build failed: %s", err.Error())
	}
	match, err = f.Match(key, hash.CloneBytes())
	if err != nil {
		t.Fatalf("Filter match failed: %s", err)
	}
	if !match {
		t.Fatal("Filter didn't match when it should have!")
	}

	// Add a script, build a filter, and test matches
	b.AddScript(addrBytes)
	f, err = b.Build()
	if err != nil {
		t.Fatalf("Filter build failed: %s", err.Error())
	}
	pushedData, err := txscript.PushedData(addrBytes)
	if err != nil {
		t.Fatalf("Couldn't extract pushed data from addrBytes script: %s",
			err.Error())
	}
	match, err = f.MatchAny(key, pushedData)
	if err != nil {
		t.Fatalf("Filter match any failed: %s", err)
	}
	if !match {
		t.Fatal("Filter didn't match when it should have!")
	}

	// Check that adding duplicate items does not increase filter size.
	originalSize := f.N()
	b.AddScript(addrBytes)
	f, err = b.Build()
	if err != nil {
		t.Fatalf("Filter build failed: %s", err.Error())
	}
	if f.N() != originalSize {
		t.Fatal("Filter size increased with duplicate items")
	}
}
