package model

// BlockRelations represents a block's parent/child relations
type BlockRelations struct {
	Parents  []*DomainHash
	Children []*DomainHash
}
