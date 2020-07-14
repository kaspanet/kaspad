package connmanager

import "github.com/kaspanet/kaspad/netadapter/server"

type connectionSet map[string]server.Connection

func (cs connectionSet) add(connection server.Connection) {
	cs[connection.Address().String()] = connection
}

func (cs connectionSet) remove(connection server.Connection) {
	delete(cs, connection.Address().String())
}

func (cs connectionSet) get(address string) server.Connection {
	return cs[address]
}

func convertToSet(connections []server.Connection) connectionSet {
	connSet := make(map[string]server.Connection, len(connections))

	for _, connection := range connections {
		connSet[connection.Address().String()] = connection
	}

	return connSet
}
