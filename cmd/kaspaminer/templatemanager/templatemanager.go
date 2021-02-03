package templatemanager

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"sync"
)

var template *appmessage.GetBlockTemplateResponseMessage
var lock *sync.Mutex

// Get returns the template to work on
func Get() *appmessage.GetBlockTemplateResponseMessage {
	lock.Lock()
	defer lock.Unlock()
	return template
}

// Set sets the current template to work on
func Set(template *appmessage.GetBlockTemplateResponseMessage) {
	lock.Lock()
	defer lock.Unlock()
	template = template
}
