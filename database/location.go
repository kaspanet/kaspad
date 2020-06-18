package database

// StoreLocation represents a location of a data chunk located at a database store
type StoreLocation []byte

// Serialize returns a byte slice that represents the StoreLocation
func (s StoreLocation) Serialize() []byte {
	serializedLocation := make([]byte, len(s))
	copy(serializedLocation, s)
	return serializedLocation
}

// Deserialize deserializes the given byte slice into s.
func (s *StoreLocation) Deserialize(serializedLocation []byte) {
	*s = make([]byte, len(serializedLocation))
	copy(*s, serializedLocation)
}
