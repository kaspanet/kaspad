package database2

import "bytes"

var separator = []byte("/")

// Bucket is a helper type meant to combine buckets,
// sub-buckets, and keys into a single full key-value
// database key.
type Bucket [][]byte

// MakeBucket creates a new Bucket using the given path
// of buckets.
func MakeBucket(path ...[]byte) Bucket {
	return path
}

// Bucket returns the sub-bucket of the current bucket
// defined by bucketBytes.
func (b Bucket) Bucket(bucketBytes []byte) Bucket {
	return append(b, bucketBytes)
}

// Key returns the key inside of the current bucket.
func (b Bucket) Key(key []byte) []byte {
	bucketPath := b.Path()

	fullKeyLength := len(bucketPath) + len(key)
	fullKey := make([]byte, fullKeyLength)
	copy(fullKey, bucketPath)
	copy(fullKey[len(bucketPath):], key)

	return fullKey
}

// Path returns the full path of the current bucket.
func (b Bucket) Path() []byte {
	bucketPath := bytes.Join(b, separator)

	bucketPathWithFinalSeperator := make([]byte, len(bucketPath)+len(separator))
	copy(bucketPathWithFinalSeperator, bucketPath)
	copy(bucketPathWithFinalSeperator[len(bucketPath):], separator)

	return bucketPathWithFinalSeperator
}
