/*
Package dagconfig defines DAG configuration parameters.

In addition to the main Kaspa network, which is intended for the transfer
of monetary value, there also exists the following standard networks:
  * testnet
  * simnet
  * devnet
These networks are incompatible with each other (each sharing a different
genesis block) and software should handle errors where input intended for
one network is used on an application instance running on a different
network.

For library packages, dagconfig provides the ability to lookup DAG
parameters and encoding magics when passed a *Params.

For main packages, a (typically global) var may be assigned the address of
one of the standard Param vars for use as the application's "active" network.
When a network parameter is needed, it may then be looked up through this
variable (either directly, or hidden in a library call).

 package main

 import (
 	"flag"
 	"fmt"
 	"log"

 	"github.com/kaspanet/kaspad/util"
 	"github.com/kaspanet/kaspad/domain/dagconfig"
 )

 var testnet = flag.Bool("testnet", false, "operate on the testnet Kaspa network")

 // By default (without --testnet), use mainnet.
 var dagParams = &dagconfig.MainnetParams

 func main() {
 	flag.Parse()

 	// Modify active network parameters if operating on testnet.
 	if *testnet {
 		dagParams = &dagconfig.TestnetParams
 	}

 	// later...

 	// Create and print new payment address, specific to the active network.
 	pubKeyHash := make([]byte, 20)
 	addr, err := util.NewAddressPubKeyHash(pubKeyHash, dagParams)
 	if err != nil {
 		log.Fatal(err)
 	}
 	fmt.Println(addr)
 }

If an application does not use one of the standard Kaspa networks, a new
Params struct may be created which defines the parameters for the non-
standard network. As a general rule of thumb, all network parameters
should be unique to the network, but parameter collisions can still occur.
*/
package dagconfig
