package pruningmanager

type PruningManager interface {
	UpdatePruningPointAndPruneIfRequired()
}
