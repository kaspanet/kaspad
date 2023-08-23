package model

import "github.com/c4ei/YunSeokYeol/domain/consensus/model/externalapi"

// ParentsManager lets is a wrapper above header parents that replaces empty parents with genesis when needed.
type ParentsManager interface {
	ParentsAtLevel(blockHeader externalapi.BlockHeader, level int) externalapi.BlockLevelParents
	Parents(blockHeader externalapi.BlockHeader) []externalapi.BlockLevelParents
}
