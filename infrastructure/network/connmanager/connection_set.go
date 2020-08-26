package connmanager

import (
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
)

type connectionSet map[string]*netadapter.NetConnection

func (cs connectionSet) add(connection *netadapter.NetConnection) {
	cs[connection.Address()] = connection
}

func (cs connectionSet) remove(connection *netadapter.NetConnection) {
	delete(cs, connection.Address())
}

func (cs connectionSet) get(address string) (*netadapter.NetConnection, bool) {
	connection, ok := cs[address]
	return connection, ok
}

func convertToSet(connections []*netadapter.NetConnection) connectionSet {
	connSet := make(connectionSet, len(connections))

	for _, connection := range connections {
		connSet[connection.Address()] = connection
	}

	return connSet
}
