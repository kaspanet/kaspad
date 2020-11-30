package model

// PruningManager resolves and manages the current pruning point
type PruningManager interface {
	FindNextPruningPoint() error
}
