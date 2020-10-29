package database

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
)

func dbBucketToDatabaseBucket(bucket model.DBBucket) *database.Bucket {
	return database.MakeBucket(bucket.Path())
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
	return &dbBucket{bucket: bucket}
}
