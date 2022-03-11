package model

import "github.com/pkg/errors"

// ErrBlockNotInSelectedParentChain is returned from CreateHeadersSelectedChainBlockLocator if one of the parameters
// passed to it are not in the headers selected parent chain
var ErrBlockNotInSelectedParentChain = errors.New("Block is not in selected parent chain")

// ErrReachedMaxTraversalAllowed is returned from AnticoneFromBlocks if `maxTraversalAllowed` was specified
// and the traversal passed it
var ErrReachedMaxTraversalAllowed = errors.New("Traversal searching for anticone passed the maxTraversalAllowed limit")
