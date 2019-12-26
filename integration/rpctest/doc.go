/*
Package rpctest provides a kaspad-specific RPC testing harness crafting and
executing integration tests by driving a `kaspad` instance via the `RPC`
interface. Each instance of an active harness comes equipped with a simple
in-memory HD wallet capable of properly syncing to the generated chain,
creating new addresses, and crafting fully signed transactions paying to an
arbitrary set of outputs.
*/
package rpctest
