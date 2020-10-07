package model

// TxContextProxy is a proxy to a database context
type TxContextProxy interface {
	ContextProxy
	IsTxContextProxy()
}

// TxContextProxy is a proxy to a database context with
// an attached database transaction
type ContextProxy interface {
	Accessor() DataAccessor
}

// DataAccessor defines the common interface by which data gets
// accessed in a generic kaspad database.
type DataAccessor interface {
	Put(key *Key, value []byte) error
	Get(key *Key) ([]byte, error)
	Has(key *Key) (bool, error)
	Delete(key *Key) error
	Cursor(bucket *Bucket) (Cursor, error)
}

// Key is a helper type meant to combine prefix
// and suffix into a single database key.
type Key interface {
	Bytes() []byte
	Bucket() *Bucket
	Suffix() []byte
}

// Bucket is a helper type meant to combine buckets
// and sub-buckets that can be used to create database
// keys and prefix-based cursors.
type Bucket interface {
	Bucket(bucketBytes []byte) *Bucket
	Key(suffix []byte) *Key
	Path() []byte
}

// Cursor iterates over database entries given some bucket
type Cursor interface {
	Next() bool
	First() bool
	Seek(key *Key) error
	Key() (*Key, error)
	Value() ([]byte, error)
	Close() error
}
