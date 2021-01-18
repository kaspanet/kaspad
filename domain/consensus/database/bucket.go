package database

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
)

func dbBucketToDatabaseBucket(bucket model.DBBucket) *database.Bucket {
	if bucket, ok := bucket.(dbBucket); ok {
		return bucket.bucket
	}
	// This assumes that MakeBucket(src).Path() == src. which is not promised anywhere.
	return database.MakeBucket(bucket.Path())
}

// MakeBucket creates a new Bucket using the given path of buckets.
func MakeBucket(path []byte) model.DBBucket {
	return dbBucket{bucket: database.MakeBucket(path)}
}

type dbBucket struct {
	bucket *database.Bucket
}

func (d dbBucket) Bucket(bucketBytes []byte) model.DBBucket {
	return newDBBucket(d.bucket.Bucket(bucketBytes))
}

func (d dbBucket) Key(suffix []byte) model.DBKey {
	return newDBKey(d.bucket.Key(suffix))
}

func (d dbBucket) Path() []byte {
	return d.bucket.Path()
}

func newDBBucket(bucket *database.Bucket) model.DBBucket {
	return dbBucket{bucket: bucket}
}
