package protocol

import "github.com/kaspanet/kaspad/addrmgr"

func (m *Manager) AddressManager() *addrmgr.AddrManager {
	return m.addressManager
}
