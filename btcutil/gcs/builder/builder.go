// Copyright (c) 2017 The btcsuite developers
// Copyright (c) 2017 The Lightning Network Developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package builder

import (
	"crypto/rand"
	"encoding/binary"

	"github.com/daglabs/btcd/txscript"
	"github.com/daglabs/btcd/wire"
	"github.com/daglabs/btcd/btcutil/gcs"
	"github.com/daglabs/btcd/dagconfig/daghash"
)

// DefaultP is the default collision probability (2^-20)
const DefaultP = 20

// GCSBuilder is a utility class that makes building GCS filters convenient.
type GCSBuilder struct {
	p   uint8
	key [gcs.KeySize]byte

	// data is a set of entries represented as strings. This is done to
	// deduplicate items as they are added.
	data map[string]struct{}
	err  error
}

// RandomKey is a utility function that returns a cryptographically random
// [gcs.KeySize]byte usable as a key for a GCS filter.
func RandomKey() ([gcs.KeySize]byte, error) {
	var key [gcs.KeySize]byte

	// Read a byte slice from rand.Reader.
	randKey := make([]byte, gcs.KeySize)
	_, err := rand.Read(randKey)

	// This shouldn't happen unless the user is on a system that doesn't
	// have a system CSPRNG. OK to panic in this case.
	if err != nil {
		return key, err
	}

	// Copy the byte slice to a [gcs.KeySize]byte array and return it.
	copy(key[:], randKey[:])
	return key, nil
}

// DeriveKey is a utility function that derives a key from a daghash.Hash by
// truncating the bytes of the hash to the appopriate key size.
func DeriveKey(keyHash *daghash.Hash) [gcs.KeySize]byte {
	var key [gcs.KeySize]byte
	copy(key[:], keyHash.CloneBytes()[:])
	return key
}

// OutPointToFilterEntry is a utility function that derives a filter entry from
// a wire.OutPoint in a standardized way for use with both building and
// querying filters.
func OutPointToFilterEntry(outpoint wire.OutPoint) []byte {
	// Size of the hash plus size of int32 index
	data := make([]byte, daghash.HashSize+4)
	copy(data[:], outpoint.Hash.CloneBytes()[:])
	binary.LittleEndian.PutUint32(data[daghash.HashSize:], outpoint.Index)
	return data
}

// Key retrieves the key with which the builder will build a filter. This is
// useful if the builder is created with a random initial key.
func (b *GCSBuilder) Key() ([gcs.KeySize]byte, error) {
	// Do nothing if the builder's errored out.
	if b.err != nil {
		return [gcs.KeySize]byte{}, b.err
	}

	return b.key, nil
}

// SetKey sets the key with which the builder will build a filter to the passed
// [gcs.KeySize]byte.
func (b *GCSBuilder) SetKey(key [gcs.KeySize]byte) *GCSBuilder {
	// Do nothing if the builder's already errored out.
	if b.err != nil {
		return b
	}

	copy(b.key[:], key[:])
	return b
}

// SetKeyFromHash sets the key with which the builder will build a filter to a
// key derived from the passed daghash.Hash using DeriveKey().
func (b *GCSBuilder) SetKeyFromHash(keyHash *daghash.Hash) *GCSBuilder {
	// Do nothing if the builder's already errored out.
	if b.err != nil {
		return b
	}

	return b.SetKey(DeriveKey(keyHash))
}

// SetP sets the filter's probability after calling Builder().
func (b *GCSBuilder) SetP(p uint8) *GCSBuilder {
	// Do nothing if the builder's already errored out.
	if b.err != nil {
		return b
	}

	// Basic sanity check.
	if p > 32 {
		b.err = gcs.ErrPTooBig
		return b
	}

	b.p = p
	return b
}

// Preallocate sets the estimated filter size after calling Builder() to reduce
// the probability of memory reallocations. If the builder has already had data
// added to it, Preallocate has no effect.
func (b *GCSBuilder) Preallocate(n uint32) *GCSBuilder {
	// Do nothing if the builder's already errored out.
	if b.err != nil {
		return b
	}

	if b.data == nil {
		b.data = make(map[string]struct{}, n)
	}

	return b
}

// AddEntry adds a []byte to the list of entries to be included in the GCS
// filter when it's built.
func (b *GCSBuilder) AddEntry(data []byte) *GCSBuilder {
	// Do nothing if the builder's already errored out.
	if b.err != nil {
		return b
	}

	b.data[string(data)] = struct{}{}
	return b
}

// AddEntries adds all the []byte entries in a [][]byte to the list of entries
// to be included in the GCS filter when it's built.
func (b *GCSBuilder) AddEntries(data [][]byte) *GCSBuilder {
	// Do nothing if the builder's already errored out.
	if b.err != nil {
		return b
	}

	for _, entry := range data {
		b.AddEntry(entry)
	}
	return b
}

// AddOutPoint adds a wire.OutPoint to the list of entries to be included in
// the GCS filter when it's built.
func (b *GCSBuilder) AddOutPoint(outpoint wire.OutPoint) *GCSBuilder {
	// Do nothing if the builder's already errored out.
	if b.err != nil {
		return b
	}

	return b.AddEntry(OutPointToFilterEntry(outpoint))
}

// AddHash adds a daghash.Hash to the list of entries to be included in the
// GCS filter when it's built.
func (b *GCSBuilder) AddHash(hash *daghash.Hash) *GCSBuilder {
	// Do nothing if the builder's already errored out.
	if b.err != nil {
		return b
	}

	return b.AddEntry(hash.CloneBytes())
}

// AddScript adds all the data pushed in the script serialized as the passed
// []byte to the list of entries to be included in the GCS filter when it's
// built.
func (b *GCSBuilder) AddScript(script []byte) *GCSBuilder {
	// Do nothing if the builder's already errored out.
	if b.err != nil {
		return b
	}

	// Ignore errors and add pushed data, if any
	data, _ := txscript.PushedData(script)
	if len(data) == 0 {
		return b
	}

	return b.AddEntries(data)
}

// Build returns a function which builds a GCS filter with the given parameters
// and data.
func (b *GCSBuilder) Build() (*gcs.Filter, error) {
	// Do nothing if the builder's already errored out.
	if b.err != nil {
		return nil, b.err
	}

	dataSlice := make([][]byte, 0, len(b.data))
	for item := range b.data {
		dataSlice = append(dataSlice, []byte(item))
	}

	return gcs.BuildGCSFilter(b.p, b.key, dataSlice)
}

// WithKeyPN creates a GCSBuilder with specified key and the passed probability
// and estimated filter size.
func WithKeyPN(key [gcs.KeySize]byte, p uint8, n uint32) *GCSBuilder {
	b := GCSBuilder{}
	return b.SetKey(key).SetP(p).Preallocate(n)
}

// WithKeyP creates a GCSBuilder with specified key and the passed probability.
// Estimated filter size is set to zero, which means more reallocations are
// done when building the filter.
func WithKeyP(key [gcs.KeySize]byte, p uint8) *GCSBuilder {
	return WithKeyPN(key, p, 0)
}

// WithKey creates a GCSBuilder with specified key. Probability is set to
// 20 (2^-20 collision probability). Estimated filter size is set to zero, which
// means more reallocations are done when building the filter.
func WithKey(key [gcs.KeySize]byte) *GCSBuilder {
	return WithKeyPN(key, DefaultP, 0)
}

// WithKeyHashPN creates a GCSBuilder with key derived from the specified
// daghash.Hash and the passed probability and estimated filter size.
func WithKeyHashPN(keyHash *daghash.Hash, p uint8, n uint32) *GCSBuilder {
	return WithKeyPN(DeriveKey(keyHash), p, n)
}

// WithKeyHashP creates a GCSBuilder with key derived from the specified
// daghash.Hash and the passed probability. Estimated filter size is set to
// zero, which means more reallocations are done when building the filter.
func WithKeyHashP(keyHash *daghash.Hash, p uint8) *GCSBuilder {
	return WithKeyHashPN(keyHash, p, 0)
}

// WithKeyHash creates a GCSBuilder with key derived from the specified
// daghash.Hash. Probability is set to 20 (2^-20 collision probability).
// Estimated filter size is set to zero, which means more reallocations are
// done when building the filter.
func WithKeyHash(keyHash *daghash.Hash) *GCSBuilder {
	return WithKeyHashPN(keyHash, DefaultP, 0)
}

// WithRandomKeyPN creates a GCSBuilder with a cryptographically random key and
// the passed probability and estimated filter size.
func WithRandomKeyPN(p uint8, n uint32) *GCSBuilder {
	key, err := RandomKey()
	if err != nil {
		b := GCSBuilder{err: err}
		return &b
	}
	return WithKeyPN(key, p, n)
}

// WithRandomKeyP creates a GCSBuilder with a cryptographically random key and
// the passed probability. Estimated filter size is set to zero, which means
// more reallocations are done when building the filter.
func WithRandomKeyP(p uint8) *GCSBuilder {
	return WithRandomKeyPN(p, 0)
}

// WithRandomKey creates a GCSBuilder with a cryptographically random key.
// Probability is set to 20 (2^-20 collision probability). Estimated filter
// size is set to zero, which means more reallocations are done when
// building the filter.
func WithRandomKey() *GCSBuilder {
	return WithRandomKeyPN(DefaultP, 0)
}

// BuildBasicFilter builds a basic GCS filter from a block. A basic GCS filter
// will contain all the previous outpoints spent within a block, as well as the
// data pushes within all the outputs created within a block.
func BuildBasicFilter(block *wire.MsgBlock) (*gcs.Filter, error) {
	blockHash := block.BlockHash()
	b := WithKeyHash(&blockHash)

	// If the filter had an issue with the specified key, then we force it
	// to bubble up here by calling the Key() function.
	_, err := b.Key()
	if err != nil {
		return nil, err
	}

	// In order to build a basic filter, we'll range over the entire block,
	// adding the outpoint data as well as the data pushes within the
	// pkScript.
	for i, tx := range block.Transactions {
		// First we'll compute the bash of the transaction and add that
		// directly to the filter.
		txHash := tx.TxHash()
		b.AddHash(&txHash)

		// Skip the inputs for the coinbase transaction
		if i != 0 {
			// Each each txin, we'll add a serialized version of
			// the txid:index to the filters data slices.
			for _, txIn := range tx.TxIn {
				b.AddOutPoint(txIn.PreviousOutPoint)
			}
		}

		// For each output in a transaction, we'll add each of the
		// individual data pushes within the script.
		for _, txOut := range tx.TxOut {
			b.AddEntry(txOut.PkScript)
		}
	}

	return b.Build()
}

// BuildExtFilter builds an extended GCS filter from a block. An extended
// filter supplements a regular basic filter by include all the scriptSig data
// found within a block. This includes all the data pushes within any signature
// scripts as well as each element of an input's scriptSig stack. Additionally,
// the _hashes_ of each transaction are also inserted into the filter.
func BuildExtFilter(block *wire.MsgBlock) (*gcs.Filter, error) {
	blockHash := block.BlockHash()
	b := WithKeyHash(&blockHash)

	// If the filter had an issue with the specified key, then we force it
	// to bubble up here by calling the Key() function.
	_, err := b.Key()
	if err != nil {
		return nil, err
	}

	// In order to build an extended filter, we add the hash of each
	// transaction as well as each piece of data included in the sigScript input.
	for i, tx := range block.Transactions {
		// Skip the inputs for the coinbase transaction
		if i != 0 {
			// Next, for each input, we'll add the sigScript (if it's present)
			for _, txIn := range tx.TxIn {
				if txIn.SignatureScript != nil {
					b.AddScript(txIn.SignatureScript)
				}
			}
		}
	}

	return b.Build()
}

// GetFilterHash returns the double-SHA256 of the filter.
func GetFilterHash(filter *gcs.Filter) (daghash.Hash, error) {
	filterData, err := filter.NBytes()
	if err != nil {
		return daghash.Hash{}, err
	}

	return daghash.DoubleHashH(filterData), nil
}

// MakeHeaderForFilter makes a filter chain header for a filter, given the
// filter and the previous filter chain header.
func MakeHeaderForFilter(filter *gcs.Filter, prevHeader daghash.Hash) (daghash.Hash, error) {
	filterTip := make([]byte, 2*daghash.HashSize)
	filterHash, err := GetFilterHash(filter)
	if err != nil {
		return daghash.Hash{}, err
	}

	// In the buffer we created above we'll compute hash || prevHash as an
	// intermediate value.
	copy(filterTip, filterHash[:])
	copy(filterTip[daghash.HashSize:], prevHeader[:])

	// The final filter hash is the double-sha256 of the hash computed
	// above.
	return daghash.DoubleHashH(filterTip), nil
}
