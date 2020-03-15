ffldb
=====

[![ISC License](http://img.shields.io/badge/license-ISC-blue.svg)](https://choosealicense.com/licenses/isc/)
[![GoDoc](https://godoc.org/github.com/kaspanet/kaspad/database/ffldb?status.png)](http://godoc.org/github.com/kaspanet/kaspad/database/ffldb)
=======

Package ffldb implements a driver for the database package that uses leveldb for
the backing metadata and flat files for block storage.

This driver is the recommended driver for use with kaspad. It makes use of leveldb
for the metadata, flat files for block storage, and checksums in key areas to
ensure data integrity.

## Usage

This package is a driver to the database package and provides the database type
of "ffldb". The parameters the Open and Create functions take are the
database path as a string and the block network.

```Go
db, err := database.Open("ffldb", "path/to/database", wire.Mainnet)
if err != nil {
	// Handle error
}
```

```Go
db, err := database.Create("ffldb", "path/to/database", wire.Mainnet)
if err != nil {
	// Handle error
}
```

