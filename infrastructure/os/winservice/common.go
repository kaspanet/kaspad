package winservice

import "github.com/kaspanet/kaspad/infrastructure/config"

type ServiceDescription struct {
	Name        string
	DisplayName string
	Description string
}

type MainFunc func(startedChan chan<- struct{}) error

// WinServiceMain is only invoked on Windows. It detects when kaspad is running
// as a service and reacts accordingly.
var WinServiceMain = func(MainFunc, *ServiceDescription, *config.Config) (bool, error) { return false, nil }
