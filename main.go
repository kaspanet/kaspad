// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	_ "net/http/pprof"
	"os"

	"github.com/kaspanet/kaspad/app"
)

func main() {
	if err := app.StartApp(); err != nil {
		os.Exit(1)
	}
}
