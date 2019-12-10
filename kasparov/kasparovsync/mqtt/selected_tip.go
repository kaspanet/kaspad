package mqtt

import "github.com/kaspanet/kaspad/kasparov/kasparovserver/controllers"

const selectedTipTopic = "dag/selected-tip"

// PublishSelectedTipNotification publishes notification for a new selected tip
func PublishSelectedTipNotification(selectedTipHash string) error {
	if !isConnected() {
		return nil
	}
	block, err := controllers.GetBlockByHashHandler(selectedTipHash)
	if err != nil {
		return err
	}
	return publish(selectedTipTopic, block)
}
