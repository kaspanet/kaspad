package bucket

var separator = []byte("/")

func BuildKey(key []byte, buckets ...[]byte) []byte {
	bucketPath := BuildBucketPath(buckets...)

	fullKeyLength := len(bucketPath) + len(key)
	fullKey := make([]byte, fullKeyLength)
	copy(fullKey, bucketPath)
	copy(fullKey[len(bucketPath):], key)

	return fullKey
}

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
