package main

import (
	"fmt"
	"os"

	"github.com/c4ei/yunseokyeol/infrastructure/config"
	"github.com/c4ei/yunseokyeol/infrastructure/network/netadapter/standalone"
)

func connectToNode() *standalone.Routes {
	cfg := activeConfig()

	c4exdConfig := config.DefaultConfig()
	c4exdConfig.NetworkFlags = cfg.NetworkFlags

	minimalNetAdapter, err := standalone.NewMinimalNetAdapter(c4exdConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating minimalNetAdapter: %+v", err)
		os.Exit(1)
	}
	routes, err := minimalNetAdapter.Connect(cfg.NodeP2PAddress)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error connecting to node: %+v", err)
		os.Exit(1)
	}
	return routes
}
