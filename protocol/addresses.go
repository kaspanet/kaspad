package protocol

import "github.com/kaspanet/kaspad/addrmgr"

// AddressManager returns the address manager associated with the manager.
func (m *Manager) AddressManager() *addrmgr.AddrManager {
	return m.addressManager
}
