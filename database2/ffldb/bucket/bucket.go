package bucket

var separator = []byte("/")

// BuildKey builds a key using the given key value and the
// given path of buckets.
// Example:
// * key: aaa
// * buckets: bbb, ccc
// * Result: bbb/ccc/aaa
func BuildKey(key []byte, buckets ...[]byte) []byte {
	bucketPath := BuildBucketPath(buckets...)

	fullKeyLength := len(bucketPath) + len(key)
	fullKey := make([]byte, fullKeyLength)
	copy(fullKey, bucketPath)
	copy(fullKey[len(bucketPath):], key)

	return fullKey
}

// BuildBucketPath builds a compound path using the given
// path of buckets.
// Example:
// * buckets: bbb, ccc
// * Result: bbb/ccc/
func BuildBucketPath(buckets ...[]byte) []byte {
	bucketPathlength := (len(buckets)) * len(separator) // length of all the separators
	for _, bucket := range buckets {
		bucketPathlength += len(bucket)
	}

	bucketPath := make([]byte, bucketPathlength)
	offset := 0
	for _, bucket := range buckets {
		copy(bucketPath[offset:], bucket)
		offset += len(bucket)
		copy(bucketPath[offset:], separator)
		offset += len(separator)
	}

	return bucketPath
}
