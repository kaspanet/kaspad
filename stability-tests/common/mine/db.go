package mine

import (
	"path/filepath"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/kaspanet/kaspad/infrastructure/db/database/ldb"
)

const leveldbCacheSizeMiB = 256

var blockIDToHashBucket = database.MakeBucket([]byte("id-to-block-hash"))
var lastMinedBlockKey = database.MakeBucket(nil).Key([]byte("last-sent-block"))

type miningDB struct {
	idToBlockHash map[string]*externalapi.DomainHash
	hashToBlockID map[externalapi.DomainHash]string
	db            *ldb.LevelDB
}

func (mdb *miningDB) hashByID(id string) *externalapi.DomainHash {
	return mdb.idToBlockHash[id]
}

func (mdb *miningDB) putID(id string, hash *externalapi.DomainHash) error {
	mdb.idToBlockHash[id] = hash
	mdb.hashToBlockID[*hash] = id
	return mdb.db.Put(blockIDToHashBucket.Key([]byte(id)), hash.ByteSlice())
}

func (mdb *miningDB) updateLastMinedBlock(id string) error {
	return mdb.db.Put(lastMinedBlockKey, []byte(id))
}

func (mdb *miningDB) lastMinedBlock() (string, error) {
	has, err := mdb.db.Has(lastMinedBlockKey)
	if err != nil {
		return "", err
	}

	if !has {
		return "0", nil
	}

	blockID, err := mdb.db.Get(lastMinedBlockKey)
	if err != nil {
		return "", err
	}

	return string(blockID), nil
}

func newMiningDB(dataDir string) (*miningDB, error) {
	idToBlockHash := make(map[string]*externalapi.DomainHash)
	hashToBlockID := make(map[externalapi.DomainHash]string)

	dbPath := filepath.Join(dataDir, "minedb")
	db, err := ldb.NewLevelDB(dbPath, leveldbCacheSizeMiB)
	if err != nil {
		return nil, err
	}

	cursor, err := db.Cursor(blockIDToHashBucket)
	if err != nil {
		return nil, err
	}

	for cursor.Next() {
		key, err := cursor.Key()
		if err != nil {
			return nil, err
		}

		value, err := cursor.Value()
		if err != nil {
			return nil, err
		}

		hash, err := externalapi.NewDomainHashFromByteSlice(value)
		if err != nil {
			return nil, err
		}

		id := string(key.Suffix())
		idToBlockHash[id] = hash
		hashToBlockID[*hash] = id
	}

	return &miningDB{
		idToBlockHash: idToBlockHash,
		hashToBlockID: hashToBlockID,
		db:            db,
	}, nil
}
