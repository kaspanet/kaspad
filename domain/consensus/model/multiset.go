package model

type Multiset interface {
	Add(data []byte)
	Remove(data []byte)
}
