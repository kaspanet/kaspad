package common

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/pkg/errors"
)

// StartCmd runs a command as a separate process.
// The `name` parameter is used for logs.
// The command executable should be in args[0]
func StartCmd(name string, args ...string) (*exec.Cmd, error) {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = NewLogWriter(log, logger.LevelTrace, fmt.Sprintf("%s-STDOUT", name))
	cmd.Stderr = NewLogWriter(log, logger.LevelWarn, fmt.Sprintf("%s-STDERR", name))
	log.Debugf("Starting command %s: %s", name, cmd)
	err := cmd.Start()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return cmd, nil
}

// NetworkCliArgumentFromNetParams returns the kaspad command line argument that starts the given network.
func NetworkCliArgumentFromNetParams(params *dagconfig.Params) string {
	return fmt.Sprintf("--%s", strings.TrimPrefix(params.Name, "kaspa-"))
}
