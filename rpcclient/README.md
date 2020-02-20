rpcclient
=========

[![ISC License](http://img.shields.io/badge/license-ISC-blue.svg)](https://choosealicense.com/licenses/isc/)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/kaspanet/kaspad/rpcclient)

rpcclient implements a Websocket-enabled Kaspa JSON-RPC client package written
in [Go](http://golang.org/). It provides a robust and easy to use client for
interfacing with a Kaspa RPC server that uses a kaspad compatible
Kaspa JSON-RPC API.

## Status

This package is currently under active development. It is already stable and
the infrastructure is complete. However, there are still several RPCs left to
implement and the API is not stable yet.

## Documentation

* [API Reference](http://godoc.org/github.com/kaspanet/kaspad/rpcclient)
* [Websockets Example](https://github.com/kaspanet/kaspad/blob/master/rpcclient/examples/websockets) 
  Connects to a kaspad RPC server using TLS-secured websockets, registers for
  block connected and block disconnected notifications, and gets the current
  block count
* [HTTP POST Example](https://github.com/kaspanet/kaspad/rpcclient/blob/master/examples/httppost) 
  Connects to a kaspad RPC server using HTTP POST mode with TLS disabled
  and gets the current block count

## Major Features

* Supports Websockets and HTTP POST mode 
* Provides callback and registration functions for kaspad notifications
* Translates to and from higher-level and easier to use Go types
* Offers a synchronous (blocking) and asynchronous API
* When running in Websockets mode (the default):
  * Automatic reconnect handling (can be disabled)
  * Outstanding commands are automatically reissued
  * Registered notifications are automatically reregistered
  * Back-off support on reconnect attempts

