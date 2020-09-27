package blocktemplatebuilder

import "github.com/kaspanet/kaspad/app/appmessage"

type BlockTemplateBuilder interface {
	GetBlockTemplate() *appmessage.MsgBlock
}
