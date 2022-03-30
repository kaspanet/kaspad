package model

import (
	consensusexternalapi "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// BlockTemplateBuilder builds block templates for miners to consume
type BlockTemplateBuilder interface {
	GetBlockTemplate(coinbaseData *consensusexternalapi.DomainCoinbaseData) (*consensusexternalapi.DomainBlockTemplate, error)
	ModifyBlockTemplate(newCoinbaseData *consensusexternalapi.DomainCoinbaseData,
		blockTemplateToModify *consensusexternalapi.DomainBlockTemplate) (*consensusexternalapi.DomainBlockTemplate, error)
}
