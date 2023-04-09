package externalapi

// BlockLocator is used to help locate a specific block. The algorithm for
// building the block locator is to add block hashes in reverse order on the
// block's selected parent chain until the desired stop block is reached.
// In order to keep the list of locator hashes to a reasonable number of entries,
// the step between each entry is doubled each loop iteration to exponentially
// decrease the number of hashes as a function of the distance from the block
// being located.
//
// For example, assume a selected parent chain with IDs as depicted below, and the
// stop block is genesis:
//
//	genesis -> 1 -> 2 -> ... -> 15 -> 16  -> 17  -> 18
//
// The block locator for block 17 would be the hashes of blocks:
//
//	[17 16 14 11 7 2 genesis]
type BlockLocator []*DomainHash

// Clone returns a clone of BlockLocator
func (locator BlockLocator) Clone() BlockLocator {
	return CloneHashes(locator)
}
