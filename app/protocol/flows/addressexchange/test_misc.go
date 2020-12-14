package addressexchange

import (
	"math/rand"
	"net"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/config"
)

var (
	defaultCfg = config.DefaultConfig()
)

func generateIPForTest() net.IP {
	ip := make([]byte, 4)
	rand.Read(ip)
	return ip
}

func generateAddressesForTest(number int) []*appmessage.NetAddress {
	addresses := make([]*appmessage.NetAddress, number)
	for i := 0; i < number; i++ {
		addresses[i] = appmessage.NewNetAddressIPPort(
			generateIPForTest(),
			uint16(rand.Intn(65536)),
			appmessage.SFNodeNetwork)
	}
	return addresses
}
