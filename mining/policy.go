// Copyright (c) 2014-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package mining

// Policy houses the policy (configuration parameters) which is used to control
// the generation of block templates.  See the documentation for
// NewBlockTemplate for more details on each of these parameters are used.
type Policy struct {
	// BlockMinSize is the minimum block size to be used when generating
	// a block template.
	BlockMinSize uint32

	// BlockMaxSize is the maximum block size to be used when generating a
	// block template.
	BlockMaxSize uint32
}
