package ldb

import "github.com/syndtr/goleveldb/leveldb/opt"

// Options is a function that returns a leveldb
// opt.Options struct for opening a database.
func Options() opt.Options {
	return opt.Options{
		Compression:            opt.NoCompression,
		DisableSeeksCompaction: true,
		NoSync:                 true,
	}
}
