package blockdag

import (
	"fmt"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/domain/blocknode"
	"github.com/kaspanet/kaspad/util/mstime"
)

// newBlockNode returns a new block node for the given block header and parents, and the
// anticone of its selected parent (parent with highest blue score).
// selectedParentAnticone is used to update reachability data we store for future reachability queries.
// This function is NOT safe for concurrent access.
func (dag *BlockDAG) newBlockNode(blockHeader *appmessage.BlockHeader, parents blocknode.Set) (node *blocknode.Node, selectedParentAnticone []*blocknode.Node) {
	node = blocknode.NewNode(blockHeader, parents, dag.Now().UnixMilliseconds())

	if len(parents) == 0 {
		// The genesis block is defined to have a blueScore of 0
		node.BlueScore = 0
		return node, nil
	}

	selectedParentAnticone, err := dag.ghostdag(node)
	if err != nil {
		panic(errors.Wrap(err, "unexpected error in GHOSTDAG"))
	}
	return node, selectedParentAnticone
}

func (dag *BlockDAG) isViolatingFinality(node *blocknode.Node) (bool, error) {
	if node.IsGenesis() {
		return false, nil
	}

	if dag.virtual.Node.Less(node) {
		isVirtualFinalityPointInNodesSelectedChain, err := dag.isInSelectedParentChainOf(
			dag.finalityPoint(dag.virtual.Node), node.SelectedParent) // use node.selectedParent because node still doesn't have reachability data
		if err != nil {
			return false, err
		}
		if !isVirtualFinalityPointInNodesSelectedChain {
			return true, nil
		}
	}

	return false, nil
}

func (dag *BlockDAG) hasValidChildren(node *blocknode.Node) bool {
	for child := range node.Children {
		if dag.Index.BlockNodeStatus(child) == blocknode.StatusValid {
			return true
		}
	}
	return false
}

func (dag *BlockDAG) checkBoundedMergeDepth(node *blocknode.Node) error {
	nonBoundedMergeDepthViolatingBlues, err := dag.nonBoundedMergeDepthViolatingBlues(node)
	if err != nil {
		return err
	}

	finalityPoint := dag.finalityPoint(node)
	for _, red := range node.Reds {
		doesRedHaveFinalityPointInPast, err := dag.isInPast(finalityPoint, red)
		if err != nil {
			return err
		}

		isRedInPastOfAnyNonFinalityViolatingBlue, err := dag.isInPastOfAny(red, nonBoundedMergeDepthViolatingBlues)
		if err != nil {
			return err
		}

		if !doesRedHaveFinalityPointInPast && !isRedInPastOfAnyNonFinalityViolatingBlue {
			return ruleError(ErrViolatingBoundedMergeDepth, "block is violating bounded merge depth")
		}
	}

	return nil
}

func (dag *BlockDAG) finalityPoint(node *blocknode.Node) *blocknode.Node {
	return node.BlockAtDepth(dag.FinalityInterval())
}

func (dag *BlockDAG) hasFinalityPointInOthersSelectedChain(node *blocknode.Node, other *blocknode.Node) (bool, error) {
	finalityPoint := dag.finalityPoint(node)
	return dag.isInSelectedParentChainOf(finalityPoint, other)
}

func (dag *BlockDAG) nonBoundedMergeDepthViolatingBlues(node *blocknode.Node) (blocknode.Set, error) {
	nonBoundedMergeDepthViolatingBlues := blocknode.NewSet()

	for _, blueNode := range node.Blues {
		notViolatingFinality, err := dag.hasFinalityPointInOthersSelectedChain(node, blueNode)
		if err != nil {
			return nil, err
		}
		if notViolatingFinality {
			nonBoundedMergeDepthViolatingBlues.Add(blueNode)
		}
	}

	return nonBoundedMergeDepthViolatingBlues, nil
}

func (dag *BlockDAG) checkMergeSizeLimit(node *blocknode.Node) error {
	mergeSetSize := len(node.Reds) + len(node.Blues)

	if mergeSetSize > mergeSetSizeLimit {
		return ruleError(ErrViolatingMergeLimit,
			fmt.Sprintf("The block merges %d blocks > %d merge set size limit", mergeSetSize, mergeSetSizeLimit))
	}

	return nil
}

func (dag *BlockDAG) finalityScore(node *blocknode.Node) uint64 {
	return node.BlueScore / dag.FinalityInterval()
}

// CalcPastMedianTime returns the median time of the previous few blocks
// prior to, and including, the block node.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) PastMedianTime(node *blocknode.Node) mstime.Time {
	window := blueBlockWindow(node, 2*dag.TimestampDeviationTolerance-1)
	medianTimestamp, err := window.medianTimestamp()
	if err != nil {
		panic(fmt.Sprintf("blueBlockWindow: %s", err))
	}
	return mstime.UnixMilliseconds(medianTimestamp)
}

func (dag *BlockDAG) selectedParentMedianTime(node *blocknode.Node) mstime.Time {
	medianTime := node.Header().Timestamp
	if !node.IsGenesis() {
		medianTime = dag.PastMedianTime(node.SelectedParent)
	}
	return medianTime
}

func (dag *BlockDAG) addNodeToIndexWithInvalidAncestor(block *util.Block) error {
	blockHeader := &block.MsgBlock().Header
	newNode, _ := dag.newBlockNode(blockHeader, blocknode.NewSet())
	newNode.Status = blocknode.StatusInvalidAncestor
	dag.Index.AddNode(newNode)

	dbTx, err := dag.DatabaseContext.NewTx()
	if err != nil {
		return err
	}
	defer dbTx.RollbackUnlessClosed()
	err = dag.Index.FlushToDB(dbTx)
	if err != nil {
		return err
	}
	return dbTx.Commit()
}

func lookupParentNodes(block *util.Block, dag *BlockDAG) (blocknode.Set, error) {
	header := block.MsgBlock().Header
	parentHashes := header.ParentHashes

	nodes := blocknode.NewSet()
	for _, parentHash := range parentHashes {
		node, ok := dag.Index.LookupNode(parentHash)
		if !ok {
			str := fmt.Sprintf("parent block %s is unknown", parentHash)
			return nil, ruleError(ErrParentBlockUnknown, str)
		} else if dag.Index.BlockNodeStatus(node).KnownInvalid() {
			str := fmt.Sprintf("parent block %s is known to be invalid", parentHash)
			return nil, ruleError(ErrInvalidAncestorBlock, str)
		}

		nodes.Add(node)
	}

	return nodes, nil
}
