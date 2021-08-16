package model

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// GHOSTManager exposes a method to calculate the GHOST chain above
// some block
type GHOSTManager interface {
	GHOST(stagingArea *StagingArea, lowHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error)
}
