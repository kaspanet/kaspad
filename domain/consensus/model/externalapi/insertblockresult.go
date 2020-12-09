package externalapi

type BlockInsertionResult struct {
	SelectedParentChainChanges *SelectedParentChainChanges
}

type SelectedParentChainChanges struct {
	Added   []*DomainHash
	Removed []*DomainHash
}
