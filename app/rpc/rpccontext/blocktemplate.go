package rpccontext

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/mining"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/kaspanet/kaspad/util/random"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// blockTemplateNonceRange is two 64-bit big-endian hexadecimal integers which
	// represent the valid ranges of nonces returned by the getBlockTemplate
	// RPC.
	blockTemplateNonceRange = "000000000000ffffffffffff"

	// blockTemplateRegenerateSeconds is the number of seconds that must pass before
	// a new template is generated when the parent block hashes has not
	// changed and there have been changes to the available transactions
	// in the memory pool.
	blockTemplateRegenerateSeconds = 60
)

var (
	// blockTemplateMutableFields are the manipulations the server allows to be made
	// to block templates generated by the getBlockTemplate RPC. It is
	// declared here to avoid the overhead of creating the slice on every
	// invocation for constant data.
	blockTemplateMutableFields = []string{
		"time", "transactions/add", "parentblock", "coinbase/append",
	}
)

// BlockTemplateState houses state that is used in between multiple RPC invocations to
// getBlockTemplate.
type BlockTemplateState struct {
	sync.Mutex

	context *Context

	lastTxUpdate  mstime.Time
	lastGenerated mstime.Time
	tipHashes     []*daghash.Hash
	minTimestamp  mstime.Time
	template      *mining.BlockTemplate
	notifyMap     map[string]map[int64]chan struct{}
	payAddress    util.Address
}

// NewBlockTemplateState returns a new instance of a BlockTemplateState with all internal
// fields initialized and ready to use.
func NewBlockTemplateState(context *Context) *BlockTemplateState {
	return &BlockTemplateState{
		context:   context,
		notifyMap: make(map[string]map[int64]chan struct{}),
	}
}

func (bt *BlockTemplateState) Update(payAddress util.Address) error {
	generator := bt.context.BlockTemplateGenerator
	lastTxUpdate := generator.TxSource().LastUpdated()
	if lastTxUpdate.IsZero() {
		lastTxUpdate = mstime.Now()
	}

	// Generate a new block template when the current best block has
	// changed or the transactions in the memory pool have been updated and
	// it has been at least gbtRegenerateSecond since the last template was
	// generated.
	var msgBlock *appmessage.MsgBlock
	var targetDifficulty string
	tipHashes := bt.context.DAG.TipHashes()
	template := bt.template
	if template == nil || bt.tipHashes == nil ||
		!daghash.AreEqual(bt.tipHashes, tipHashes) ||
		bt.payAddress.String() != payAddress.String() ||
		(bt.lastTxUpdate != lastTxUpdate &&
			mstime.Now().After(bt.lastGenerated.Add(time.Second*
				blockTemplateRegenerateSeconds))) {

		// Reset the previous best hash the block template was generated
		// against so any errors below cause the next invocation to try
		// again.
		bt.tipHashes = nil

		// Create a new block template that has a coinbase which anyone
		// can redeem. This is only acceptable because the returned
		// block template doesn't include the coinbase, so the caller
		// will ultimately create their own coinbase which pays to the
		// appropriate address(es).

		extraNonce, err := random.Uint64()
		if err != nil {
			return errors.Wrapf(err, "failed to randomize extra nonce")
		}

		blockTemplate, err := generator.NewBlockTemplate(payAddress, extraNonce)
		if err != nil {
			return errors.Wrapf(err, "failed to create new block template")
		}
		template = blockTemplate
		msgBlock = template.Block
		targetDifficulty = fmt.Sprintf("%064x", util.CompactToBig(msgBlock.Header.Bits))

		// Get the minimum allowed timestamp for the block based on the
		// median timestamp of the last several blocks per the DAG
		// consensus rules.
		minTimestamp := bt.context.DAG.NextBlockMinimumTime()

		// Update work state to ensure another block template isn't
		// generated until needed.
		bt.template = template
		bt.lastGenerated = mstime.Now()
		bt.lastTxUpdate = lastTxUpdate
		bt.tipHashes = tipHashes
		bt.minTimestamp = minTimestamp
		bt.payAddress = payAddress

		log.Debugf("Generated block template (timestamp %s, "+
			"target %s, merkle root %s)",
			msgBlock.Header.Timestamp, targetDifficulty,
			msgBlock.Header.HashMerkleRoot)

		// Notify any clients that are long polling about the new
		// template.
		bt.notifyLongPollers(tipHashes, lastTxUpdate)
	} else {
		// At this point, there is a saved block template and another
		// request for a template was made, but either the available
		// transactions haven't change or it hasn't been long enough to
		// trigger a new block template to be generated. So, update the
		// existing block template.

		// Set locals for convenience.
		msgBlock = template.Block
		targetDifficulty = fmt.Sprintf("%064x",
			util.CompactToBig(msgBlock.Header.Bits))

		// Update the time of the block template to the current time
		// while accounting for the median time of the past several
		// blocks per the DAG consensus rules.
		err := generator.UpdateBlockTime(msgBlock)
		if err != nil {
			return errors.Wrapf(err, "failed to update block time")
		}
		msgBlock.Header.Nonce = 0

		log.Debugf("Updated block template (timestamp %s, "+
			"target %s)", msgBlock.Header.Timestamp,
			targetDifficulty)
	}

	return nil
}

func (bt *BlockTemplateState) Response() (*appmessage.GetBlockTemplateResponseMessage, error) {
	dag := bt.context.DAG
	// Ensure the timestamps are still in valid range for the template.
	// This should really only ever happen if the local clock is changed
	// after the template is generated, but it's important to avoid serving
	// block templates that will be delayed on other nodes.
	template := bt.template
	msgBlock := template.Block
	header := &msgBlock.Header
	adjustedTime := dag.Now()
	maxTime := adjustedTime.Add(time.Millisecond * time.Duration(dag.TimestampDeviationTolerance))
	if header.Timestamp.After(maxTime) {
		errorMessage := &appmessage.GetBlockTemplateResponseMessage{}
		errorMessage.Error = &appmessage.RPCError{
			Message: fmt.Sprintf("The template time is after the "+
				"maximum allowed time for a block - template "+
				"time %s, maximum time %s", adjustedTime,
				maxTime),
		}
		return errorMessage, nil
	}

	// Convert each transaction in the block template to a template result
	// transaction. The result does not include the coinbase, so notice
	// the adjustments to the various lengths and indices.
	numTx := len(msgBlock.Transactions)
	transactions := make([]appmessage.GetBlockTemplateTransactionMessage, 0, numTx-1)
	txIndex := make(map[daghash.TxID]int64, numTx)
	for i, tx := range msgBlock.Transactions {
		txID := tx.TxID()
		txIndex[*txID] = int64(i)

		// Create an array of 1-based indices to transactions that come
		// before this one in the transactions list which this one
		// depends on. This is necessary since the created block must
		// ensure proper ordering of the dependencies. A map is used
		// before creating the final array to prevent duplicate entries
		// when multiple inputs reference the same transaction.
		dependsMap := make(map[int64]struct{})
		for _, txIn := range tx.TxIn {
			if idx, ok := txIndex[txIn.PreviousOutpoint.TxID]; ok {
				dependsMap[idx] = struct{}{}
			}
		}
		depends := make([]int64, 0, len(dependsMap))
		for idx := range dependsMap {
			depends = append(depends, idx)
		}

		// Serialize the transaction for later conversion to hex.
		txBuf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))
		if err := tx.Serialize(txBuf); err != nil {
			errorMessage := &appmessage.GetBlockTemplateResponseMessage{}
			errorMessage.Error = &appmessage.RPCError{
				Message: fmt.Sprintf("Failed to serialize transaction: %s", err),
			}
			return errorMessage, nil
		}

		resultTx := appmessage.GetBlockTemplateTransactionMessage{
			Data:    hex.EncodeToString(txBuf.Bytes()),
			ID:      txID.String(),
			Depends: depends,
			Mass:    template.TxMasses[i],
			Fee:     template.Fees[i],
		}
		transactions = append(transactions, resultTx)
	}

	// Generate the block template reply. Note that following mutations are
	// implied by the included or omission of fields:
	//  Including MinTime -> time/decrement
	//  Omitting CoinbaseTxn -> coinbase, generation
	targetDifficulty := fmt.Sprintf("%064x", util.CompactToBig(header.Bits))
	longPollID := bt.encodeLongPollID(bt.tipHashes, bt.payAddress, bt.lastGenerated)

	// Check whether this node is synced with the rest of of the
	// network. There's almost never a good reason to mine on top
	// of an unsynced DAG, and miners are generally expected not to
	// mine when isSynced is false.
	// This is not a straight-up error because the choice of whether
	// to mine or not is the responsibility of the miner rather
	// than the node's.
	isSynced := bt.context.BlockTemplateGenerator.IsSynced()
	isConnected := bt.context.ConnectionManager.ConnectionCount() == 0

	reply := appmessage.GetBlockTemplateResponseMessage{
		Bits:                 strconv.FormatInt(int64(header.Bits), 16),
		CurrentTime:          header.Timestamp.UnixMilliseconds(),
		ParentHashes:         daghash.Strings(header.ParentHashes),
		MassLimit:            appmessage.MaxMassPerBlock,
		Transactions:         transactions,
		HashMerkleRoot:       header.HashMerkleRoot.String(),
		AcceptedIDMerkleRoot: header.AcceptedIDMerkleRoot.String(),
		UTXOCommitment:       header.UTXOCommitment.String(),
		Version:              header.Version,
		LongPollID:           longPollID,
		TargetDifficulty:     targetDifficulty,
		MinTime:              bt.minTimestamp.UnixMilliseconds(),
		MaxTime:              maxTime.UnixMilliseconds(),
		MutableFields:        blockTemplateMutableFields,
		NonceRange:           blockTemplateNonceRange,
		IsSynced:             isSynced,
		IsConnected:          isConnected,
	}

	return &reply, nil
}

// notifyLongPollers notifies any channels that have been registered to be
// notified when block templates are stale.
//
// This function MUST be called with the state locked.
func (bt *BlockTemplateState) notifyLongPollers(tipHashes []*daghash.Hash, lastGenerated mstime.Time) {
	// Notify anything that is waiting for a block template update from
	// hashes which are not the current tip hashes.
	tipHashesStr := daghash.JoinHashesStrings(tipHashes, "")
	for hashesStr, channels := range bt.notifyMap {
		if hashesStr != tipHashesStr {
			for _, c := range channels {
				close(c)
			}
			delete(bt.notifyMap, hashesStr)
		}
	}

	// Return now if the provided last generated timestamp has not been
	// initialized.
	if lastGenerated.IsZero() {
		return
	}

	// Return now if there is nothing registered for updates to the current
	// best block hash.
	channels, ok := bt.notifyMap[tipHashesStr]
	if !ok {
		return
	}

	// Notify anything that is waiting for a block template update from a
	// block template generated before the most recently generated block
	// template.
	lastGeneratedUnix := lastGenerated.UnixSeconds()
	for lastGen, c := range channels {
		if lastGen < lastGeneratedUnix {
			close(c)
			delete(channels, lastGen)
		}
	}

	// Remove the entry altogether if there are no more registered
	// channels.
	if len(channels) == 0 {
		delete(bt.notifyMap, tipHashesStr)
	}
}

// NotifyBlockAdded uses the newly-added block to notify any long poll
// clients with a new block template when their existing block template is
// stale due to the newly added block.
func (bt *BlockTemplateState) NotifyBlockAdded(block *util.Block) {
	spawn("BlockTemplateState.NotifyBlockAdded", func() {
		bt.Lock()
		defer bt.Unlock()

		bt.notifyLongPollers(block.MsgBlock().Header.ParentHashes, bt.lastTxUpdate)
	})
}

// NotifyMempoolTx uses the new last updated time for the transaction memory
// pool to notify any long poll clients with a new block template when their
// existing block template is stale due to enough time passing and the contents
// of the memory pool changing.
func (bt *BlockTemplateState) NotifyMempoolTx() {
	lastUpdated := bt.context.Mempool.LastUpdated()
	spawn("BlockTemplateState", func() {
		bt.Lock()
		defer bt.Unlock()

		// No need to notify anything if no block templates have been generated
		// yet.
		if bt.tipHashes == nil || bt.lastGenerated.IsZero() {
			return
		}

		if mstime.Now().After(bt.lastGenerated.Add(time.Second *
			blockTemplateRegenerateSeconds)) {

			bt.notifyLongPollers(bt.tipHashes, lastUpdated)
		}
	})
}

// BlockTemplateOrLongPollChan returns a block template if the
// template identified by the provided long poll ID is stale or
// invalid. Otherwise, it returns a channel that will notify
// when there's a more current template.
func (bt *BlockTemplateState) BlockTemplateOrLongPollChan(longPollID string,
	payAddress util.Address) (*appmessage.GetBlockTemplateResponseMessage, chan struct{}, error) {

	bt.Lock()
	defer bt.Unlock()

	if err := bt.Update(payAddress); err != nil {
		return nil, nil, err
	}

	// Just return the current block template if the long poll ID provided by
	// the caller is invalid.
	parentHashes, lastGenerated, err := bt.decodeLongPollID(longPollID)
	if err != nil {
		result, err := bt.Response()
		if err != nil {
			return nil, nil, err
		}

		return result, nil, nil
	}

	// Return the block template now if the specific block template
	// identified by the long poll ID no longer matches the current block
	// template as this means the provided template is stale.
	areHashesEqual := daghash.AreEqual(bt.template.Block.Header.ParentHashes, parentHashes)
	if !areHashesEqual ||
		lastGenerated != bt.lastGenerated.UnixSeconds() {

		// Include whether or not it is valid to submit work against the
		// old block template depending on whether or not a solution has
		// already been found and added to the block DAG.
		result, err := bt.Response()
		if err != nil {
			return nil, nil, err
		}

		return result, nil, nil
	}

	// Register the parent hashes and last generated time for notifications
	// Get a channel that will be notified when the template associated with
	// the provided ID is stale and a new block template should be returned to
	// the caller.
	longPollChan := bt.templateUpdateChan(parentHashes, lastGenerated)
	return nil, longPollChan, nil
}

// templateUpdateChan returns a channel that will be closed once the block
// template associated with the passed parent hashes and last generated time
// is stale. The function will return existing channels for duplicate
// parameters which allows multiple clients to wait for the same block template
// without requiring a different channel for each client.
//
// This function MUST be called with the state locked.
func (bt *BlockTemplateState) templateUpdateChan(tipHashes []*daghash.Hash, lastGenerated int64) chan struct{} {
	tipHashesStr := daghash.JoinHashesStrings(tipHashes, "")
	// Either get the current list of channels waiting for updates about
	// changes to block template for the parent hashes or create a new one.
	channels, ok := bt.notifyMap[tipHashesStr]
	if !ok {
		m := make(map[int64]chan struct{})
		bt.notifyMap[tipHashesStr] = m
		channels = m
	}

	// Get the current channel associated with the time the block template
	// was last generated or create a new one.
	c, ok := channels[lastGenerated]
	if !ok {
		c = make(chan struct{})
		channels[lastGenerated] = c
	}

	return c
}

// encodeLongPollID encodes the passed details into an ID that can be used to
// uniquely identify a block template.
func (bt *BlockTemplateState) encodeLongPollID(parentHashes []*daghash.Hash, miningAddress util.Address, lastGenerated mstime.Time) string {
	return fmt.Sprintf("%s-%s-%d", daghash.JoinHashesStrings(parentHashes, ""), miningAddress, lastGenerated.UnixSeconds())
}

// decodeLongPollID decodes an ID that is used to uniquely identify a block
// template. This is mainly used as a mechanism to track when to update clients
// that are using long polling for block templates. The ID consists of the
// parent blocks hashes for the associated template and the time the associated
// template was generated.
func (bt *BlockTemplateState) decodeLongPollID(longPollID string) ([]*daghash.Hash, int64, error) {
	fields := strings.Split(longPollID, "-")
	if len(fields) != 2 {
		return nil, 0, errors.New("decodeLongPollID: invalid number of fields")
	}

	parentHashesStr := fields[0]
	if len(parentHashesStr)%daghash.HashSize != 0 {
		return nil, 0, errors.New("decodeLongPollID: invalid parent hashes format")
	}
	numberOfHashes := len(parentHashesStr) / daghash.HashSize

	parentHashes := make([]*daghash.Hash, 0, numberOfHashes)

	for i := 0; i < len(parentHashesStr); i += daghash.HashSize {
		hash, err := daghash.NewHashFromStr(parentHashesStr[i : i+daghash.HashSize])
		if err != nil {
			return nil, 0, errors.Errorf("decodeLongPollID: NewHashFromStr: %s", err)
		}
		parentHashes = append(parentHashes, hash)
	}

	lastGenerated, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return nil, 0, errors.Errorf("decodeLongPollID: Cannot parse timestamp %s: %s", fields[1], err)
	}

	return parentHashes, lastGenerated, nil
}
