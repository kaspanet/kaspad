// Copyright (c) 2015-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

// Parts of this interface were inspired heavily by the excellent boltdb project
// at https://github.com/boltdb/bolt by Ben B. Johnson.

package database

import (
	"github.com/kaspanet/kaspad/util/daghash"
)

type Cursor interface{}

// Bucket represents a collection of key/value pairs.
type Bucket interface {
	Get(key []byte) []byte
}

// BlockRegion specifies a particular region of a block identified by the
// specified hash, given an offset and length.
type BlockRegion struct {
	Hash   *daghash.Hash
	Offset uint32
	Len    uint32
}

type Tx interface {
	// Metadata returns the top-most bucket for all metadata storage.
	Metadata() Bucket
}

type DB interface {
	View(fn func(tx Tx) error) error

	// Close cleanly shuts down the database and syncs all data. It will
	// block until all database transactions have been finalized (rolled
	// back or committed).
	Close() error

	// FlushCache flushes the db cache to the disk.
	FlushCache() error
}
