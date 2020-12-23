WALLET
======

## IMPORTANT:

### This software is for TESTING ONLY. Do NOT use it for handling real money.

`wallet` is a simple, no-frills wallet software operated via the command line.\
It is capable of generating wallet key-pairs, printing a wallet's current balance, and sending simple transactions.


Usage
-----

* Create a new wallet key-pair: `wallet create --testnet`
* Print a wallet's current balance:
  `wallet balance --testnet --address=kaspatest:000000000000000000000000000000000000000000`
* Send funds to another wallet:
  `wallet send --testnet --private-key=0000000000000000000000000000000000000000000000000000000000000000 --send-amount=50 --to-address=kaspatest:000000000000000000000000000000000000000000`