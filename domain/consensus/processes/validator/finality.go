package validator

import "github.com/kaspanet/kaspad/domain/consensus/model"

// ValidateFinality makes sure the block does not violate finality
func (v *validator) ValidateFinality(block *model.DomainBlock) error {
	// genesis block can't violate finality
	if len(block.Header.ParentHashes) == 0 {
		return nil
	}

	if node.dag.virtual.less(node) {
		isVirtualFinalityPointInNodesSelectedChain, err := node.dag.isInSelectedParentChainOf(
			node.dag.virtual.finalityPoint(), node.selectedParent) // use node.selectedParent because node still doesn't have reachability data
		if err != nil {
			return false, err
		}
		if !isVirtualFinalityPointInNodesSelectedChain {
			return true, nil
		}
	}

	return false, nil
}
