package main

import "github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/server"

func startDaemon(conf *startDaemonConfig) error {
	return server.Start(conf.NetParams(), conf.Listen, conf.RPCServer, conf.KeysFile, conf.PublicKey)
}
