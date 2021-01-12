package externalapi

// BlockInsertionResult is auxiliary data returned from ValidateAndInsertBlock
type BlockInsertionResult struct {
	VirtualSelectedParentChainChanges *SelectedParentChainChanges
}

// SelectedParentChainChanges is the set of changes made to the selected parent chain
type SelectedParentChainChanges struct {
	Added   []*DomainHash
	Removed []*DomainHash
}
