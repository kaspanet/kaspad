package flowcontext

import "github.com/kaspanet/kaspad/addrmgr"

// AddressManager returns the address manager associated to the flow context.
func (f *FlowContext) AddressManager() *addrmgr.AddrManager {
	return f.addressManager
}
