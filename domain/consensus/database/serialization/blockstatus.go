package serialization

import "github.com/kaspanet/kaspad/domain/consensus/model"

func DomainBlockStatusToDbBlockStatus(domainBlockStatus model.BlockStatus) *DbBlockStatus {
	return &DbBlockStatus{
		Status: uint32(domainBlockStatus),
	}
}

func DbBlockStatusToDomainBlockStatus(dbBlockStatus *DbBlockStatus) model.BlockStatus {
	return model.BlockStatus(dbBlockStatus.Status)
}
