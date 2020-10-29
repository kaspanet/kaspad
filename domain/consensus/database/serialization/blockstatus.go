package serialization

import "github.com/kaspanet/kaspad/domain/consensus/model"

// DomainBlockStatusToDbBlockStatus converts model.BlockStatus to DbBlockStatus
func DomainBlockStatusToDbBlockStatus(domainBlockStatus model.BlockStatus) *DbBlockStatus {
	return &DbBlockStatus{
		Status: uint32(domainBlockStatus),
	}
}

// DbBlockStatusToDomainBlockStatus converts DbBlockStatus to model.BlockStatus
func DbBlockStatusToDomainBlockStatus(dbBlockStatus *DbBlockStatus) model.BlockStatus {
	return model.BlockStatus(dbBlockStatus.Status)
}
