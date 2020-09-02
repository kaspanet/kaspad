package blockdag

import (
	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/pkg/errors"
)

// TxAcceptanceData stores a transaction together with an indication
// if it was accepted or not by some block
type TxAcceptanceData struct {
	Tx         *util.Tx
	Fee        uint64
	IsAccepted bool
}

// BlockTxsAcceptanceData stores all transactions in a block with an indication
// if they were accepted or not by some other block
type BlockTxsAcceptanceData struct {
	BlockHash        daghash.Hash
	TxAcceptanceData []TxAcceptanceData
}

// MultiBlockTxsAcceptanceData stores data about which transactions were accepted by a block
// It's a slice of the block's blues block IDs and their transaction acceptance data
type MultiBlockTxsAcceptanceData []BlockTxsAcceptanceData

// FindAcceptanceData finds the BlockTxsAcceptanceData that matches blockHash
func (data MultiBlockTxsAcceptanceData) FindAcceptanceData(blockHash *daghash.Hash) (*BlockTxsAcceptanceData, bool) {
	for _, acceptanceData := range data {
		if acceptanceData.BlockHash.IsEqual(blockHash) {
			return &acceptanceData, true
		}
	}
	return nil, false
}

// TxsAcceptedByVirtual retrieves transactions accepted by the current virtual block
//
// This function MUST be called with the DAG read-lock held
func (dag *BlockDAG) TxsAcceptedByVirtual() (MultiBlockTxsAcceptanceData, error) {
	_, _, txsAcceptanceData, err := dag.pastUTXO(dag.virtual.blockNode)
	return txsAcceptanceData, err
}

// TxsAcceptedByBlockHash retrieves transactions accepted by the given block
//
// This function MUST be called with the DAG read-lock held
func (dag *BlockDAG) TxsAcceptedByBlockHash(blockHash *daghash.Hash) (MultiBlockTxsAcceptanceData, error) {
	node, ok := dag.index.LookupNode(blockHash)
	if !ok {
		return nil, errors.Errorf("Couldn't find block %s", blockHash)
	}
	_, _, txsAcceptanceData, err := dag.pastUTXO(node)
	return txsAcceptanceData, err
}

func (dag *BlockDAG) meldVirtualUTXO(newVirtualUTXODiffSet *DiffUTXOSet) error {
	return newVirtualUTXODiffSet.meldToBase()
}

type utxoVerificationOutput struct {
	newBlockPastUTXO  UTXOSet
	txsAcceptanceData MultiBlockTxsAcceptanceData
	newBlockMultiset  *secp256k1.MultiSet
}

// verifyAndBuildUTXO verifies all transactions in the given block and builds its UTXO
// to save extra traversals it returns the transactions acceptance data
// for the new block and its multiset.
func (node *blockNode) verifyAndBuildUTXO(transactions []*util.Tx) (*utxoVerificationOutput, error) {
	pastUTXO, selectedParentPastUTXO, txsAcceptanceData, err := node.dag.pastUTXO(node)
	if err != nil {
		return nil, err
	}

	err = node.validateAcceptedIDMerkleRoot(node.dag, txsAcceptanceData)
	if err != nil {
		return nil, err
	}

	err = node.dag.checkConnectBlockToPastUTXO(node, pastUTXO, transactions)
	if err != nil {
		return nil, err
	}

	multiset, err := node.calcMultiset(txsAcceptanceData, selectedParentPastUTXO)
	if err != nil {
		return nil, err
	}

	err = node.validateUTXOCommitment(multiset)
	if err != nil {
		return nil, err
	}

	return &utxoVerificationOutput{
		newBlockPastUTXO:  pastUTXO,
		txsAcceptanceData: txsAcceptanceData,
		newBlockMultiset:  multiset}, nil
}

func genesisPastUTXO(virtual *virtualBlock) UTXOSet {
	// The genesis has no past UTXO, so we create an empty UTXO
	// set by creating a diff UTXO set with the virtual UTXO
	// set, and adding all of its entries in toRemove
	diff := NewUTXODiff()
	for outpoint, entry := range virtual.utxoSet.utxoCollection {
		diff.toRemove[outpoint] = entry
	}
	genesisPastUTXO := UTXOSet(NewDiffUTXOSet(virtual.utxoSet, diff))
	return genesisPastUTXO
}

// applyBlueBlocks adds all transactions in the blue blocks to the selectedParent's past UTXO set
// Purposefully ignoring failures - these are just unaccepted transactions
// Writing down which transactions were accepted or not in txsAcceptanceData
func (node *blockNode) applyBlueBlocks(selectedParentPastUTXO UTXOSet, blueBlocks []*util.Block) (
	pastUTXO UTXOSet, multiBlockTxsAcceptanceData MultiBlockTxsAcceptanceData, err error) {

	pastUTXO = selectedParentPastUTXO.(*DiffUTXOSet).cloneWithoutBase()
	multiBlockTxsAcceptanceData = make(MultiBlockTxsAcceptanceData, len(blueBlocks))

	// We obtain the median time of the selected parent block (unless it's genesis block)
	// in order to determine if transactions in the current block are final.
	selectedParentMedianTime := node.selectedParentMedianTime()
	accumulatedMass := uint64(0)

	// Add blueBlocks to multiBlockTxsAcceptanceData in topological order. This
	// is so that anyone who iterates over it would process blocks (and transactions)
	// in their order of appearance in the DAG.
	for i := 0; i < len(blueBlocks); i++ {
		blueBlock := blueBlocks[i]
		transactions := blueBlock.Transactions()
		blockTxsAcceptanceData := BlockTxsAcceptanceData{
			BlockHash:        *blueBlock.Hash(),
			TxAcceptanceData: make([]TxAcceptanceData, len(transactions)),
		}
		isSelectedParent := i == 0

		for j, tx := range transactions {
			var isAccepted bool
			var txFee uint64

			isAccepted, txFee, accumulatedMass, err =
				node.maybeAcceptTx(tx, isSelectedParent, pastUTXO, accumulatedMass, selectedParentMedianTime)
			if err != nil {
				return nil, nil, err
			}

			blockTxsAcceptanceData.TxAcceptanceData[j] = TxAcceptanceData{
				Tx:         tx,
				Fee:        txFee,
				IsAccepted: isAccepted}
		}
		multiBlockTxsAcceptanceData[i] = blockTxsAcceptanceData
	}

	return pastUTXO, multiBlockTxsAcceptanceData, nil
}

func (node *blockNode) maybeAcceptTx(tx *util.Tx, isSelectedParent bool, pastUTXO UTXOSet,
	accumulatedMassBefore uint64, selectedParentMedianTime mstime.Time) (
	isAccepted bool, txFee uint64, accumulatedMassAfter uint64, err error) {

	accumulatedMass := accumulatedMassBefore

	// Coinbase transaction outputs are added to the UTXO-set only if they are in the selected parent chain.
	if tx.IsCoinBase() {
		if !isSelectedParent {
			return false, 0, 0, nil
		}
		txMass := CalcTxMass(tx, nil)
		accumulatedMass += txMass

		_, err = pastUTXO.AddTx(tx.MsgTx(), node.blueScore)
		if err != nil {
			return false, 0, 0, err
		}

		return true, 0, accumulatedMass, nil
	}

	txFee, accumulatedMassAfter, err = node.dag.checkConnectTransactionToPastUTXO(
		node, tx, pastUTXO, accumulatedMassBefore, selectedParentMedianTime)
	if err != nil {
		if !errors.As(err, &(RuleError{})) {
			return false, 0, 0, err
		}

		isAccepted = false
	} else {
		isAccepted = true
		accumulatedMass = accumulatedMassAfter

		_, err = pastUTXO.AddTx(tx.MsgTx(), node.blueScore)
		if err != nil {
			return false, 0, 0, err
		}
	}
	return isAccepted, txFee, accumulatedMass, nil
}

// pastUTXO returns the UTXO of a given block's past
// To save traversals over the blue blocks, it also returns the transaction acceptance data for
// all blue blocks
func (dag *BlockDAG) pastUTXO(node *blockNode) (
	pastUTXO, selectedParentPastUTXO UTXOSet, bluesTxsAcceptanceData MultiBlockTxsAcceptanceData, err error) {

	if node.isGenesis() {
		return genesisPastUTXO(dag.virtual), nil, MultiBlockTxsAcceptanceData{}, nil
	}

	selectedParentPastUTXO, err = dag.restorePastUTXO(node.selectedParent)
	if err != nil {
		return nil, nil, nil, err
	}

	blueBlocks, err := dag.fetchBlueBlocks(node)
	if err != nil {
		return nil, nil, nil, err
	}

	pastUTXO, bluesTxsAcceptanceData, err = node.applyBlueBlocks(selectedParentPastUTXO, blueBlocks)
	if err != nil {
		return nil, nil, nil, err
	}

	return pastUTXO, selectedParentPastUTXO, bluesTxsAcceptanceData, nil
}

// restorePastUTXO restores the UTXO of a given block from its diff
func (dag *BlockDAG) restorePastUTXO(node *blockNode) (UTXOSet, error) {
	stack := []*blockNode{}

	// Iterate over the chain of diff-childs from node till virtual and add them
	// all into a stack
	for current := node; current != nil; {
		stack = append(stack, current)
		var err error
		current, err = dag.utxoDiffStore.diffChildByNode(current)
		if err != nil {
			return nil, err
		}
	}

	// Start with the top item in the stack, going over it top-to-bottom,
	// applying the UTXO-diff one-by-one.
	topNode, stack := stack[len(stack)-1], stack[:len(stack)-1] // pop the top item in the stack
	topNodeDiff, err := dag.utxoDiffStore.diffByNode(topNode)
	if err != nil {
		return nil, err
	}
	accumulatedDiff := topNodeDiff.clone()

	for i := len(stack) - 1; i >= 0; i-- {
		diff, err := dag.utxoDiffStore.diffByNode(stack[i])
		if err != nil {
			return nil, err
		}
		// Use withDiffInPlace, otherwise copying the diffs again and again create a polynomial overhead
		err = accumulatedDiff.withDiffInPlace(diff)
		if err != nil {
			return nil, err
		}
	}

	return NewDiffUTXOSet(dag.virtual.utxoSet, accumulatedDiff), nil
}

// updateValidTipsUTXO builds and applies new diff UTXOs for all the DAG's valid tips
func updateValidTipsUTXO(dag *BlockDAG, virtualUTXO UTXOSet) error {
	for validTip := range dag.validTips {
		if dag.index.BlockNodeStatus(validTip) != statusValid {
			continue
		}

		validTipPastUTXO, err := dag.restorePastUTXO(validTip)
		if err != nil {
			return err
		}
		diff, err := virtualUTXO.diffFrom(validTipPastUTXO)
		if err != nil {
			return err
		}
		err = dag.utxoDiffStore.setBlockDiff(validTip, diff)
		if err != nil {
			return err
		}
	}

	return nil
}

// updateParentsDiffs updates the diff of any parent whose DiffChild is this block
func (node *blockNode) updateParentsDiffs(dag *BlockDAG, newBlockPastUTXO UTXOSet) error {
	for parent := range node.parents {
		if node.dag.index.BlockNodeStatus(parent) == statusUTXONotVerified {
			continue
		}

		diffChild, err := dag.utxoDiffStore.diffChildByNode(parent)
		if err != nil {
			return err
		}
		if diffChild == nil {
			parentPastUTXO, err := dag.restorePastUTXO(parent)
			if err != nil {
				return err
			}
			err = dag.utxoDiffStore.setBlockDiffChild(parent, node)
			if err != nil {
				return err
			}
			diff, err := newBlockPastUTXO.diffFrom(parentPastUTXO)
			if err != nil {
				return err
			}
			err = dag.utxoDiffStore.setBlockDiff(parent, diff)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (node *blockNode) updateDiffAndDiffChild(dag *BlockDAG, newBlockPastUTXO UTXOSet) error {
	var diffChild *blockNode
	for child := range node.children {
		if node.dag.index.BlockNodeStatus(child) == statusValid {
			diffChild = child
			break
		}
	}

	var diffChildUTXOSet UTXOSet = dag.virtual.utxoSet
	if diffChild != nil {
		var err error
		diffChildUTXOSet, err = dag.restorePastUTXO(diffChild)
		if err != nil {
			return err
		}
	}

	diffFromDiffChild, err := diffChildUTXOSet.diffFrom(newBlockPastUTXO)
	if err != nil {
		return err
	}

	err = dag.utxoDiffStore.setBlockDiff(node, diffFromDiffChild)
	if err != nil {
		return err
	}

	if diffChild != nil {
		err = dag.utxoDiffStore.setBlockDiffChild(node, diffChild)
		if err != nil {
			return err
		}
	}
	return nil
}
