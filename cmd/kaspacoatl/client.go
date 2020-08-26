package main

type client struct{}

func connectToServer(cfg *configFlags) (*client, error) {
	return &client{}, nil
}

func (c *client) disconnect() {

}
