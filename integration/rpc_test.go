package integration

import (
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/wire"

	rpcclient "github.com/kaspanet/kaspad/rpc/client"
)

type rpcClient struct {
	*rpcclient.Client
	onBlockAdded func(*wire.BlockHeader)
}

func newRPCClient(rpcAddress string) (*rpcClient, error) {
	client := &rpcClient{}
	notificationHandlers := &rpcclient.NotificationHandlers{
		OnFilteredBlockAdded: func(height uint64, header *wire.BlockHeader, txs []*util.Tx) {
			if client.onBlockAdded != nil {
				client.onBlockAdded(header)
			}
		},
	}

	connConfig := &rpcclient.ConnConfig{
		Host:           rpcAddress,
		Endpoint:       "ws",
		User:           rpcUser,
		Pass:           rpcPass,
		DisableTLS:     true,
		RequestTimeout: defaultTimeout,
	}

	var err error
	client.Client, err = rpcclient.New(connConfig, notificationHandlers)
	return client, err
}
