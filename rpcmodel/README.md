rpcmodel
=======

[![ISC License](http://img.shields.io/badge/license-ISC-blue.svg)](http://copyfree.org)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/kaspanet/kaspad/rpcmodel)

Package rpcmodel implements concrete types for marshalling to and from the
kaspa JSON-RPC API. A comprehensive suite of tests is provided to ensure
proper functionality.

Note that although it's possible to use this package directly to implement an
RPC client, it is not recommended since it is only intended as an infrastructure
package. Instead, RPC clients should use the rpcclient package which provides
a full blown RPC client with many features such as automatic connection
management, websocket support, automatic notification re-registration on
reconnect, and conversion from the raw underlying RPC types (strings, floats,
ints, etc) to higher-level types with many nice and useful properties.

## Examples

* [Marshal Command](http://godoc.org/github.com/kaspanet/kaspad/rpcmodel#example-MarshalCmd) 
  Demonstrates how to create and marshal a command into a JSON-RPC request.

* [Unmarshal Command](http://godoc.org/github.com/kaspanet/kaspad/rpcmodel#example-UnmarshalCmd) 
  Demonstrates how to unmarshal a JSON-RPC request and then unmarshal the
  concrete request into a concrete command.

* [Marshal Response](http://godoc.org/github.com/kaspanet/kaspad/rpcmodel#example-MarshalResponse) 
  Demonstrates how to marshal a JSON-RPC response.

* [Unmarshal Response](http://godoc.org/github.com/kaspanet/kaspad/rpcmodel#example-package--UnmarshalResponse) 
  Demonstrates how to unmarshal a JSON-RPC response and then unmarshal the
  result field in the response to a concrete type.

