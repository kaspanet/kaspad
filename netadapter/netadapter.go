package netadapter

type netAdapter struct {
	routerInitializer func() router
}

func newNetAdapter(routerInitializer func() router) *netAdapter {
	return &netAdapter{
		routerInitializer: routerInitializer,
	}
}
