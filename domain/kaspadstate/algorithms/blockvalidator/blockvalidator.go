package blockvalidator

import "github.com/kaspanet/kaspad/app/appmessage"

type BlockValidator interface {
	ValidateHeaderInIsolation(block *appmessage.MsgBlock) error
	ValidateHeaderInContext(block *appmessage.MsgBlock) error
	ValidateBodyInIsolation(block *appmessage.MsgBlock) error
	ValidateBodyInContext(block *appmessage.MsgBlock) error
}
