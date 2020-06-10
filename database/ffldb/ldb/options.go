package ldb

import "github.com/syndtr/goleveldb/leveldb/opt"

var (
	defaultOptions = opt.Options{
		Compression:          opt.NoCompression,
		BlockCacheCapacity:   512 * opt.MiB,
		WriteBuffer:          512 * opt.MiB,
		IteratorSamplingRate: 512 * opt.MiB,
	}

	// Options is a function that returns a leveldb
	// opt.Options struct for opening a database.
	// It's defined as a variable for the sake of testing.
	Options = func() *opt.Options {
		return &defaultOptions
	}
)
