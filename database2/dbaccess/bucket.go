package dbaccess

var separator = []byte("/")

func buildKey(buckets ...[]byte) []byte {
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

func buildBucketKey(buckets ...[]byte) []byte {
	key := buildKey(buckets...)
	size := len(key) + len(separator)
	bucketKey := make([]byte, size)

	copy(bucketKey, key)
	copy(bucketKey[len(key):], separator)

	return bucketKey
}
