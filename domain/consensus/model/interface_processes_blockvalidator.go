package model

import "github.com/kaspanet/kaspad/app/appmessage"

// BlockValidator exposes a set of validation classes, after which
// it's possible to determine whether a block is valid
type BlockValidator interface {
	ValidateHeaderInIsolation(block *appmessage.MsgBlock) error
	ValidateBodyInIsolation(block *appmessage.MsgBlock) error
	ValidateHeaderInContext(block *appmessage.MsgBlock) error
	ValidateBodyInContext(block *appmessage.MsgBlock) error
	ValidateAgainstPastUTXO(block *appmessage.MsgBlock) error
	ValidateFinality(block *appmessage.MsgBlock) error
}
