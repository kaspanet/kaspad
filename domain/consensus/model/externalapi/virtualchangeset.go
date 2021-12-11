package externalapi

// VirtualChangeSet is auxiliary data returned from ValidateAndInsertBlock and ResolveVirtual
type VirtualChangeSet struct {
	VirtualSelectedParentChainChanges *SelectedChainPath
	VirtualUTXODiff                   UTXODiff
	VirtualParents                    []*DomainHash
}

// SelectedChainPath is a path the of the selected chains between two blocks.
type SelectedChainPath struct {
	Added   []*DomainHash
	Removed []*DomainHash
}
