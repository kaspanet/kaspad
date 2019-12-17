// Copyright (c) 2014-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

/*
Package rpcclient implements a websocket-enabled kaspa JSON-RPC client.

Overview

This client provides a robust and easy to use client for interfacing with a
kaspa RPC server that uses a kaspa compatible kaspa JSON-RPC
API.

In addition to the compatible standard HTTP POST JSON-RPC API, kaspad
provides a websocket interface that is more efficient than the standard
HTTP POST method of accessing RPC. The section below discusses the differences
between HTTP POST and websockets.

By default, this client assumes the RPC server supports websockets and has
TLS enabled.

Websockets vs HTTP POST

In HTTP POST-based JSON-RPC, every request creates a new HTTP connection,
issues the call, waits for the response, and closes the connection. This adds
quite a bit of overhead to every call and lacks flexibility for features such as
notifications.

In contrast, the websocket-based JSON-RPC interface provided by kaspad
only uses a single connection that remains open and allows
asynchronous bi-directional communication.

The websocket interface supports all of the same commands as HTTP POST, but they
can be invoked without having to go through a connect/disconnect cycle for every
call. In addition, the websocket interface provides other nice features such as
the ability to register for asynchronous notifications of various events.

Synchronous vs Asynchronous API

The client provides both a synchronous (blocking) and asynchronous API.

The synchronous (blocking) API is typically sufficient for most use cases. It
works by issuing the RPC and blocking until the response is received. This
allows  straightforward code where you have the response as soon as the function
returns.

The asynchronous API works on the concept of futures. When you invoke the async
version of a command, it will quickly return an instance of a type that promises
to provide the result of the RPC at some future time. In the background, the
RPC call is issued and the result is stored in the returned instance. Invoking
the Receive method on the returned instance will either return the result
immediately if it has already arrived, or block until it has. This is useful
since it provides the caller with greater control over concurrency.

Notifications

The first important part of notifications is to realize that they will only
work when connected via websockets. This should intuitively make sense
because HTTP POST mode does not keep a connection open!

All notifications provided by kaspad require registration to opt-in. For example,
if you want to be notified when funds are received by a set of addresses, you
register the addresses via the NotifyReceived (or NotifyReceivedAsync) function.

Notification Handlers

Notifications are exposed by the client through the use of callback handlers
which are setup via a NotificationHandlers instance that is specified by the
caller when creating the client.

It is important that these notification handlers complete quickly since they
are intentionally in the main read loop and will block further reads until
they complete. This provides the caller with the flexibility to decide what to
do when notifications are coming in faster than they are being handled.

In particular this means issuing a blocking RPC call from a callback handler
will cause a deadlock as more server responses won't be read until the callback
returns, but the callback would be waiting for a response. Thus, any
additional RPCs must be issued an a completely decoupled manner.

Automatic Reconnection

By default, when running in websockets mode, this client will automatically
keep trying to reconnect to the RPC server should the connection be lost. There
is a back-off in between each connection attempt until it reaches one try per
minute. Once a connection is re-established, all previously registered
notifications are automatically re-registered and any in-flight commands are
re-issued. This means from the caller's perspective, the request simply takes
longer to complete.

The caller may invoke the Shutdown method on the client to force the client
to cease reconnect attempts and return ErrClientShutdown for all outstanding
commands.

The automatic reconnection can be disabled by setting the DisableAutoReconnect
flag to true in the connection config when creating the client.

Errors

There are 3 categories of errors that will be returned throughout this package:

  - Errors related to the client connection such as authentication, endpoint,
    disconnect, and shutdown
  - Errors that occur before communicating with the remote RPC server such as
    command creation and marshaling errors or issues talking to the remote
    server
  - Errors returned from the remote RPC server like unimplemented commands,
    nonexistent requested blocks and transactions, malformed data, and incorrect
    networks

The first category of errors are typically one of ErrInvalidAuth,
ErrInvalidEndpoint, ErrClientDisconnect, or ErrClientShutdown.

NOTE: The ErrClientDisconnect will not be returned unless the
DisableAutoReconnect flag is set since the client automatically handles
reconnect by default as previously described.

The second category of errors typically indicates a programmer error and as such
the type can vary, but usually will be best handled by simply showing/logging
it.

The third category of errors, that is errors returned by the server, can be
detected by type asserting the error in a *rpcmodel.RPCError. For example, to
detect if a command is unimplemented by the remote RPC server:

  amount, err := client.GetBalance("")
  if err != nil {
  	if jerr, ok := err.(*rpcmodel.RPCError); ok {
  		switch jerr.Code {
  		case rpcmodel.ErrRPCUnimplemented:
  			// Handle not implemented error

  		// Handle other specific errors you care about
		}
  	}

  	// Log or otherwise handle the error knowing it was not one returned
  	// from the remote RPC server.
  }

Example Usage

The following full-blown client examples are in the examples directory:

 - httppost
   Connects to a kaspa RPC server using HTTP POST mode with TLS disabled
   and gets the current block count
 - websockets
   Connects to a kaspad RPC server using TLS-secured websockets, registers for
   block connected and block disconnected notifications, and gets the current
   block count
*/
package rpcclient
