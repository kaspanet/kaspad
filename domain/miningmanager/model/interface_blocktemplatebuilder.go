package model

import (
	consensusmodel "github.com/kaspanet/kaspad/domain/consensus/model"
)

// BlockTemplateBuilder builds block templates for miners to consume
type BlockTemplateBuilder interface {
	GetBlockTemplate(payAddress DomainAddress, extraData []byte) *consensusmodel.DomainBlock
}
