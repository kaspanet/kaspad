package model

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/util/daghash"
)

// OnBlockAddedToDAGHandler is a handler function for when a block
// has been successfully added to the DAG
type OnBlockAddedToDAGHandler func(block *appmessage.MsgBlock)

// OnChainChangedHandler is a handler function for when the virtual
// block's selected parent chain had changed
type OnChainChangedHandler func(removedChainBlockHashes []*daghash.Hash, addedChainBlockHashes []*daghash.Hash)

// OnFinalityConflictHandler is a handler function for when a
// conflict in finality occurs
type OnFinalityConflictHandler func(violatingBlockHash *daghash.Hash)

// OnFinalityConflictResolvedHandler is a handler function for when
// an existing finality conflict has been resolved
type OnFinalityConflictResolvedHandler func(finalityBlockHash *daghash.Hash)
