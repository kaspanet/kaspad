package externalapi

// ConsensusEvent is an interface type that is implemented by all events raised by consensus
type ConsensusEvent interface {
	isConsensusEvent()
}

// BlockAdded is an event raised by consensus when a block was added to the dag
type BlockAdded struct {
	Block *DomainBlock
}

func (*BlockAdded) isConsensusEvent() {}

// VirtualChangeSet is an event raised by consensus when virtual changes
type VirtualChangeSet struct {
	VirtualSelectedParentChainChanges *SelectedChainPath
	VirtualUTXODiff                   UTXODiff
	VirtualParents                    []*DomainHash
	VirtualSelectedParentBlueScore    uint64
	VirtualDAAScore                   uint64
}

func (*VirtualChangeSet) isConsensusEvent() {}

// SelectedChainPath is a path the of the selected chains between two blocks.
type SelectedChainPath struct {
	Added   []*DomainHash
	Removed []*DomainHash
}
