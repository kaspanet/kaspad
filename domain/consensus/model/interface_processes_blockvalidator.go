package model

// BlockValidator exposes a set of validation classes, after which
// it's possible to determine whether a block is valid
type BlockValidator interface {
	ValidateHeaderInIsolation(block *DomainBlockHeader) error
	ValidateBodyInIsolation(block *DomainBlock) error
	ValidateHeaderInContext(block *DomainBlockHeader) error
	ValidateBodyInContext(block *DomainBlock) error
	ValidateAgainstPastUTXO(block *DomainBlock) error
	ValidateFinality(block *DomainBlock) error
}
