package bucket

var separator = []byte("/")

func BuildKey(buckets ...[]byte) []byte {
	size := (len(buckets) - 1) * len(separator) // initialized to include the size of the separators
	for _, bucket := range buckets {
		size += len(bucket)
	}

	key := make([]byte, size)
	offset := 0
	for i, bucket := range buckets {
		copy(key[offset:], bucket)
		offset += len(bucket)
		if i == len(buckets)-1 {
			break
		}
		copy(key[offset:], separator)
		offset += len(separator)
	}

	return key
}

func BuildBucketKey(buckets ...[]byte) []byte {
	key := BuildKey(buckets...)
	size := len(key) + len(separator)
	bucketKey := make([]byte, size)

	copy(bucketKey, key)
	copy(bucketKey[len(key):], separator)

	return bucketKey
}
