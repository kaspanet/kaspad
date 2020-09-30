package blocktemplatebuilder

import "github.com/kaspanet/kaspad/app/appmessage"

// BlockTemplateBuilder ...
type BlockTemplateBuilder interface {
	GetBlockTemplate() *appmessage.MsgBlock
}
