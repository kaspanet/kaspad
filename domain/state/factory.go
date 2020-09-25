package state

type Factory interface {
	NewState() State
}

type factory struct {
}

func (f *factory) NewState() State {
	return &state{}
}

func NewFactory() Factory {
	return &factory{}
}
