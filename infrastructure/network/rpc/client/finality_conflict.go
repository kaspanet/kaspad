package client

import (
	"encoding/json"

	"github.com/kaspanet/kaspad/infrastructure/network/rpc/model"
	"github.com/pkg/errors"
)

// FutureGetFinalityConflictsResult is a future promise to deliver the result of a
// getFinalityConflicts RPC invocation (or an applicable error).
type FutureGetFinalityConflictsResult chan *response

// GetFinalityConflictsAsync returns an instance of a type that can be used to get
// the result of the RPC at some future time by invoking the Receive function on
// the returned instance.
//
// See GetFinalityConflicts for the blocking version and more details
func (c *Client) GetFinalityConflictsAsync() FutureGetFinalityConflictsResult {
	cmd := model.NewGetFinalityConflictsCmd()
	return c.sendCmd(cmd)
}

// Receive waits for the response promised by the future and returns an error if
// any occurred
func (r FutureGetFinalityConflictsResult) Receive() (*model.GetFinalityConflictsResult, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	var result model.GetFinalityConflictsResult
	if err := json.Unmarshal(res, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetFinalityConflicts request a list of finality conflicts from the server
func (c *Client) GetFinalityConflicts() (*model.GetFinalityConflictsResult, error) {
	return c.GetFinalityConflictsAsync().Receive()
}

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
func (c *Client) ResolveFinalityConflictAsync(
	finalityConflitID int, validBlocks []string, invalidBlocks []string) FutureResolveFinalityConflictResult {

	cmd := model.NewResolveFinalityConflictCmd(finalityConflitID, validBlocks, invalidBlocks)
	return c.sendCmd(cmd)
}

// ResolveFinalityConflict tells the kaspa node how to resolve a finality conflict.
func (c *Client) ResolveFinalityConflict(finalityConflitID int, validBlockHashes []string, invalidBlockHashes []string) error {
	return c.ResolveFinalityConflictAsync(finalityConflitID, validBlockHashes, invalidBlockHashes).Receive()
}
