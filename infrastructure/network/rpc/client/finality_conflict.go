package client

import (
	"encoding/json"

	"github.com/kaspanet/kaspad/infrastructure/network/rpc/model"
	"github.com/pkg/errors"
)

// FutureResolveFinalityConflictResult is a future promise to deliver the result of a
// resolveFinalityConflictAsync RPC invocation (or an applicable error).
type FutureResolveFinalityConflictResult chan *response

// Receive waits for the response promised by the future and returns an error if
// any occurred when resolving the finality conflict.
func (r FutureResolveFinalityConflictResult) Receive() error {
	res, err := receiveFuture(r)
	if err != nil {
		return err
	}

	if string(res) != "null" {
		var result string
		err = json.Unmarshal(res, &result)
		if err != nil {
			return err
		}

		return errors.New(result)
	}

	return nil
}

// ResolveFinalityConflictAsync returns an instance of a type that can be used to get the
// result of the RPC at some future time by invoking the Receive function on the
// returned instance.
//
// See ResolveFinalityConflict for the blocking version and more details.
func (c *Client) ResolveFinalityConflictAsync(finalityBlockHash string) FutureResolveFinalityConflictResult {

	cmd := model.NewResolveFinalityConflictCmd(finalityBlockHash)
	return c.sendCmd(cmd)
}

// ResolveFinalityConflict tells the kaspa node how to resolve a finality conflict.
func (c *Client) ResolveFinalityConflict(finalityBlockHash string) error {
	return c.ResolveFinalityConflictAsync(finalityBlockHash).Receive()
}
