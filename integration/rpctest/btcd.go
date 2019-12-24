// Copyright (c) 2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package rpctest

import (
	"github.com/pkg/errors"
	"go/build"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
)

var (
	// compileMtx guards access to the executable path so that the project is
	// only compiled once.
	compileMtx sync.Mutex

	// executablePath is the path to the compiled executable. This is the empty
	// string until kaspad is compiled. This should not be accessed directly;
	// instead use the function kaspadExecutablePath().
	executablePath string
)

// kaspadExecutablePath returns a path to the kaspad executable to be used by
// rpctests. To ensure the code tests against the most up-to-date version of
// kaspad, this method compiles kaspad the first time it is called. After that, the
// generated binary is used for subsequent test harnesses. The executable file
// is not cleaned up, but since it lives at a static path in a temp directory,
// it is not a big deal.
func kaspadExecutablePath() (string, error) {
	compileMtx.Lock()
	defer compileMtx.Unlock()

	// If kaspad has already been compiled, just use that.
	if len(executablePath) != 0 {
		return executablePath, nil
	}

	testDir, err := baseDir()
	if err != nil {
		return "", err
	}

	// Determine import path of this package. Not necessarily kaspanet/kaspad if
	// this is a forked repo.
	_, rpctestDir, _, ok := runtime.Caller(1)
	if !ok {
		return "", errors.Errorf("Cannot get path to kaspad source code")
	}
	kaspadPkgPath := filepath.Join(rpctestDir, "..", "..", "..")
	kaspadPkg, err := build.ImportDir(kaspadPkgPath, build.FindOnly)
	if err != nil {
		return "", errors.Errorf("Failed to build kaspad: %s", err)
	}

	// Build kaspad and output an executable in a static temp path.
	outputPath := filepath.Join(testDir, "kaspad")
	if runtime.GOOS == "windows" {
		outputPath += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", outputPath, kaspadPkg.ImportPath)
	err = cmd.Run()
	if err != nil {
		return "", errors.Errorf("Failed to build kaspad: %s", err)
	}

	// Save executable path so future calls do not recompile.
	executablePath = outputPath
	return executablePath, nil
}
