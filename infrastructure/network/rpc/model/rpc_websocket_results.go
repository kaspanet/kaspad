// Copyright (c) 2015-2017 The btcsuite developers
// Copyright (c) 2015-2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package model

// SessionResult models the data from the session command.
type SessionResult struct {
	SessionID uint64 `json:"sessionId"`
}
