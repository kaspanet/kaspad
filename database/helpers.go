package database

import (
	"bytes"
	"encoding/hex"
)

var bucketSeparator = []byte("/")

// Key is a helper type meant to combine prefix
// and suffix into a single full key-value
// database key.
type Key struct {
	prefix, suffix []byte
}

// Bytes returns the prefix concatenated to the key.
func (k *Key) Bytes() []byte {
	keyBytes := make([]byte, len(k.prefix)+len(k.suffix))
	copy(keyBytes, k.prefix)
	copy(keyBytes[len(k.prefix):], k.suffix)
	return keyBytes
}

func (k *Key) String() string {
	return hex.EncodeToString(k.Bytes())
}

// Prefix returns the prefix part of the key.
func (k *Key) Prefix() []byte {
	return k.prefix
}

// Suffix returns the suffix part of the key.
func (k *Key) Suffix() []byte {
	return k.suffix
}

// NewKey returns a new key composed
// of the given prefix and suffix
func NewKey(prefix, suffix []byte) *Key {
	return &Key{prefix: prefix, suffix: suffix}
}

// Bucket is a helper type meant to combine buckets
// and sub-buckets that can be used to create database
// keys and prefix-based cursors.
type Bucket struct {
	path [][]byte
}

// MakeBucket creates a new Bucket using the given path
// of buckets.
func MakeBucket(path ...[]byte) *Bucket {
	return &Bucket{path: path}
}

// Bucket returns the sub-bucket of the current bucket
// defined by bucketBytes.
func (b *Bucket) Bucket(bucketBytes []byte) *Bucket {
	newPath := make([][]byte, len(b.path)+1)
	copy(newPath, b.path)
	copy(newPath[len(b.path):], [][]byte{bucketBytes})

	return MakeBucket(newPath...)
}

// Key returns the key inside of the current bucket.
func (b *Bucket) Key(key []byte) *Key {
	return NewKey(b.Path(), key)
}

// Path returns the full path of the current bucket.
func (b *Bucket) Path() []byte {
	bucketPath := bytes.Join(b.path, bucketSeparator)

	bucketPathWithFinalSeparator := make([]byte, len(bucketPath)+len(bucketSeparator))
	copy(bucketPathWithFinalSeparator, bucketPath)
	copy(bucketPathWithFinalSeparator[len(bucketPath):], bucketSeparator)

	return bucketPathWithFinalSeparator
}
