package serialization

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// HeaderTipsToDBHeaderTips converts a slice of hashes to DbHeaderTips
func HeaderTipsToDBHeaderTips(tips []*externalapi.DomainHash) *DbHeaderTips {
	return &DbHeaderTips{
		Tips: DomainHashesToDbHashes(tips),
	}
}

// DBHeaderTipsTOHeaderTips converts DbHeaderTips to a slice of hashes
func DBHeaderTipsTOHeaderTips(dbHeaderTips *DbHeaderTips) ([]*externalapi.DomainHash, error) {
	return DbHashesToDomainHashes(dbHeaderTips.Tips)
}
