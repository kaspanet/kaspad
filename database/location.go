package database

type StoreLocation []byte

func (s StoreLocation) Serialize() []byte {
	serializedLocation := make([]byte, len(s))
	copy(serializedLocation, s)
	return serializedLocation
}

func (s *StoreLocation) Deserialize(serializedLocation []byte) {
	*s = make([]byte, len(serializedLocation))
	copy(*s, serializedLocation)
}
