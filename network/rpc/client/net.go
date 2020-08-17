// Copyright (c) 2014-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package client

import (
	"bytes"
	"encoding/hex"
	"encoding/json"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/pointers"

	"github.com/kaspanet/kaspad/network/rpc/model"
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

// ConnectNodeAsync returns an instance of a type that can be used to get the result
// of the RPC at some future time by invoking the Receive function on the
// returned instance.
//
// See Connect for the blocking version and more details.
func (c *Client) ConnectNodeAsync(host string) FutureAddNodeResult {
	cmd := model.NewConnectCmd(host, pointers.Bool(false))
	return c.sendCmd(cmd)
}

// ConnectNode attempts to perform the passed command on the passed persistent peer.
// For example, it can be used to add or a remove a persistent peer, or to do
// a one time connection to a peer.
//
// It may not be used to remove non-persistent peers.
func (c *Client) ConnectNode(host string) error {
	return c.ConnectNodeAsync(host).Receive()
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
	cmd := model.NewGetConnectionCountCmd()
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
	cmd := model.NewPingCmd()
	return c.sendCmd(cmd)
}

// Ping queues a ping to be sent to each connected peer.
//
// Use the GetConnectedPeerInfo function and examine the PingTime and PingWait fields to
// access the ping times.
func (c *Client) Ping() error {
	return c.PingAsync().Receive()
}

// FutureGetConnectedPeerInfo is a future promise to deliver the result of a
// GetConnectedPeerInfoAsync RPC invocation (or an applicable error).
type FutureGetConnectedPeerInfo chan *response

// Receive waits for the response promised by the future and returns  data about
// each connected network peer.
func (r FutureGetConnectedPeerInfo) Receive() ([]model.GetConnectedPeerInfoResult, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as an array of getConnectedPeerInfo result objects.
	var peerInfo []model.GetConnectedPeerInfoResult
	err = json.Unmarshal(res, &peerInfo)
	if err != nil {
		return nil, err
	}

	return peerInfo, nil
}

// GetConnectedPeerInfoAsync returns an instance of a type that can be used to get the
// result of the RPC at some future time by invoking the Receive function on the
// returned instance.
//
// See GetConnectedPeerInfo for the blocking version and more details.
func (c *Client) GetConnectedPeerInfoAsync() FutureGetConnectedPeerInfo {
	cmd := model.NewGetConnectedPeerInfoCmd()
	return c.sendCmd(cmd)
}

// GetConnectedPeerInfo returns data about each connected network peer.
func (c *Client) GetConnectedPeerInfo() ([]model.GetConnectedPeerInfoResult, error) {
	return c.GetConnectedPeerInfoAsync().Receive()
}

// FutureGetPeerAddresses is a future promise to deliver the result of a
// GetPeerAddresses RPC invocation (or an applicable error).
type FutureGetPeerAddresses chan *response

// Receive waits for the response promised by the future and returns data about
// peer addresses.
func (r FutureGetPeerAddresses) Receive() (*model.GetPeerAddressesResult, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as an array of getPeerAddresses result objects.
	peerAddresses := &model.GetPeerAddressesResult{}
	err = json.Unmarshal(res, peerAddresses)
	if err != nil {
		return nil, err
	}

	return peerAddresses, nil
}

// GetPeerAddressesAsync returns an instance of a type that can be used to get the
// result of the RPC at some future time by invoking the Receive function on the
// returned instance.
//
// See GetPeerAddresses for the blocking version and more details.
func (c *Client) GetPeerAddressesAsync() FutureGetPeerAddresses {
	cmd := model.NewGetPeerAddressesCmd()
	return c.sendCmd(cmd)
}

// GetPeerAddresses returns data about each connected network peer.
func (c *Client) GetPeerAddresses() (*model.GetPeerAddressesResult, error) {
	return c.GetPeerAddressesAsync().Receive()
}

// FutureGetNetTotalsResult is a future promise to deliver the result of a
// GetNetTotalsAsync RPC invocation (or an applicable error).
type FutureGetNetTotalsResult chan *response

// Receive waits for the response promised by the future and returns network
// traffic statistics.
func (r FutureGetNetTotalsResult) Receive() (*model.GetNetTotalsResult, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as a getnettotals result object.
	var totals model.GetNetTotalsResult
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
	cmd := model.NewGetNetTotalsCmd()
	return c.sendCmd(cmd)
}

// GetNetTotals returns network traffic statistics.
func (c *Client) GetNetTotals() (*model.GetNetTotalsResult, error) {
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
	cmd := model.NewDebugLevelCmd(levelSpec)
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
func (r FutureGetSelectedTipResult) Receive() (*appmessage.MsgBlock, error) {
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
	var msgBlock appmessage.MsgBlock
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
	cmd := model.NewGetSelectedTipCmd(pointers.Bool(false), pointers.Bool(false))
	return c.sendCmd(cmd)
}

// GetSelectedTip returns the block of the selected DAG tip
func (c *Client) GetSelectedTip() (*model.GetBlockVerboseResult, error) {
	return c.GetSelectedTipVerboseAsync().Receive()
}

// FutureGetSelectedTipVerboseResult is a future promise to deliver the result of a
// GetSelectedTipVerboseAsync RPC invocation (or an applicable error).
type FutureGetSelectedTipVerboseResult chan *response

// Receive waits for the response promised by the future and returns the data
// structure from the server with information about the requested block.
func (r FutureGetSelectedTipVerboseResult) Receive() (*model.GetBlockVerboseResult, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal the raw result into a BlockResult.
	var blockResult model.GetBlockVerboseResult
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
	cmd := model.NewGetSelectedTipCmd(pointers.Bool(true), pointers.Bool(false))
	return c.sendCmd(cmd)
}

// FutureGetCurrentNetResult is a future promise to deliver the result of a
// GetCurrentNetAsync RPC invocation (or an applicable error).
type FutureGetCurrentNetResult chan *response

// Receive waits for the response promised by the future and returns the network
// the server is running on.
func (r FutureGetCurrentNetResult) Receive() (appmessage.KaspaNet, error) {
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

	return appmessage.KaspaNet(net), nil
}

// GetCurrentNetAsync returns an instance of a type that can be used to get the
// result of the RPC at some future time by invoking the Receive function on the
// returned instance.
//
// See GetCurrentNet for the blocking version and more details.
func (c *Client) GetCurrentNetAsync() FutureGetCurrentNetResult {
	cmd := model.NewGetCurrentNetCmd()
	return c.sendCmd(cmd)
}

// GetCurrentNet returns the network the server is running on.
func (c *Client) GetCurrentNet() (appmessage.KaspaNet, error) {
	return c.GetCurrentNetAsync().Receive()
}

// FutureGetHeadersResult is a future promise to deliver the result of a
// getheaders RPC invocation (or an applicable error).
type FutureGetHeadersResult chan *response

// Receive waits for the response promised by the future and returns the
// getheaders result.
func (r FutureGetHeadersResult) Receive() ([]appmessage.BlockHeader, error) {
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

	// Deserialize the []string into []appmessage.BlockHeader.
	headers := make([]appmessage.BlockHeader, len(result))
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
		hash = pointers.String(highHash.String())
	}
	cmd := model.NewGetTopHeadersCmd(hash)
	return c.sendCmd(cmd)
}

// GetTopHeaders sends a getTopHeaders rpc command to the server.
func (c *Client) GetTopHeaders(highHash *daghash.Hash) ([]appmessage.BlockHeader, error) {
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
	cmd := model.NewGetHeadersCmd(lowHashStr, highHashStr)
	return c.sendCmd(cmd)
}

// GetHeaders mimics the appmessage protocol getheaders and headers messages by
// returning all headers in the DAG after the first known block in the
// locators, up until a block hash matches highHash.
func (c *Client) GetHeaders(lowHash, highHash *daghash.Hash) ([]appmessage.BlockHeader, error) {
	return c.GetHeadersAsync(lowHash, highHash).Receive()
}

// FutureSessionResult is a future promise to deliver the result of a
// SessionAsync RPC invocation (or an applicable error).
type FutureSessionResult chan *response

// Receive waits for the response promised by the future and returns the
// session result.
func (r FutureSessionResult) Receive() (*model.SessionResult, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as a session result object.
	var session model.SessionResult
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

	cmd := model.NewSessionCmd()
	return c.sendCmd(cmd)
}

// Session returns details regarding a websocket client's current connection.
//
// This RPC requires the client to be running in websocket mode.
func (c *Client) Session() (*model.SessionResult, error) {
	return c.SessionAsync().Receive()
}

// FutureVersionResult is a future promise to deliver the result of a version
// RPC invocation (or an applicable error).
type FutureVersionResult chan *response

// Receive waits for the response promised by the future and returns the version
// result.
func (r FutureVersionResult) Receive() (map[string]model.VersionResult,
	error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as a version result object.
	var vr map[string]model.VersionResult
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
	cmd := model.NewVersionCmd()
	return c.sendCmd(cmd)
}

// Version returns information about the server's JSON-RPC API versions.
func (c *Client) Version() (map[string]model.VersionResult, error) {
	return c.VersionAsync().Receive()
}
