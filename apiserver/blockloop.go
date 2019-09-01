package main

func blockLoop(client *apiServerClient, doneChan chan struct{}) error {
	baka, err := client.GetChainFromBlock(true, nil)
	if err != nil {
		return nil
	}
	log.Warnf("aaaa, %+v", baka)

loop:
	for {
		select {
		case blockAdded := <-client.onBlockAdded:
			log.Infof("blockAdded: %s", blockAdded.header)
		case chainChanged := <-client.onChainChanged:
			log.Infof("chainChanged: %+v", chainChanged)
		case <-doneChan:
			log.Infof("blockLoop stopped")
			break loop
		}
	}
	return nil
}
