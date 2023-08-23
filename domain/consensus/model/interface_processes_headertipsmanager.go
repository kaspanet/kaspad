package model

import "github.com/c4ei/yunseokyeol/domain/consensus/model/externalapi"

// HeadersSelectedTipManager manages the state of the headers selected tip
type HeadersSelectedTipManager interface {
	AddHeaderTip(stagingArea *StagingArea, hash *externalapi.DomainHash) error
}
