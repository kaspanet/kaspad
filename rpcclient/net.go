// Copyright (c) 2014-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package rpcclient

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"github.com/kaspanet/kaspad/util/copytopointer"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"

	"github.com/kaspanet/kaspad/rpcmodel"
)

// FutureAddNodeResult is a future promise to deliver the result of an
// AddNodeAsync RPC invocation (or an applicable error).
type FutureAddNodeResult chan *response

// Receive waits for the response promised by the future and returns an error if
// any occurred when performing the specified command.
func (r FutureAddNodeResult) Receive() error {
	_, err := receiveFuture(r)
	return err
}

// AddManualNodeAsync returns an instance of a type that can be used to get the result
// of the RPC at some future time by invoking the Receive function on the
// returned instance.
//
// See AddNode for the blocking version and more details.
func (c *Client) AddManualNodeAsync(host string) FutureAddNodeResult {
	cmd := rpcmodel.NewAddManualNodeCmd(host, copytopointer.Bool(false))
	return c.sendCmd(cmd)
}

// AddManualNode attempts to perform the passed command on the passed persistent peer.
// For example, it can be used to add or a remove a persistent peer, or to do
// a one time connection to a peer.
//
// It may not be used to remove non-persistent peers.
func (c *Client) AddManualNode(host string) error {
	return c.AddManualNodeAsync(host).Receive()
}

// FutureGetManualNodeInfoResult is a future promise to deliver the result of a
// GetManualNodeInfoAsync RPC invocation (or an applicable error).
type FutureGetManualNodeInfoResult chan *response

// Receive waits for the response promised by the future and returns information
// about manually added (persistent) peers.
func (r FutureGetManualNodeInfoResult) Receive() ([]rpcmodel.GetManualNodeInfoResult, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal as an array of getmanualnodeinfo result objects.
	var nodeInfo []rpcmodel.GetManualNodeInfoResult
	err = json.Unmarshal(res, &nodeInfo)
	if err != nil {
		return nil, err
	}

	return nodeInfo, nil
}

// GetManualNodeInfoAsync returns an instance of a type that can be used to get
// the result of the RPC at some future time by invoking the Receive function on
// the returned instance.
//
// See GetManualNodeInfo for the blocking version and more details.
func (c *Client) GetManualNodeInfoAsync(peer string) FutureGetManualNodeInfoResult {
	cmd := rpcmodel.NewGetManualNodeInfoCmd(peer, nil)
	return c.sendCmd(cmd)
}

// GetManualNodeInfo returns information about manually added (persistent) peers.
//
// See GetManualNodeInfoNoDNS to retrieve only a list of the added (persistent)
// peers.
func (c *Client) GetManualNodeInfo(peer string) ([]rpcmodel.GetManualNodeInfoResult, error) {
	return c.GetManualNodeInfoAsync(peer).Receive()
}

// FutureGetManualNodeInfoNoDNSResult is a future promise to deliver the result
// of a GetManualNodeInfoNoDNSAsync RPC invocation (or an applicable error).
type FutureGetManualNodeInfoNoDNSResult chan *response

// Receive waits for the response promised by the future and returns a list of
// manually added (persistent) peers.
func (r FutureGetManualNodeInfoNoDNSResult) Receive() ([]string, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as an array of strings.
	var nodes []string
	err = json.Unmarshal(res, &nodes)
	if err != nil {
		return nil, err
	}

	return nodes, nil
}

// GetManualNodeInfoNoDNSAsync returns an instance of a type that can be used to
// get the result of the RPC at some future time by invoking the Receive
// function on the returned instance.
//
// See GetManualNodeInfoNoDNS for the blocking version and more details.
func (c *Client) GetManualNodeInfoNoDNSAsync(peer string) FutureGetManualNodeInfoNoDNSResult {
	cmd := rpcmodel.NewGetManualNodeInfoCmd(peer, copytopointer.Bool(false))
	return c.sendCmd(cmd)
}

// GetManualNodeInfoNoDNS returns a list of manually added (persistent) peers.
// This works by setting the dns flag to false in the underlying RPC.
//
// See GetManualNodeInfo to obtain more information about each added (persistent)
// peer.
func (c *Client) GetManualNodeInfoNoDNS(peer string) ([]string, error) {
	return c.GetManualNodeInfoNoDNSAsync(peer).Receive()
}

// FutureGetConnectionCountResult is a future promise to deliver the result
// of a GetConnectionCountAsync RPC invocation (or an applicable error).
type FutureGetConnectionCountResult chan *response

// Receive waits for the response promised by the future and returns the number
// of active connections to other peers.
func (r FutureGetConnectionCountResult) Receive() (int64, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return 0, err
	}

	// Unmarshal result as an int64.
	var count int64
	err = json.Unmarshal(res, &count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetConnectionCountAsync returns an instance of a type that can be used to get
// the result of the RPC at some future time by invoking the Receive function on
// the returned instance.
//
// See GetConnectionCount for the blocking version and more details.
func (c *Client) GetConnectionCountAsync() FutureGetConnectionCountResult {
	cmd := rpcmodel.NewGetConnectionCountCmd()
	return c.sendCmd(cmd)
}

// GetConnectionCount returns the number of active connections to other peers.
func (c *Client) GetConnectionCount() (int64, error) {
	return c.GetConnectionCountAsync().Receive()
}

// FuturePingResult is a future promise to deliver the result of a PingAsync RPC
// invocation (or an applicable error).
type FuturePingResult chan *response

// Receive waits for the response promised by the future and returns the result
// of queueing a ping to be sent to each connected peer.
func (r FuturePingResult) Receive() error {
	_, err := receiveFuture(r)
	return err
}

// PingAsync returns an instance of a type that can be used to get the result of
// the RPC at some future time by invoking the Receive function on the returned
// instance.
//
// See Ping for the blocking version and more details.
func (c *Client) PingAsync() FuturePingResult {
	cmd := rpcmodel.NewPingCmd()
	return c.sendCmd(cmd)
}

// Ping queues a ping to be sent to each connected peer.
//
// Use the GetPeerInfo function and examine the PingTime and PingWait fields to
// access the ping times.
func (c *Client) Ping() error {
	return c.PingAsync().Receive()
}

// FutureGetPeerInfoResult is a future promise to deliver the result of a
// GetPeerInfoAsync RPC invocation (or an applicable error).
type FutureGetPeerInfoResult chan *response

// Receive waits for the response promised by the future and returns  data about
// each connected network peer.
func (r FutureGetPeerInfoResult) Receive() ([]rpcmodel.GetPeerInfoResult, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as an array of getpeerinfo result objects.
	var peerInfo []rpcmodel.GetPeerInfoResult
	err = json.Unmarshal(res, &peerInfo)
	if err != nil {
		return nil, err
	}

	return peerInfo, nil
}

// GetPeerInfoAsync returns an instance of a type that can be used to get the
// result of the RPC at some future time by invoking the Receive function on the
// returned instance.
//
// See GetPeerInfo for the blocking version and more details.
func (c *Client) GetPeerInfoAsync() FutureGetPeerInfoResult {
	cmd := rpcmodel.NewGetPeerInfoCmd()
	return c.sendCmd(cmd)
}

// GetPeerInfo returns data about each connected network peer.
func (c *Client) GetPeerInfo() ([]rpcmodel.GetPeerInfoResult, error) {
	return c.GetPeerInfoAsync().Receive()
}

// FutureGetNetTotalsResult is a future promise to deliver the result of a
// GetNetTotalsAsync RPC invocation (or an applicable error).
type FutureGetNetTotalsResult chan *response

// Receive waits for the response promised by the future and returns network
// traffic statistics.
func (r FutureGetNetTotalsResult) Receive() (*rpcmodel.GetNetTotalsResult, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as a getnettotals result object.
	var totals rpcmodel.GetNetTotalsResult
	err = json.Unmarshal(res, &totals)
	if err != nil {
		return nil, err
	}

	return &totals, nil
}

// GetNetTotalsAsync returns an instance of a type that can be used to get the
// result of the RPC at some future time by invoking the Receive function on the
// returned instance.
//
// See GetNetTotals for the blocking version and more details.
func (c *Client) GetNetTotalsAsync() FutureGetNetTotalsResult {
	cmd := rpcmodel.NewGetNetTotalsCmd()
	return c.sendCmd(cmd)
}

// GetNetTotals returns network traffic statistics.
func (c *Client) GetNetTotals() (*rpcmodel.GetNetTotalsResult, error) {
	return c.GetNetTotalsAsync().Receive()
}

// FutureDebugLevelResult is a future promise to deliver the result of a
// DebugLevelAsync RPC invocation (or an applicable error).
type FutureDebugLevelResult chan *response

// Receive waits for the response promised by the future and returns the result
// of setting the debug logging level to the passed level specification or the
// list of of the available subsystems for the special keyword 'show'.
func (r FutureDebugLevelResult) Receive() (string, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return "", err
	}

	// Unmashal the result as a string.
	var result string
	err = json.Unmarshal(res, &result)
	if err != nil {
		return "", err
	}
	return result, nil
}

// DebugLevelAsync returns an instance of a type that can be used to get the
// result of the RPC at some future time by invoking the Receive function on
// the returned instance.
//
// See DebugLevel for the blocking version and more details.
func (c *Client) DebugLevelAsync(levelSpec string) FutureDebugLevelResult {
	cmd := rpcmodel.NewDebugLevelCmd(levelSpec)
	return c.sendCmd(cmd)
}

// DebugLevel dynamically sets the debug logging level to the passed level
// specification.
//
// The levelspec can be either a debug level or of the form:
// 	<subsystem>=<level>,<subsystem2>=<level2>,...
//
// Additionally, the special keyword 'show' can be used to get a list of the
// available subsystems.
func (c *Client) DebugLevel(levelSpec string) (string, error) {
	return c.DebugLevelAsync(levelSpec).Receive()
}

// FutureGetSelectedTipResult is a future promise to deliver the result of a
// GetSelectedTipAsync RPC invocation (or an applicable error).
type FutureGetSelectedTipResult chan *response

// Receive waits for the response promised by the future and returns the
// selected tip block.
func (r FutureGetSelectedTipResult) Receive() (*wire.MsgBlock, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as a string.
	var blockHex string
	err = json.Unmarshal(res, &blockHex)
	if err != nil {
		return nil, err
	}

	// Decode the serialized block hex to raw bytes.
	serializedBlock, err := hex.DecodeString(blockHex)
	if err != nil {
		return nil, err
	}

	// Deserialize the block and return it.
	var msgBlock wire.MsgBlock
	err = msgBlock.Deserialize(bytes.NewReader(serializedBlock))
	if err != nil {
		return nil, err
	}
	return &msgBlock, nil
}

// GetSelectedTipAsync returns an instance of a type that can be used to get the
// result of the RPC at some future time by invoking the Receive function on the
// returned instance.
//
// See GetSelectedTip for the blocking version and more details.
func (c *Client) GetSelectedTipAsync() FutureGetSelectedTipResult {
	cmd := rpcmodel.NewGetSelectedTipCmd(copytopointer.Bool(false), copytopointer.Bool(false))
	return c.sendCmd(cmd)
}

// GetSelectedTip returns the block of the selected DAG tip
func (c *Client) GetSelectedTip() (*rpcmodel.GetBlockVerboseResult, error) {
	return c.GetSelectedTipVerboseAsync().Receive()
}

// FutureGetSelectedTipVerboseResult is a future promise to deliver the result of a
// GetSelectedTipVerboseAsync RPC invocation (or an applicable error).
type FutureGetSelectedTipVerboseResult chan *response

// Receive waits for the response promised by the future and returns the data
// structure from the server with information about the requested block.
func (r FutureGetSelectedTipVerboseResult) Receive() (*rpcmodel.GetBlockVerboseResult, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal the raw result into a BlockResult.
	var blockResult rpcmodel.GetBlockVerboseResult
	err = json.Unmarshal(res, &blockResult)
	if err != nil {
		return nil, err
	}
	return &blockResult, nil
}

// GetSelectedTipVerboseAsync returns an instance of a type that can be used to get
// the result of the RPC at some future time by invoking the Receive function on
// the returned instance.
//
// See GeSelectedTipBlockVerbose for the blocking version and more details.
func (c *Client) GetSelectedTipVerboseAsync() FutureGetSelectedTipVerboseResult {
	cmd := rpcmodel.NewGetSelectedTipCmd(copytopointer.Bool(true), copytopointer.Bool(false))
	return c.sendCmd(cmd)
}

// FutureGetCurrentNetResult is a future promise to deliver the result of a
// GetCurrentNetAsync RPC invocation (or an applicable error).
type FutureGetCurrentNetResult chan *response

// Receive waits for the response promised by the future and returns the network
// the server is running on.
func (r FutureGetCurrentNetResult) Receive() (wire.KaspaNet, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return 0, err
	}

	// Unmarshal result as an int64.
	var net int64
	err = json.Unmarshal(res, &net)
	if err != nil {
		return 0, err
	}

	return wire.KaspaNet(net), nil
}

// GetCurrentNetAsync returns an instance of a type that can be used to get the
// result of the RPC at some future time by invoking the Receive function on the
// returned instance.
//
// See GetCurrentNet for the blocking version and more details.
func (c *Client) GetCurrentNetAsync() FutureGetCurrentNetResult {
	cmd := rpcmodel.NewGetCurrentNetCmd()
	return c.sendCmd(cmd)
}

// GetCurrentNet returns the network the server is running on.
func (c *Client) GetCurrentNet() (wire.KaspaNet, error) {
	return c.GetCurrentNetAsync().Receive()
}

// FutureGetHeadersResult is a future promise to deliver the result of a
// getheaders RPC invocation (or an applicable error).
type FutureGetHeadersResult chan *response

// Receive waits for the response promised by the future and returns the
// getheaders result.
func (r FutureGetHeadersResult) Receive() ([]wire.BlockHeader, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as a slice of strings.
	var result []string
	err = json.Unmarshal(res, &result)
	if err != nil {
		return nil, err
	}

	// Deserialize the []string into []wire.BlockHeader.
	headers := make([]wire.BlockHeader, len(result))
	for i, headerHex := range result {
		serialized, err := hex.DecodeString(headerHex)
		if err != nil {
			return nil, err
		}
		err = headers[i].Deserialize(bytes.NewReader(serialized))
		if err != nil {
			return nil, err
		}
	}
	return headers, nil
}

// GetTopHeadersAsync returns an instance of a type that can be used to get the result
// of the RPC at some future time by invoking the Receive function on the returned instance.
//
// See GetTopHeaders for the blocking version and more details.
func (c *Client) GetTopHeadersAsync(highHash *daghash.Hash) FutureGetHeadersResult {
	var hash *string
	if highHash != nil {
		hash = copytopointer.String(highHash.String())
	}
	cmd := rpcmodel.NewGetTopHeadersCmd(hash)
	return c.sendCmd(cmd)
}

// GetTopHeaders sends a getTopHeaders rpc command to the server.
func (c *Client) GetTopHeaders(highHash *daghash.Hash) ([]wire.BlockHeader, error) {
	return c.GetTopHeadersAsync(highHash).Receive()
}

// GetHeadersAsync returns an instance of a type that can be used to get the result
// of the RPC at some future time by invoking the Receive function on the returned instance.
//
// See GetHeaders for the blocking version and more details.
func (c *Client) GetHeadersAsync(lowHash, highHash *daghash.Hash) FutureGetHeadersResult {
	lowHashStr := ""
	if lowHash != nil {
		lowHashStr = lowHash.String()
	}
	highHashStr := ""
	if highHash != nil {
		highHashStr = highHash.String()
	}
	cmd := rpcmodel.NewGetHeadersCmd(lowHashStr, highHashStr)
	return c.sendCmd(cmd)
}

// GetHeaders mimics the wire protocol getheaders and headers messages by
// returning all headers in the DAG after the first known block in the
// locators, up until a block hash matches highHash.
func (c *Client) GetHeaders(lowHash, highHash *daghash.Hash) ([]wire.BlockHeader, error) {
	return c.GetHeadersAsync(lowHash, highHash).Receive()
}

// FutureSessionResult is a future promise to deliver the result of a
// SessionAsync RPC invocation (or an applicable error).
type FutureSessionResult chan *response

// Receive waits for the response promised by the future and returns the
// session result.
func (r FutureSessionResult) Receive() (*rpcmodel.SessionResult, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as a session result object.
	var session rpcmodel.SessionResult
	err = json.Unmarshal(res, &session)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

// SessionAsync returns an instance of a type that can be used to get the result
// of the RPC at some future time by invoking the Receive function on the
// returned instance.
//
// See Session for the blocking version and more details.
func (c *Client) SessionAsync() FutureSessionResult {
	// Not supported in HTTP POST mode.
	if c.config.HTTPPostMode {
		return newFutureError(ErrWebsocketsRequired)
	}

	cmd := rpcmodel.NewSessionCmd()
	return c.sendCmd(cmd)
}

// Session returns details regarding a websocket client's current connection.
//
// This RPC requires the client to be running in websocket mode.
func (c *Client) Session() (*rpcmodel.SessionResult, error) {
	return c.SessionAsync().Receive()
}

// FutureVersionResult is a future promise to deliver the result of a version
// RPC invocation (or an applicable error).
type FutureVersionResult chan *response

// Receive waits for the response promised by the future and returns the version
// result.
func (r FutureVersionResult) Receive() (map[string]rpcmodel.VersionResult,
	error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as a version result object.
	var vr map[string]rpcmodel.VersionResult
	err = json.Unmarshal(res, &vr)
	if err != nil {
		return nil, err
	}

	return vr, nil
}

// VersionAsync returns an instance of a type that can be used to get the result
// of the RPC at some future time by invoking the Receive function on the
// returned instance.
//
// See Version for the blocking version and more details.
func (c *Client) VersionAsync() FutureVersionResult {
	cmd := rpcmodel.NewVersionCmd()
	return c.sendCmd(cmd)
}

// Version returns information about the server's JSON-RPC API versions.
func (c *Client) Version() (map[string]rpcmodel.VersionResult, error) {
	return c.VersionAsync().Receive()
}
