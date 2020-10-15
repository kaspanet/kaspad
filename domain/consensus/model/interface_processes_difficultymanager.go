package model

// DifficultyManager provides a method to resolve the
// difficulty value of a block
type DifficultyManager interface {
	RequiredDifficulty(parents []*DomainHash) uint32
}
