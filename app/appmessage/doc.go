/*
Package appmessage implements the c4ex appmessage protocol.

At a high level, this package provides support for marshalling and unmarshalling
supported c4ex messages to and from the appmessage. This package does not deal
with the specifics of message handling such as what to do when a message is
received. This provides the caller with a high level of flexibility.

# C4ex Message Overview

The c4ex protocol consists of exchanging messages between peers. Each
message is preceded by a header which identifies information about it such as
which c4ex network it is a part of, its type, how big it is, and a checksum
to verify validity. All encoding and decoding of message headers is handled by
this package.

To accomplish this, there is a generic interface for c4ex messages named
Message which allows messages of any type to be read, written, or passed around
through channels, functions, etc. In addition, concrete implementations of most
of the currently supported c4ex messages are provided. For these supported
messages, all of the details of marshalling and unmarshalling to and from the
appmessage using c4ex encoding are handled so the caller doesn't have to concern
themselves with the specifics.

# Message Interaction

The following provides a quick summary of how the c4ex messages are intended
to interact with one another. As stated above, these interactions are not
directly handled by this package.

The initial handshake consists of two peers sending each other a version message
(MsgVersion) followed by responding with a verack message (MsgVerAck). Both
peers use the information in the version message (MsgVersion) to negotiate
things such as protocol version and supported services with each other. Once
the initial handshake is complete, the following chart indicates message
interactions in no particular order.

	Peer A Sends                            Peer B Responds
	----------------------------------------------------------------------------
	getaddr message (MsgRequestAddresses)       addr message (MsgAddresses)
	getblockinvs message (MsgGetBlockInvs)  inv message (MsgInv)
	inv message (MsgInv)                    getdata message (MsgGetData)
	getdata message (MsgGetData)            block message (MsgBlock) -or-
	                                        tx message (MsgTx) -or-
	                                        notfound message (MsgNotFound)
	ping message (MsgPing)                  pong message (MsgPong)

# Common Parameters

There are several common parameters that arise when using this package to read
and write c4ex messages. The following sections provide a quick overview of
these parameters so the next sections can build on them.

# Protocol Version

The protocol version should be negotiated with the remote peer at a higher
level than this package via the version (MsgVersion) message exchange, however,
this package provides the appmessage.ProtocolVersion constant which indicates the
latest protocol version this package supports and is typically the value to use
for all outbound connections before a potentially lower protocol version is
negotiated.

# C4ex Network

The c4ex network is a magic number which is used to identify the start of a
message and which c4ex network the message applies to. This package provides
the following constants:

	appmessage.Mainnet
	appmessage.Testnet (Test network)
	appmessage.Simnet  (Simulation test network)
	appmessage.Devnet  (Development network)

# Determining Message Type

As discussed in the c4ex message overview section, this package reads
and writes c4ex messages using a generic interface named Message. In
order to determine the actual concrete type of the message, use a type
switch or type assertion. An example of a type switch follows:

	// Assumes msg is already a valid concrete message such as one created
	// via NewMsgVersion or read via ReadMessage.
	switch msg := msg.(type) {
	case *appmessage.MsgVersion:
		// The message is a pointer to a MsgVersion struct.
		fmt.Printf("Protocol version: %d", msg.ProtocolVersion)
	case *appmessage.MsgBlock:
		// The message is a pointer to a MsgBlock struct.
		fmt.Printf("Number of tx in block: %d", msg.Header.TxnCount)
	}

# Reading Messages

In order to unmarshall c4ex messages from the appmessage, use the ReadMessage
function. It accepts any io.Reader, but typically this will be a net.Conn to
a remote node running a c4ex peer. Example syntax is:

	// Reads and validates the next c4ex message from conn using the
	// protocol version pver and the c4ex network c4exNet. The returns
	// are a appmessage.Message, a []byte which contains the unmarshalled
	// raw payload, and a possible error.
	msg, rawPayload, err := appmessage.ReadMessage(conn, pver, c4exNet)
	if err != nil {
		// Log and handle the error
	}

# Writing Messages

In order to marshall c4ex messages to the appmessage, use the WriteMessage
function. It accepts any io.Writer, but typically this will be a net.Conn to
a remote node running a c4ex peer. Example syntax to request addresses
from a remote peer is:

	// Create a new getaddr c4ex message.
	msg := appmessage.NewMsgRequestAddresses()

	// Writes a c4ex message msg to conn using the protocol version
	// pver, and the c4ex network c4exNet. The return is a possible
	// error.
	err := appmessage.WriteMessage(conn, msg, pver, c4exNet)
	if err != nil {
		// Log and handle the error
	}

# Errors

Errors returned by this package are either the raw errors provided by underlying
calls to read/write from streams such as io.EOF, io.ErrUnexpectedEOF, and
io.ErrShortWrite, or of type appmessage.MessageError. This allows the caller to
differentiate between general IO errors and malformed messages through type
assertions.
*/
package appmessage
