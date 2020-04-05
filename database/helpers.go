package database

import (
	"bytes"
)

var separator = []byte("/")

// Key is a helper type meant to combine prefix
// and keys into a single full key-value
// database key.
type Key struct {
	prefix, key []byte
}

// FullKey returns the prefix concatenated to the key.
func (k *Key) FullKey() []byte {
	keyPath := make([]byte, len(k.prefix)+len(k.key))
	copy(keyPath, k.prefix)
	copy(keyPath[len(k.prefix):], k.key)
	return keyPath
}

func (k *Key) String() string {
	return string(k.FullKey())
}

// Key returns the key part of the key.
func (k *Key) Key() []byte {
	return k.key
}

// NewKey returns a new key composed
// of the given prefix and key
func NewKey(prefix, key []byte) *Key {
	return &Key{prefix: prefix, key: key}
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
	bucketPath := bytes.Join(b.path, separator)

	bucketPathWithFinalSeparator := make([]byte, len(bucketPath)+len(separator))
	copy(bucketPathWithFinalSeparator, bucketPath)
	copy(bucketPathWithFinalSeparator[len(bucketPath):], separator)

	return bucketPathWithFinalSeparator
}
