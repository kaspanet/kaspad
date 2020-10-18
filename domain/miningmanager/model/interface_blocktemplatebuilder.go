package model

import (
	consensusmodel "github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/util"
)

// BlockTemplateBuilder builds block templates for miners to consume
type BlockTemplateBuilder interface {
	GetBlockTemplate(payAddress util.Address, extraData []byte) *consensusmodel.DomainBlock
}
