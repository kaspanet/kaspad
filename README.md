
Kaspad
====
Warning: This is pre-alpha software. There's no guarantee anything works.
====

[![ISC License](http://img.shields.io/badge/license-ISC-blue.svg)](https://choosealicense.com/licenses/isc/)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/kaspanet/kaspad)

Kaspad is the reference full node Kaspa implementation written in Go (golang).

This project is currently under active development and is in a pre-Alpha state. 
Some things still don't work and APIs are far from finalized. The code is provided for reference only.

## Requirements

Latest version of [Go](http://golang.org) (currently 1.13).

## Installation

#### Build from Source

- Install Go according to the installation instructions here:
  http://golang.org/doc/install

- Ensure Go was installed properly and is a supported version:

```bash
$ go version
$ go env GOROOT GOPATH
```

NOTE: The `GOROOT` and `GOPATH` above must not be the same path. It is
recommended that `GOPATH` is set to a directory in your home directory such as
`~/dev/go` to avoid write permission issues. It is also recommended to add
`$GOPATH/bin` to your `PATH` at this point.

- Run the following commands to obtain and install kaspad including all dependencies:

```bash
$ git clone https://github.com/kaspanet/kaspad $GOPATH/src/github.com/kaspanet/kaspad
$ cd $GOPATH/src/github.com/kaspanet/kaspad
$ go install . ./cmd/...
```

- Kaspad (and utilities) should now be installed in `$GOPATH/bin`. If you did
  not already add the bin directory to your system path during Go installation,
  you are encouraged to do so now.


## Getting Started

Kaspad has several configuration options available to tweak how it runs, but all
of the basic operations work with zero configuration.

#### Linux/BSD/POSIX/Source

```bash
$ ./kaspad
```

## Discord
Join our discord server using the following link: https://discord.gg/WmGhhzk

## Issue Tracker

The [integrated github issue tracker](https://github.com/kaspanet/kaspad/issues)
is used for this project.

## Documentation

The documentation is a work-in-progress.

## License

Kaspad is licensed under the copyfree [ISC License](https://choosealicense.com/licenses/isc/).

