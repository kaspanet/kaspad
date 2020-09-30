package executor

import (
	"fmt"
	"os"
	"runtime"

	"github.com/kaspanet/kaspad/infrastructure/os/limits"
)

func Initialize(desiredLimits *limits.DesiredLimits) {
	// Use all processor cores.
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Up some limits.
	if err := limits.SetLimits(desiredLimits); err != nil {
		fmt.Fprintf(os.Stderr, "failed to set limits: %s\n", err)
		os.Exit(1)
	}

}
