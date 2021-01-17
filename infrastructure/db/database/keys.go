package database

import (
	"encoding/hex"
)

var bucketSeparator = byte('/')

// Key is a helper type meant to combine prefix
// and suffix into a single database key.
type Key struct {
	bucket *Bucket
	suffix []byte
}

// Bytes returns the full key bytes that are consisted
// from the bucket path concatenated to the suffix.
func (k *Key) Bytes() []byte {
	bucketPath := k.bucket.Path()
	keyBytes := make([]byte, len(bucketPath)+len(k.suffix))
	copy(keyBytes, bucketPath)
	copy(keyBytes[len(bucketPath):], k.suffix)
	return keyBytes
}

func (k *Key) String() string {
	return hex.EncodeToString(k.Bytes())
}

// Bucket returns the key bucket.
func (k *Key) Bucket() *Bucket {
	return k.bucket
}

// Suffix returns the key suffix.
func (k *Key) Suffix() []byte {
	return k.suffix
}

// newKey returns a new key composed
// of the given bucket and suffix
func newKey(bucket *Bucket, suffix []byte) *Key {
	return &Key{bucket: bucket, suffix: suffix}
}

// Bucket is a helper type meant to combine buckets
// and sub-buckets that can be used to create database
// keys and prefix-based cursors.
type Bucket struct {
	path []byte
}

// MakeBucket creates a new Bucket using the given path
// of buckets.
func MakeBucket(path []byte) *Bucket {
	if len(path) > 0 {
		path = append(path, bucketSeparator)
	}
	return &Bucket{path: path}
}

// Bucket returns the sub-bucket of the current bucket
// defined by bucketBytes.
func (b *Bucket) Bucket(bucketBytes []byte) *Bucket {
	newPath := make([]byte, 0, len(b.path)+len(bucketBytes)+1) // +1 for the separator in MakeBucket
	newPath = append(newPath, b.path...)
	newPath = append(newPath, bucketBytes...)
	return MakeBucket(newPath)
}

// Key returns a key in the current bucket with the
// given suffix.
func (b *Bucket) Key(suffix []byte) *Key {
	return newKey(b, suffix)
}

// Path returns the full path of the current bucket.
func (b *Bucket) Path() []byte {
	return b.path
}
