package miningmanager

type Factory interface {
	NewMiningManager() MiningManager
}

type factory struct{}

func (f *factory) NewMiningManager() MiningManager {
	return nil
}

func NewFactory() Factory {
	return &factory{}
}
