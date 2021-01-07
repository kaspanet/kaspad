
Kaspad
====
Warning: This is pre-alpha software. There's no guarantee anything works.
====

[![ISC License](http://img.shields.io/badge/license-ISC-blue.svg)](https://choosealicense.com/licenses/isc/)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/kaspanet/kaspad)

Kaspad is the reference full node Kaspa implementation written in Go (golang).

This project is currently under active development and is in a pre-Alpha state.
Some things still don't work and APIs are far from finalized. The code is provided for reference only.

## What is kaspa

Kaspa is an attempt at a proof-of-work cryptocurrency that operates in internet speed, with subsecond block times, based on a generalization of Nakamoto Consensus, the PHANTOM protocol.

## Requirements

Go 1.14 or later.

## Installation

#### Build from Source

- Install Go according to the installation instructions here:
  http://golang.org/doc/install

- Ensure Go was installed properly and is a supported version:

```bash
$ go version
```

- Run the following commands to obtain and install kaspad including all dependencies:

```bash
$ git clone https://github.com/kaspanet/kaspad
$ cd kaspad
$ go install . ./cmd/...
```

- Kaspad (and utilities) should now be installed in `$(go env GOPATH)/bin`. If you did
  not already add the bin directory to your system path during Go installation,
  you are encouraged to do so now.


## Getting Started

Kaspad has several configuration options available to tweak how it runs, but all
of the basic operations work with zero configuration.

```bash
$ kaspad
```

## Discord
Join our discord server using the following link: https://discord.gg/WmGhhzk

## Issue Tracker

The [integrated github issue tracker](https://github.com/kaspanet/kaspad/issues)
is used for this project.

## Documentation

The documentation is a work-in-progress. It is located in the [docs](https://github.com/kaspanet/kaspad/tree/master/docs) folder.

## License

Kaspad is licensed under the copyfree [ISC License](https://choosealicense.com/licenses/isc/).
