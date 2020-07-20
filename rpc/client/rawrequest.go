// Copyright (c) 2014-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package client

import (
	"encoding/json"
	"github.com/kaspanet/kaspad/rpc/model"
	"github.com/pkg/errors"
)

// FutureRawResult is a future promise to deliver the result of a RawRequest RPC
// invocation (or an applicable error).
type FutureRawResult chan *response

// Receive waits for the response promised by the future and returns the raw
// response, or an error if the request was unsuccessful.
func (r FutureRawResult) Receive() (json.RawMessage, error) {
	return receiveFuture(r)
}

// RawRequestAsync returns an instance of a type that can be used to get the
// result of a custom RPC request at some future time by invoking the Receive
// function on the returned instance.
//
// See RawRequest for the blocking version and more details.
func (c *Client) RawRequestAsync(method string, params []json.RawMessage) FutureRawResult {
	// Method may not be empty.
	if method == "" {
		return newFutureError(errors.New("no method"))
	}

	// Marshal parameters as "[]" instead of "null" when no parameters
	// are passed.
	if params == nil {
		params = []json.RawMessage{}
	}

	// Create a raw JSON-RPC request using the provided method and params
	// and marshal it. This is done rather than using the sendCmd function
	// since that relies on marshalling registered jsonrpc commands rather
	// than custom commands.
	id := c.NextID()
	rawRequest := &model.Request{
		JSONRPC: "1.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}
	marshalledJSON, err := json.Marshal(rawRequest)
	if err != nil {
		return newFutureError(err)
	}

	// Generate the request.
	jReqData := &jsonRequestData{
		id:             id,
		method:         method,
		cmd:            nil,
		marshalledJSON: marshalledJSON,
	}

	// Send the request and return its response channel
	return c.sendRequest(jReqData)
}

// RawRequest allows the caller to send a raw or custom request to the server.
// This method may be used to send and receive requests and responses for
// requests that are not handled by this client package, or to proxy partially
// unmarshaled requests to another JSON-RPC server if a request cannot be
// handled directly.
func (c *Client) RawRequest(method string, params []json.RawMessage) (json.RawMessage, error) {
	return c.RawRequestAsync(method, params).Receive()
}
