package prefix

import "github.com/pkg/errors"

const (
	prefixZero byte = 0
	prefixOne  byte = 1
)

// Prefix is a database prefix that is used to manage more than one database at once.
type Prefix struct {
	value byte
}

// Serialize serializes the prefix into a byte slice
func (p *Prefix) Serialize() []byte {
	return []byte{p.value}
}

// Equal returns whether p equals to other
func (p *Prefix) Equal(other *Prefix) bool {
	return p.value == other.value
}

// Flip returns the opposite of the current prefix
func (p *Prefix) Flip() *Prefix {
	value := prefixZero
	if p.value == prefixZero {
		value = prefixOne
	}

	return &Prefix{value: value}
}

// Deserialize deserializes a prefix from a byte slice
func Deserialize(prefixBytes []byte) (*Prefix, error) {
	if len(prefixBytes) > 1 {
		return nil, errors.Errorf("invalid length %d for prefix", len(prefixBytes))
	}

	if prefixBytes[0] != prefixZero && prefixBytes[0] != prefixOne {
		return nil, errors.Errorf("invalid prefix %x", prefixBytes)
	}

	return &Prefix{value: prefixBytes[0]}, nil
}
