# c4exminer

C4exminer is a CPU-based miner for c4exd

## Requirements

Go 1.19 or later.

## Installation

#### Build from Source

- Install Go according to the installation instructions here:
  http://golang.org/doc/install

- Ensure Go was installed properly and is a supported version:

```bash
$ go version
```

- Run the following commands to obtain and install c4exd including all dependencies:

```bash
$ git clone https://github.com/c4ei/c4exd
$ cd c4exd/cmd/c4exminer
$ go install .
```

- Kapaminer should now be installed in `$(go env GOPATH)/bin`. If you did
  not already add the bin directory to your system path during Go installation,
  you are encouraged to do so now.
  
## Usage

The full c4exminer configuration options can be seen with:

```bash
$ c4exminer --help
```

But the minimum configuration needed to run it is:
```bash
$ c4exminer --miningaddr=<YOUR_MINING_ADDRESS>
```