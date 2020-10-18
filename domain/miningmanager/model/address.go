package model

// DomainAddress is the domain representation of a kaspad
// address
type DomainAddress interface {
	ScriptAddress() []byte
}
