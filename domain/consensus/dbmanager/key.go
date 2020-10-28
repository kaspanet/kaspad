package dbmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
)

func dbKeyToDatabaseKey(key model.DBKey) *database.Key {
	return dbBucketToDatabaseBucket(key.Bucket()).Key(key.Suffix())
}

type dbKey struct {
	key *database.Key
}

func (d dbKey) Bytes() []byte {
	return d.key.Bytes()
}

func (d dbKey) Bucket() model.DBBucket {
	return newDBBucket(d.key.Bucket())
}

func (d dbKey) Suffix() []byte {
	return d.key.Suffix()
}

func newDBKey(key *database.Key) model.DBKey {
	return &dbKey{key: key}
}
