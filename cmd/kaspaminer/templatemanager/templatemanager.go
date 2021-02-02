package templatemanager

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"sync/atomic"
	"unsafe"
)

var protectedTemplate *appmessage.GetBlockTemplateResponseMessage

// Get returns the template to work on
func Get() *appmessage.GetBlockTemplateResponseMessage {
	return (*appmessage.GetBlockTemplateResponseMessage)(
		atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&protectedTemplate))))
}

// Set sets the current template to work on
func Set(template *appmessage.GetBlockTemplateResponseMessage) {
	atomic.StorePointer(
		(*unsafe.Pointer)(unsafe.Pointer(&protectedTemplate)),
		unsafe.Pointer(template),
	)
}
