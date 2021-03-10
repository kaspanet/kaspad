package externalapi

// BlockInsertionResult is auxiliary data returned from ValidateAndInsertBlock
type BlockInsertionResult struct {
	VirtualSelectedParentChainChanges *SelectedChainPath
	VirtualUTXODiff                   UTXODiff
	VirtualParents                    []*DomainHash
}

// SelectedChainPath is a path the of the selected chains between two blocks.
type SelectedChainPath struct {
	Added   []*DomainHash
	Removed []*DomainHash
}
