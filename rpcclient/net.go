// Copyright (c) 2014-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package rpcclient

import (
	"encoding/json"

	"github.com/daglabs/btcd/btcjson"
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
	cmd := btcjson.NewAddManualNodeCmd(host, btcjson.Bool(false))
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
func (r FutureGetManualNodeInfoResult) Receive() ([]btcjson.GetManualNodeInfoResult, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal as an array of getmanualnodeinfo result objects.
	var nodeInfo []btcjson.GetManualNodeInfoResult
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
	cmd := btcjson.NewGetManualNodeInfoCmd(peer, nil)
	return c.sendCmd(cmd)
}

// GetManualNodeInfo returns information about manually added (persistent) peers.
//
// See GetManualNodeInfoNoDNS to retrieve only a list of the added (persistent)
// peers.
func (c *Client) GetManualNodeInfo(peer string) ([]btcjson.GetManualNodeInfoResult, error) {
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
	cmd := btcjson.NewGetManualNodeInfoCmd(peer, btcjson.Bool(false))
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
	cmd := btcjson.NewGetConnectionCountCmd()
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
	cmd := btcjson.NewPingCmd()
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
func (r FutureGetPeerInfoResult) Receive() ([]btcjson.GetPeerInfoResult, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as an array of getpeerinfo result objects.
	var peerInfo []btcjson.GetPeerInfoResult
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
	cmd := btcjson.NewGetPeerInfoCmd()
	return c.sendCmd(cmd)
}

// GetPeerInfo returns data about each connected network peer.
func (c *Client) GetPeerInfo() ([]btcjson.GetPeerInfoResult, error) {
	return c.GetPeerInfoAsync().Receive()
}

// FutureGetNetTotalsResult is a future promise to deliver the result of a
// GetNetTotalsAsync RPC invocation (or an applicable error).
type FutureGetNetTotalsResult chan *response

// Receive waits for the response promised by the future and returns network
// traffic statistics.
func (r FutureGetNetTotalsResult) Receive() (*btcjson.GetNetTotalsResult, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as a getnettotals result object.
	var totals btcjson.GetNetTotalsResult
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
	cmd := btcjson.NewGetNetTotalsCmd()
	return c.sendCmd(cmd)
}

// GetNetTotals returns network traffic statistics.
func (c *Client) GetNetTotals() (*btcjson.GetNetTotalsResult, error) {
	return c.GetNetTotalsAsync().Receive()
}
