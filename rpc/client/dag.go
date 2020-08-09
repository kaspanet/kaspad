// Copyright (c) 2014-2017 The btcsuite developers
// Copyright (c) 2015-2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package client

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"github.com/kaspanet/kaspad/util/pointers"

	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/domainmessage"
	"github.com/kaspanet/kaspad/rpc/model"
	"github.com/kaspanet/kaspad/util/daghash"
)

// FutureGetSelectedTipHashResult is a future promise to deliver the result of a
// GetSelectedTipAsync RPC invocation (or an applicable error).
type FutureGetSelectedTipHashResult chan *response

// Receive waits for the response promised by the future and returns the hash of
// the best block in the longest block dag.
func (r FutureGetSelectedTipHashResult) Receive() (*daghash.Hash, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as a string.
	var txHashStr string
	err = json.Unmarshal(res, &txHashStr)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't decode getSelectedTip response")
	}
	return daghash.NewHashFromStr(txHashStr)
}

// GetSelectedTipHashAsync returns an instance of a type that can be used to get
// the result of the RPC at some future time by invoking the Receive function on
// the returned instance.
//
// See GetSelectedTipHash for the blocking version and more details.
func (c *Client) GetSelectedTipHashAsync() FutureGetSelectedTipHashResult {
	cmd := model.NewGetSelectedTipHashCmd()
	return c.sendCmd(cmd)
}

// GetSelectedTipHash returns the hash of the selected tip of the
// Block DAG.
func (c *Client) GetSelectedTipHash() (*daghash.Hash, error) {
	return c.GetSelectedTipHashAsync().Receive()
}

// FutureGetBlockResult is a future promise to deliver the result of a
// GetBlockAsync RPC invocation (or an applicable error).
type FutureGetBlockResult chan *response

// Receive waits for the response promised by the future and returns the raw
// block requested from the server given its hash.
func (r FutureGetBlockResult) Receive() (*domainmessage.MsgBlock, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as a string.
	var blockHex string
	err = json.Unmarshal(res, &blockHex)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't decode getBlock response")
	}

	// Decode the serialized block hex to raw bytes.
	serializedBlock, err := hex.DecodeString(blockHex)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't decode block hex")
	}

	// Deserialize the block and return it.
	var msgBlock domainmessage.MsgBlock
	err = msgBlock.Deserialize(bytes.NewReader(serializedBlock))
	if err != nil {
		return nil, err
	}
	return &msgBlock, nil
}

// GetBlockAsync returns an instance of a type that can be used to get the
// result of the RPC at some future time by invoking the Receive function on the
// returned instance.
//
// See GetBlock for the blocking version and more details.
func (c *Client) GetBlockAsync(blockHash *daghash.Hash, subnetworkID *string) FutureGetBlockResult {
	hash := ""
	if blockHash != nil {
		hash = blockHash.String()
	}

	cmd := model.NewGetBlockCmd(hash, pointers.Bool(false), pointers.Bool(false), subnetworkID)
	return c.sendCmd(cmd)
}

// GetBlock returns a raw block from the server given its hash.
//
// See GetBlockVerbose to retrieve a data structure with information about the
// block instead.
func (c *Client) GetBlock(blockHash *daghash.Hash, subnetworkID *string) (*domainmessage.MsgBlock, error) {
	return c.GetBlockAsync(blockHash, subnetworkID).Receive()
}

// FutureGetBlocksResult is a future promise to deliver the result of a
// GetBlocksAsync RPC invocation (or an applicable error).
type FutureGetBlocksResult chan *response

// Receive waits for the response promised by the future and returns the blocks
// starting from lowHash up to the virtual ordered by blue score.
func (r FutureGetBlocksResult) Receive() (*model.GetBlocksResult, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	var result model.GetBlocksResult
	if err := json.Unmarshal(res, &result); err != nil {
		return nil, errors.Wrap(err, string(res))
	}
	return &result, nil
}

// GetBlocksAsync returns an instance of a type that can be used to get the
// result of the RPC at some future time by invoking the Receive function on the
// returned instance.
//
// See GetBlocks for the blocking version and more details.
func (c *Client) GetBlocksAsync(includeRawBlockData bool, IncludeVerboseBlockData bool, lowHash *string) FutureGetBlocksResult {
	cmd := model.NewGetBlocksCmd(includeRawBlockData, IncludeVerboseBlockData, lowHash)
	return c.sendCmd(cmd)
}

// GetBlocks returns the blocks starting from lowHash up to the virtual ordered
// by blue score.
func (c *Client) GetBlocks(includeRawBlockData bool, includeVerboseBlockData bool, lowHash *string) (*model.GetBlocksResult, error) {
	return c.GetBlocksAsync(includeRawBlockData, includeVerboseBlockData, lowHash).Receive()
}

// FutureGetBlockVerboseResult is a future promise to deliver the result of a
// GetBlockVerboseAsync RPC invocation (or an applicable error).
type FutureGetBlockVerboseResult chan *response

// Receive waits for the response promised by the future and returns the data
// structure from the server with information about the requested block.
func (r FutureGetBlockVerboseResult) Receive() (*model.GetBlockVerboseResult, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal the raw result into a BlockResult.
	var blockResult model.GetBlockVerboseResult
	err = json.Unmarshal(res, &blockResult)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't decode getBlock response")
	}
	return &blockResult, nil
}

// GetBlockVerboseAsync returns an instance of a type that can be used to get
// the result of the RPC at some future time by invoking the Receive function on
// the returned instance.
//
// See GetBlockVerbose for the blocking version and more details.
func (c *Client) GetBlockVerboseAsync(blockHash *daghash.Hash, subnetworkID *string) FutureGetBlockVerboseResult {
	hash := ""
	if blockHash != nil {
		hash = blockHash.String()
	}

	cmd := model.NewGetBlockCmd(hash, pointers.Bool(true), pointers.Bool(false), subnetworkID)
	return c.sendCmd(cmd)
}

// GetBlockVerbose returns a data structure from the server with information
// about a block given its hash.
//
// See GetBlockVerboseTx to retrieve transaction data structures as well.
// See GetBlock to retrieve a raw block instead.
func (c *Client) GetBlockVerbose(blockHash *daghash.Hash, subnetworkID *string) (*model.GetBlockVerboseResult, error) {
	return c.GetBlockVerboseAsync(blockHash, subnetworkID).Receive()
}

// GetBlockVerboseTxAsync returns an instance of a type that can be used to get
// the result of the RPC at some future time by invoking the Receive function on
// the returned instance.
//
// See GetBlockVerboseTx or the blocking version and more details.
func (c *Client) GetBlockVerboseTxAsync(blockHash *daghash.Hash, subnetworkID *string) FutureGetBlockVerboseResult {
	hash := ""
	if blockHash != nil {
		hash = blockHash.String()
	}

	cmd := model.NewGetBlockCmd(hash, pointers.Bool(true), pointers.Bool(true), subnetworkID)
	return c.sendCmd(cmd)
}

// GetBlockVerboseTx returns a data structure from the server with information
// about a block and its transactions given its hash.
//
// See GetBlockVerbose if only transaction hashes are preferred.
// See GetBlock to retrieve a raw block instead.
func (c *Client) GetBlockVerboseTx(blockHash *daghash.Hash, subnetworkID *string) (*model.GetBlockVerboseResult, error) {
	return c.GetBlockVerboseTxAsync(blockHash, subnetworkID).Receive()
}

// FutureGetBlockCountResult is a future promise to deliver the result of a
// GetBlockCountAsync RPC invocation (or an applicable error).
type FutureGetBlockCountResult chan *response

// Receive waits for the response promised by the future and returns the number
// of blocks in the longest block dag.
func (r FutureGetBlockCountResult) Receive() (int64, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return 0, err
	}

	// Unmarshal the result as an int64.
	var count int64
	err = json.Unmarshal(res, &count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetBlockCountAsync returns an instance of a type that can be used to get the
// result of the RPC at some future time by invoking the Receive function on the
// returned instance.
//
// See GetBlockCount for the blocking version and more details.
func (c *Client) GetBlockCountAsync() FutureGetBlockCountResult {
	cmd := model.NewGetBlockCountCmd()
	return c.sendCmd(cmd)
}

// GetBlockCount returns the number of blocks in the longest block dag.
func (c *Client) GetBlockCount() (int64, error) {
	return c.GetBlockCountAsync().Receive()
}

// FutureGetChainFromBlockResult is a future promise to deliver the result of a
// GetChainFromBlockAsync RPC invocation (or an applicable error).
type FutureGetChainFromBlockResult chan *response

// Receive waits for the response promised by the future and returns the selected
// parent chain starting from startHash up to the virtual. If startHash is not in
// the selected parent chain, it goes down the DAG until it does reach a hash in
// the selected parent chain while collecting hashes into RemovedChainBlockHashes.
func (r FutureGetChainFromBlockResult) Receive() (*model.GetChainFromBlockResult, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	var result model.GetChainFromBlockResult
	if err := json.Unmarshal(res, &result); err != nil {
		return nil, errors.Wrap(err, "couldn't decode getChainFromBlock response")
	}
	return &result, nil
}

// GetChainFromBlockAsync returns an instance of a type that can be used to get the
// result of the RPC at some future time by invoking the Receive function on the
// returned instance.
//
// See GetChainFromBlock for the blocking version and more details.
func (c *Client) GetChainFromBlockAsync(includeBlocks bool, startHash *string) FutureGetChainFromBlockResult {
	cmd := model.NewGetChainFromBlockCmd(includeBlocks, startHash)
	return c.sendCmd(cmd)
}

// GetChainFromBlock returns the selected parent chain starting from startHash
// up to the virtual. If startHash is not in the selected parent chain, it goes
// down the DAG until it does reach a hash in the selected parent chain while
// collecting hashes into RemovedChainBlockHashes.
func (c *Client) GetChainFromBlock(includeBlocks bool, startHash *string) (*model.GetChainFromBlockResult, error) {
	return c.GetChainFromBlockAsync(includeBlocks, startHash).Receive()
}

// FutureGetDifficultyResult is a future promise to deliver the result of a
// GetDifficultyAsync RPC invocation (or an applicable error).
type FutureGetDifficultyResult chan *response

// Receive waits for the response promised by the future and returns the
// proof-of-work difficulty as a multiple of the minimum difficulty.
func (r FutureGetDifficultyResult) Receive() (float64, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return 0, err
	}

	// Unmarshal the result as a float64.
	var difficulty float64
	err = json.Unmarshal(res, &difficulty)
	if err != nil {
		return 0, err
	}
	return difficulty, nil
}

// GetDifficultyAsync returns an instance of a type that can be used to get the
// result of the RPC at some future time by invoking the Receive function on the
// returned instance.
//
// See GetDifficulty for the blocking version and more details.
func (c *Client) GetDifficultyAsync() FutureGetDifficultyResult {
	cmd := model.NewGetDifficultyCmd()
	return c.sendCmd(cmd)
}

// GetDifficulty returns the proof-of-work difficulty as a multiple of the
// minimum difficulty.
func (c *Client) GetDifficulty() (float64, error) {
	return c.GetDifficultyAsync().Receive()
}

// FutureGetBlockDAGInfoResult is a promise to deliver the result of a
// GetBlockDAGInfoAsync RPC invocation (or an applicable error).
type FutureGetBlockDAGInfoResult chan *response

// Receive waits for the response promised by the future and returns dag info
// result provided by the server.
func (r FutureGetBlockDAGInfoResult) Receive() (*model.GetBlockDAGInfoResult, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	var dagInfo model.GetBlockDAGInfoResult
	if err := json.Unmarshal(res, &dagInfo); err != nil {
		return nil, errors.Wrap(err, "couldn't decode getBlockDagInfo response")
	}
	return &dagInfo, nil
}

// GetBlockDAGInfoAsync returns an instance of a type that can be used to get
// the result of the RPC at some future time by invoking the Receive function
// on the returned instance.
//
// See GetBlockDAGInfo for the blocking version and more details.
func (c *Client) GetBlockDAGInfoAsync() FutureGetBlockDAGInfoResult {
	cmd := model.NewGetBlockDAGInfoCmd()
	return c.sendCmd(cmd)
}

// GetBlockDAGInfo returns information related to the processing state of
// various dag-specific details such as the current difficulty from the tip
// of the main dag.
func (c *Client) GetBlockDAGInfo() (*model.GetBlockDAGInfoResult, error) {
	return c.GetBlockDAGInfoAsync().Receive()
}

// FutureGetBlockHashResult is a future promise to deliver the result of a
// GetBlockHashAsync RPC invocation (or an applicable error).
type FutureGetBlockHashResult chan *response

// Receive waits for the response promised by the future and returns the hash of
// the block in the best block dag at the given height.
func (r FutureGetBlockHashResult) Receive() (*daghash.Hash, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal the result as a string-encoded sha.
	var txHashStr string
	err = json.Unmarshal(res, &txHashStr)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return daghash.NewHashFromStr(txHashStr)
}

// FutureGetBlockHeaderResult is a future promise to deliver the result of a
// GetBlockHeaderAsync RPC invocation (or an applicable error).
type FutureGetBlockHeaderResult chan *response

// Receive waits for the response promised by the future and returns the
// blockheader requested from the server given its hash.
func (r FutureGetBlockHeaderResult) Receive() (*domainmessage.BlockHeader, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as a string.
	var bhHex string
	err = json.Unmarshal(res, &bhHex)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't decode getBlockHeader response")
	}

	serializedBH, err := hex.DecodeString(bhHex)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't decode block header hex")
	}

	// Deserialize the blockheader and return it.
	var bh domainmessage.BlockHeader
	err = bh.Deserialize(bytes.NewReader(serializedBH))
	if err != nil {
		return nil, err
	}

	return &bh, err
}

// GetBlockHeaderAsync returns an instance of a type that can be used to get the
// result of the RPC at some future time by invoking the Receive function on the
// returned instance.
//
// See GetBlockHeader for the blocking version and more details.
func (c *Client) GetBlockHeaderAsync(blockHash *daghash.Hash) FutureGetBlockHeaderResult {
	hash := ""
	if blockHash != nil {
		hash = blockHash.String()
	}

	cmd := model.NewGetBlockHeaderCmd(hash, pointers.Bool(false))
	return c.sendCmd(cmd)
}

// GetBlockHeader returns the blockheader from the server given its hash.
//
// See GetBlockHeaderVerbose to retrieve a data structure with information about the
// block instead.
func (c *Client) GetBlockHeader(blockHash *daghash.Hash) (*domainmessage.BlockHeader, error) {
	return c.GetBlockHeaderAsync(blockHash).Receive()
}

// FutureGetBlockHeaderVerboseResult is a future promise to deliver the result of a
// GetBlockAsync RPC invocation (or an applicable error).
type FutureGetBlockHeaderVerboseResult chan *response

// Receive waits for the response promised by the future and returns the
// data structure of the blockheader requested from the server given its hash.
func (r FutureGetBlockHeaderVerboseResult) Receive() (*model.GetBlockHeaderVerboseResult, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as a string.
	var bh model.GetBlockHeaderVerboseResult
	err = json.Unmarshal(res, &bh)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't decode getBlockHeader response")
	}

	return &bh, nil
}

// GetBlockHeaderVerboseAsync returns an instance of a type that can be used to get the
// result of the RPC at some future time by invoking the Receive function on the
// returned instance.
//
// See GetBlockHeader for the blocking version and more details.
func (c *Client) GetBlockHeaderVerboseAsync(blockHash *daghash.Hash) FutureGetBlockHeaderVerboseResult {
	hash := ""
	if blockHash != nil {
		hash = blockHash.String()
	}

	cmd := model.NewGetBlockHeaderCmd(hash, pointers.Bool(true))
	return c.sendCmd(cmd)
}

// GetBlockHeaderVerbose returns a data structure with information about the
// blockheader from the server given its hash.
//
// See GetBlockHeader to retrieve a blockheader instead.
func (c *Client) GetBlockHeaderVerbose(blockHash *daghash.Hash) (*model.GetBlockHeaderVerboseResult, error) {
	return c.GetBlockHeaderVerboseAsync(blockHash).Receive()
}

// FutureGetMempoolEntryResult is a future promise to deliver the result of a
// GetMempoolEntryAsync RPC invocation (or an applicable error).
type FutureGetMempoolEntryResult chan *response

// Receive waits for the response promised by the future and returns a data
// structure with information about the transaction in the memory pool given
// its hash.
func (r FutureGetMempoolEntryResult) Receive() (*model.GetMempoolEntryResult, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal the result as an array of strings.
	var mempoolEntryResult model.GetMempoolEntryResult
	err = json.Unmarshal(res, &mempoolEntryResult)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't decode getMempoolEntry response")
	}

	return &mempoolEntryResult, nil
}

// GetMempoolEntryAsync returns an instance of a type that can be used to get the
// result of the RPC at some future time by invoking the Receive function on the
// returned instance.
//
// See GetMempoolEntry for the blocking version and more details.
func (c *Client) GetMempoolEntryAsync(txHash string) FutureGetMempoolEntryResult {
	cmd := model.NewGetMempoolEntryCmd(txHash)
	return c.sendCmd(cmd)
}

// GetMempoolEntry returns a data structure with information about the
// transaction in the memory pool given its hash.
func (c *Client) GetMempoolEntry(txHash string) (*model.GetMempoolEntryResult, error) {
	return c.GetMempoolEntryAsync(txHash).Receive()
}

// FutureGetRawMempoolResult is a future promise to deliver the result of a
// GetRawMempoolAsync RPC invocation (or an applicable error).
type FutureGetRawMempoolResult chan *response

// Receive waits for the response promised by the future and returns the hashes
// of all transactions in the memory pool.
func (r FutureGetRawMempoolResult) Receive() ([]*daghash.Hash, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal the result as an array of strings.
	var txHashStrs []string
	err = json.Unmarshal(res, &txHashStrs)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't decode getRawMempool response")
	}

	// Create a slice of ShaHash arrays from the string slice.
	txHashes := make([]*daghash.Hash, 0, len(txHashStrs))
	for _, hashStr := range txHashStrs {
		txHash, err := daghash.NewHashFromStr(hashStr)
		if err != nil {
			return nil, err
		}
		txHashes = append(txHashes, txHash)
	}

	return txHashes, nil
}

// GetRawMempoolAsync returns an instance of a type that can be used to get the
// result of the RPC at some future time by invoking the Receive function on the
// returned instance.
//
// See GetRawMempool for the blocking version and more details.
func (c *Client) GetRawMempoolAsync() FutureGetRawMempoolResult {
	cmd := model.NewGetRawMempoolCmd(pointers.Bool(false))
	return c.sendCmd(cmd)
}

// GetRawMempool returns the hashes of all transactions in the memory pool.
//
// See GetRawMempoolVerbose to retrieve data structures with information about
// the transactions instead.
func (c *Client) GetRawMempool() ([]*daghash.Hash, error) {
	return c.GetRawMempoolAsync().Receive()
}

// FutureGetRawMempoolVerboseResult is a future promise to deliver the result of
// a GetRawMempoolVerboseAsync RPC invocation (or an applicable error).
type FutureGetRawMempoolVerboseResult chan *response

// Receive waits for the response promised by the future and returns a map of
// transaction hashes to an associated data structure with information about the
// transaction for all transactions in the memory pool.
func (r FutureGetRawMempoolVerboseResult) Receive() (map[string]model.GetRawMempoolVerboseResult, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal the result as a map of strings (tx shas) to their detailed
	// results.
	var mempoolItems map[string]model.GetRawMempoolVerboseResult
	err = json.Unmarshal(res, &mempoolItems)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't decode getRawMempool response")
	}
	return mempoolItems, nil
}

// GetRawMempoolVerboseAsync returns an instance of a type that can be used to
// get the result of the RPC at some future time by invoking the Receive
// function on the returned instance.
//
// See GetRawMempoolVerbose for the blocking version and more details.
func (c *Client) GetRawMempoolVerboseAsync() FutureGetRawMempoolVerboseResult {
	cmd := model.NewGetRawMempoolCmd(pointers.Bool(true))
	return c.sendCmd(cmd)
}

// GetRawMempoolVerbose returns a map of transaction hashes to an associated
// data structure with information about the transaction for all transactions in
// the memory pool.
//
// See GetRawMempool to retrieve only the transaction hashes instead.
func (c *Client) GetRawMempoolVerbose() (map[string]model.GetRawMempoolVerboseResult, error) {
	return c.GetRawMempoolVerboseAsync().Receive()
}

// FutureGetSubnetworkResult is a future promise to deliver the result of a
// GetSubnetworkAsync RPC invocation (or an applicable error).
type FutureGetSubnetworkResult chan *response

// Receive waits for the response promised by the future and returns information
// regarding the requested subnetwork
func (r FutureGetSubnetworkResult) Receive() (*model.GetSubnetworkResult, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as a getSubnetwork result object.
	var getSubnetworkResult *model.GetSubnetworkResult
	err = json.Unmarshal(res, &getSubnetworkResult)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't decode getSubnetwork response")
	}

	return getSubnetworkResult, nil
}

// GetSubnetworkAsync returns an instance of a type that can be used to get the result
// of the RPC at some future time by invoking the Receive function on the
// returned instance.
//
// See GetSubnetwork for the blocking version and more details.
func (c *Client) GetSubnetworkAsync(subnetworkID string) FutureGetSubnetworkResult {
	cmd := model.NewGetSubnetworkCmd(subnetworkID)
	return c.sendCmd(cmd)
}

// GetSubnetwork provides information about a subnetwork given its ID.
func (c *Client) GetSubnetwork(subnetworkID string) (*model.GetSubnetworkResult, error) {
	return c.GetSubnetworkAsync(subnetworkID).Receive()
}

// FutureGetTxOutResult is a future promise to deliver the result of a
// GetTxOutAsync RPC invocation (or an applicable error).
type FutureGetTxOutResult chan *response

// Receive waits for the response promised by the future and returns a
// transaction given its hash.
func (r FutureGetTxOutResult) Receive() (*model.GetTxOutResult, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// take care of the special case where the output has been spent already
	// it should return the string "null"
	if string(res) == "null" {
		return nil, nil
	}

	// Unmarshal result as an gettxout result object.
	var txOutInfo *model.GetTxOutResult
	err = json.Unmarshal(res, &txOutInfo)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't decode getTxOut response")
	}

	return txOutInfo, nil
}

// GetTxOutAsync returns an instance of a type that can be used to get
// the result of the RPC at some future time by invoking the Receive function on
// the returned instance.
//
// See GetTxOut for the blocking version and more details.
func (c *Client) GetTxOutAsync(txHash *daghash.Hash, index uint32, mempool bool) FutureGetTxOutResult {
	hash := ""
	if txHash != nil {
		hash = txHash.String()
	}

	cmd := model.NewGetTxOutCmd(hash, index, &mempool)
	return c.sendCmd(cmd)
}

// GetTxOut returns the transaction output info if it's unspent and
// nil, otherwise.
func (c *Client) GetTxOut(txHash *daghash.Hash, index uint32, mempool bool) (*model.GetTxOutResult, error) {
	return c.GetTxOutAsync(txHash, index, mempool).Receive()
}

// FutureRescanBlocksResult is a future promise to deliver the result of a
// RescanBlocksAsync RPC invocation (or an applicable error).
type FutureRescanBlocksResult chan *response

// Receive waits for the response promised by the future and returns the
// discovered rescanblocks data.
func (r FutureRescanBlocksResult) Receive() ([]model.RescannedBlock, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	var rescanBlocksResult []model.RescannedBlock
	err = json.Unmarshal(res, &rescanBlocksResult)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't decode rescanBlocks response")
	}

	return rescanBlocksResult, nil
}

// RescanBlocksAsync returns an instance of a type that can be used to get the
// result of the RPC at some future time by invoking the Receive function on the
// returned instance.
//
// See RescanBlocks for the blocking version and more details.
func (c *Client) RescanBlocksAsync(blockHashes []*daghash.Hash) FutureRescanBlocksResult {
	strBlockHashes := make([]string, len(blockHashes))
	for i := range blockHashes {
		strBlockHashes[i] = blockHashes[i].String()
	}

	cmd := model.NewRescanBlocksCmd(strBlockHashes)
	return c.sendCmd(cmd)
}

// RescanBlocks rescans the blocks identified by blockHashes, in order, using
// the client's loaded transaction filter. The blocks do not need to be on the
// main dag, but they do need to be adjacent to each other.
func (c *Client) RescanBlocks(blockHashes []*daghash.Hash) ([]model.RescannedBlock, error) {
	return c.RescanBlocksAsync(blockHashes).Receive()
}
