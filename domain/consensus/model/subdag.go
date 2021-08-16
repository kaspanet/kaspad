package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// SubDAG represents a context-free representation of a partial DAG
type SubDAG struct {
	RootHashes []*externalapi.DomainHash
	TipHashes  []*externalapi.DomainHash
	Blocks     map[externalapi.DomainHash]*SubDAGBlock
}

// SubDAGBlock represents a block in a SubDAG
type SubDAGBlock struct {
	BlockHash    *externalapi.DomainHash
	ParentHashes []*externalapi.DomainHash
	ChildHashes  []*externalapi.DomainHash
}
