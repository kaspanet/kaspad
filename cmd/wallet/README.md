WALLET
======

## IMPORTANT:

### This software is for TESTING ONLY. Do NOT use it for handling real money.

`wallet` is a simple, no-frills wallet software operated via the command line.\
It is capable of generating wallet key-pairs, printing a wallet's current balance, and sending simple transactions.

## Requirements

Go 1.16 or later.

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
$ cd kaspad/cmd/wallet
$ go install .
```

- Wallet should now be installed in `$(go env GOPATH)/bin`. If you did
  not already add the bin directory to your system path during Go installation,
  you are encouraged to do so now.


Usage
-----

* Create a new wallet key-pair: `wallet create --testnet`
* Print a wallet's current balance:
  `wallet balance --testnet --address=kaspatest:000000000000000000000000000000000000000000`
* Send funds to another wallet:
  `wallet send --testnet --private-key=0000000000000000000000000000000000000000000000000000000000000000 --send-amount=50 --to-address=kaspatest:000000000000000000000000000000000000000000`