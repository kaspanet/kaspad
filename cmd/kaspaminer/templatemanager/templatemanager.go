package templatemanager

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"sync"
)

var currentTemplate *externalapi.DomainBlock
var isSynced bool
var lock = &sync.Mutex{}

// Get returns the template to work on
func Get() (*externalapi.DomainBlock, bool) {
	lock.Lock()
	defer lock.Unlock()
	// Shallow copy the block so when the user replaces the header it won't affect the template here.
	if currentTemplate == nil {
		return nil, false
	}
	block := *currentTemplate
	return &block, isSynced
}

// Set sets the current template to work on
func Set(template *appmessage.GetBlockTemplateResponseMessage) error {
	block, err := appmessage.RPCBlockToDomainBlock(template.Block)
	if err != nil {
		return err
	}
	lock.Lock()
	defer lock.Unlock()
	currentTemplate = block
	isSynced = template.IsSynced
	return nil
}
