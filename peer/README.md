peer
====

[![ISC License](http://img.shields.io/badge/license-ISC-blue.svg)](https://choosealicense.com/licenses/isc/)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/kaspanet/kaspad/peer)

Package peer provides a common base for creating and managing kaspa network
peers.

## Overview

This package builds upon the wire package, which provides the fundamental
primitives necessary to speak the kaspa wire protocol, in order to simplify
the process of creating fully functional peers.

A quick overview of the major features peer provides are as follows:

 - Provides a basic concurrent safe kaspa peer for handling kaspa
   communications via the peer-to-peer protocol
 - Full duplex reading and writing of kaspa protocol messages
 - Automatic handling of the initial handshake process including protocol
   version negotiation
 - Asynchronous message queueing of outbound messages with optional channel for
   notification when the message is actually sent
 - Flexible peer configuration
   - Caller is responsible for creating outgoing connections and listening for
     incoming connections so they have flexibility to establish connections as
     they see fit (proxies, etc)
   - User agent name and version
   - Maximum supported protocol version
   - Ability to register callbacks for handling kaspa protocol messages
 - Inventory message batching and send trickling with known inventory detection
   and avoidance
 - Automatic periodic keep-alive pinging and pong responses
 - Random nonce generation and self connection detection
 - Proper handling of bloom filter related commands when the caller does not
   specify the related flag to signal support
   - Disconnects the peer when the protocol version is high enough
   - Does not invoke the related callbacks for older protocol versions
 - Snapshottable peer statistics such as the total number of bytes read and
   written, the remote address, user agent, and negotiated protocol version
 - Helper functions pushing addresses, getblockinvs, getheaders, and reject
   messages
   - These could all be sent manually via the standard message output function,
     but the helpers provide additional nice functionality such as duplicate
     filtering and address randomization
 - Ability to wait for shutdown/disconnect
 - Comprehensive test coverage

