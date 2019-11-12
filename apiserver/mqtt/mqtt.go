package mqtt

import (
	"errors"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var client mqtt.Client

func Client() (mqtt.Client, error) {
	if client == nil {
		return nil, errors.New("MQTT is not connected")
	}
	return client, nil
}
