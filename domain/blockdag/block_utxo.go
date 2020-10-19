package blockdag

import (
	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/domain/blocknode"
	"github.com/kaspanet/kaspad/domain/utxo"
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
	_, _, txsAcceptanceData, err := dag.pastUTXO(dag.virtual.Node)
	return txsAcceptanceData, err
}

// TxsAcceptedByBlockHash retrieves transactions accepted by the given block
//
// This function MUST be called with the DAG read-lock held
func (dag *BlockDAG) TxsAcceptedByBlockHash(blockHash *daghash.Hash) (MultiBlockTxsAcceptanceData, error) {
	node, ok := dag.Index.LookupNode(blockHash)
	if !ok {
		return nil, errors.Errorf("Couldn't find block %s", blockHash)
	}
	_, _, txsAcceptanceData, err := dag.pastUTXO(node)
	return txsAcceptanceData, err
}

func (dag *BlockDAG) meldVirtualUTXO(newVirtualUTXODiffSet *utxo.DiffUTXOSet) error {
	return newVirtualUTXODiffSet.MeldToBase()
}

type utxoVerificationOutput struct {
	newBlockPastUTXO  utxo.Set
	txsAcceptanceData MultiBlockTxsAcceptanceData
	newBlockMultiset  *secp256k1.MultiSet
}

// verifyAndBuildUTXO verifies all transactions in the given block and builds its UTXO
// to save extra traversals it returns the transactions acceptance data
// for the new block and its multiset.
func (dag *BlockDAG) verifyAndBuildUTXO(node *blocknode.Node, transactions []*util.Tx) (*utxoVerificationOutput, error) {
	pastUTXO, selectedParentPastUTXO, txsAcceptanceData, err := dag.pastUTXO(node)
	if err != nil {
		return nil, err
	}

	err = dag.validateAcceptedIDMerkleRoot(node, txsAcceptanceData)
	if err != nil {
		return nil, err
	}

	err = dag.checkConnectBlockToPastUTXO(node, pastUTXO, transactions)
	if err != nil {
		return nil, err
	}

	multiset, err := dag.calcMultiset(node, txsAcceptanceData, selectedParentPastUTXO)
	if err != nil {
		return nil, err
	}

	err = validateUTXOCommitment(node, multiset)
	if err != nil {
		return nil, err
	}

	return &utxoVerificationOutput{
		newBlockPastUTXO:  pastUTXO,
		txsAcceptanceData: txsAcceptanceData,
		newBlockMultiset:  multiset}, nil
}

func genesisPastUTXO(virtual *virtualBlock) utxo.Set {
	// The genesis has no past UTXO, so we create an empty UTXO
	// set by creating a diff UTXO set with the virtual UTXO
	// set, and adding all of its entries in toRemove
	diff := utxo.NewDiff()
	for outpoint, entry := range virtual.utxoSet.UTXOCache {
		diff.ToRemove[outpoint] = entry
	}
	genesisPastUTXO := utxo.Set(utxo.NewDiffUTXOSet(virtual.utxoSet, diff))
	return genesisPastUTXO
}

// applyBlueBlocks adds all transactions in the blue blocks to the selectedParent's past UTXO set
// Purposefully ignoring failures - these are just unaccepted transactions
// Writing down which transactions were accepted or not in txsAcceptanceData
func (dag *BlockDAG) applyBlueBlocks(node *blocknode.Node, selectedParentPastUTXO utxo.Set, blueBlocks []*util.Block) (
	pastUTXO utxo.Set, multiBlockTxsAcceptanceData MultiBlockTxsAcceptanceData, err error) {

	pastUTXO = selectedParentPastUTXO.(*utxo.DiffUTXOSet).CloneWithoutBase()
	multiBlockTxsAcceptanceData = make(MultiBlockTxsAcceptanceData, len(blueBlocks))

	// We obtain the median time of the selected parent block (unless it's genesis block)
	// in order to determine if transactions in theF current block are final.
	selectedParentMedianTime := dag.selectedParentMedianTime(node)
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
				dag.maybeAcceptTx(node, tx, isSelectedParent, pastUTXO, accumulatedMass, selectedParentMedianTime)
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

func (dag *BlockDAG) maybeAcceptTx(node *blocknode.Node, tx *util.Tx, isSelectedParent bool, pastUTXO utxo.Set,
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

		_, err = pastUTXO.AddTx(tx.MsgTx(), node.BlueScore)
		if err != nil {
			return false, 0, 0, err
		}

		return true, 0, accumulatedMass, nil
	}

	txFee, accumulatedMassAfter, err = dag.checkConnectTransactionToPastUTXO(
		node, tx, pastUTXO, accumulatedMassBefore, selectedParentMedianTime)
	if err != nil {
		if !errors.As(err, &(RuleError{})) {
			return false, 0, 0, err
		}

		isAccepted = false
	} else {
		isAccepted = true
		accumulatedMass = accumulatedMassAfter

		_, err = pastUTXO.AddTx(tx.MsgTx(), node.BlueScore)
		if err != nil {
			return false, 0, 0, err
		}
	}
	return isAccepted, txFee, accumulatedMass, nil
}

// pastUTXO returns the UTXO of a given block's past
// To save traversals over the blue blocks, it also returns the transaction acceptance data for
// all blue blocks
func (dag *BlockDAG) pastUTXO(node *blocknode.Node) (
	pastUTXO, selectedParentPastUTXO utxo.Set, bluesTxsAcceptanceData MultiBlockTxsAcceptanceData, err error) {

	if node.IsGenesis() {
		return genesisPastUTXO(dag.virtual), nil, MultiBlockTxsAcceptanceData{}, nil
	}

	selectedParentPastUTXO, err = dag.restorePastUTXO(node.SelectedParent)
	if err != nil {
		return nil, nil, nil, err
	}

	blueBlocks, err := dag.fetchBlueBlocks(node)
	if err != nil {
		return nil, nil, nil, err
	}

	pastUTXO, bluesTxsAcceptanceData, err = dag.applyBlueBlocks(node, selectedParentPastUTXO, blueBlocks)
	if err != nil {
		return nil, nil, nil, err
	}

	return pastUTXO, selectedParentPastUTXO, bluesTxsAcceptanceData, nil
}

// restorePastUTXO restores the UTXO of a given block from its diff
func (dag *BlockDAG) restorePastUTXO(node *blocknode.Node) (utxo.Set, error) {
	stack := []*blocknode.Node{}

	// Iterate over the chain of diff-childs from node till virtual and add them
	// all into a stack
	for current := node; current != nil; {
		stack = append(stack, current)
		var err error
		current, err = dag.UTXODiffStore.DiffChildByNode(current)
		if err != nil {
			return nil, err
		}
	}

	// Start with the top item in the stack, going over it top-to-bottom,
	// applying the UTXO-diff one-by-one.
	topNode, stack := stack[len(stack)-1], stack[:len(stack)-1] // Pop the top item in the stack
	topNodeDiff, err := dag.UTXODiffStore.DiffByNode(topNode)
	if err != nil {
		return nil, err
	}
	accumulatedDiff := topNodeDiff.Clone()

	for i := len(stack) - 1; i >= 0; i-- {
		diff, err := dag.UTXODiffStore.DiffByNode(stack[i])
		if err != nil {
			return nil, err
		}
		// Use withDiffInPlace, otherwise copying the diffs again and again create a polynomial overhead
		err = accumulatedDiff.WithDiffInPlace(diff)
		if err != nil {
			return nil, err
		}
	}

	return utxo.NewDiffUTXOSet(dag.virtual.utxoSet, accumulatedDiff), nil
}

// updateValidTipsUTXO builds and applies new diff UTXOs for all the DAG's valid tips
func updateValidTipsUTXO(dag *BlockDAG, virtualUTXO utxo.Set) error {
	for validTip := range dag.validTips {
		validTipPastUTXO, err := dag.restorePastUTXO(validTip)
		if err != nil {
			return err
		}
		diff, err := virtualUTXO.DiffFrom(validTipPastUTXO)
		if err != nil {
			return err
		}
		err = dag.UTXODiffStore.SetBlockDiff(validTip, diff)
		if err != nil {
			return err
		}
	}

	return nil
}

// updateParentsDiffs updates the diff of any parent whose DiffChild is this block
func (dag *BlockDAG) updateParentsDiffs(node *blocknode.Node, newBlockPastUTXO utxo.Set) error {
	for parent := range node.Parents {
		if dag.Index.BlockNodeStatus(parent) == blocknode.StatusUTXOPendingVerification {
			continue
		}

		diffChild, err := dag.UTXODiffStore.DiffChildByNode(parent)
		if err != nil {
			return err
		}
		if diffChild == nil {
			parentPastUTXO, err := dag.restorePastUTXO(parent)
			if err != nil {
				return err
			}
			err = dag.UTXODiffStore.SetBlockDiffChild(parent, node)
			if err != nil {
				return err
			}
			diff, err := newBlockPastUTXO.DiffFrom(parentPastUTXO)
			if err != nil {
				return err
			}
			err = dag.UTXODiffStore.SetBlockDiff(parent, diff)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (dag *BlockDAG) updateDiffAndDiffChild(node *blocknode.Node, newBlockPastUTXO utxo.Set) error {
	var diffChild *blocknode.Node
	for child := range node.Children {
		if dag.Index.BlockNodeStatus(child) == blocknode.StatusValid {
			diffChild = child
			break
		}
	}

	// If there's no diffChild, then virtual is the de-facto diffChild
	var diffChildUTXOSet utxo.Set = dag.virtual.utxoSet
	if diffChild != nil {
		var err error
		diffChildUTXOSet, err = dag.restorePastUTXO(diffChild)
		if err != nil {
			return err
		}
	}

	diffFromDiffChild, err := diffChildUTXOSet.DiffFrom(newBlockPastUTXO)
	if err != nil {
		return err
	}

	err = dag.UTXODiffStore.SetBlockDiff(node, diffFromDiffChild)
	if err != nil {
		return err
	}

	if diffChild != nil {
		err = dag.UTXODiffStore.SetBlockDiffChild(node, diffChild)
		if err != nil {
			return err
		}
	}
	return nil
}
