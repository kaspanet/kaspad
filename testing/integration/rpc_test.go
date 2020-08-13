package integration

import (
	"github.com/kaspanet/kaspad/network/domainmessage"
	"github.com/kaspanet/kaspad/util"

	rpcclient "github.com/kaspanet/kaspad/network/rpc/client"
)

type rpcClient struct {
	*rpcclient.Client
	onBlockAdded func(*domainmessage.BlockHeader)
}

func newRPCClient(rpcAddress string) (*rpcClient, error) {
	client := &rpcClient{}
	notificationHandlers := &rpcclient.NotificationHandlers{
		OnFilteredBlockAdded: func(height uint64, header *domainmessage.BlockHeader, txs []*util.Tx) {
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
