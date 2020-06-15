package ldb

import "github.com/syndtr/goleveldb/leveldb/opt"

var (
	defaultOptions = opt.Options{
		Compression:            opt.NoCompression,
		BlockCacheCapacity:     256 * opt.MiB,
		WriteBuffer:            128 * opt.MiB,
		DisableSeeksCompaction: true,
	}

	// Options is a function that returns a leveldb
	// opt.Options struct for opening a database.
	// It's defined as a variable for the sake of testing.
	Options = func() *opt.Options {
		return &defaultOptions
	}
)
