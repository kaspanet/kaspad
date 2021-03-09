package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/stability-tests/common"
	"github.com/pkg/errors"
)

type commandFailure struct {
	cmd *exec.Cmd
	err error
}

func (cf commandFailure) String() string {
	return fmt.Sprintf("command `%s` failed: %s", cf.cmd, cf.err)
}

func commandLoop(argsChan <-chan []string) ([]commandFailure, error) {
	failures := make([]commandFailure, 0)
	dataDirectoryPath, err := common.TempDir("kaspadsanity-kaspad-datadir")
	if err != nil {
		return nil, errors.Wrapf(err, "error creating temp dir")
	}
	defer os.RemoveAll(dataDirectoryPath)

	for args := range argsChan {
		err := os.RemoveAll(dataDirectoryPath)
		if err != nil {
			return nil, err
		}

		args, err = handleDataDirArg(args, dataDirectoryPath)
		if err != nil {
			return nil, err
		}

		cmd := exec.Command("kaspad", args...)
		cmd.Stdout = common.NewLogWriter(log, logger.LevelTrace, "KASPAD-STDOUT")
		cmd.Stderr = common.NewLogWriter(log, logger.LevelWarn, "KASPAD-STDERR")

		log.Infof("Running `%s`", cmd)
		errChan := make(chan error)
		spawn("commandLoop-cmd.Run", func() {
			errChan <- cmd.Run()
		})

		const timeout = time.Minute
		select {
		case err := <-errChan:
			failure := commandFailure{
				cmd: cmd,
				err: err,
			}
			log.Error(failure)
			failures = append(failures, failure)
		case <-time.After(timeout):
			err := cmd.Process.Kill()
			if err != nil {
				return nil, errors.Wrapf(err, "error in Kill")
			}
			log.Infof("Successfully run `%s`", cmd)
		}
	}
	return failures, nil
}

func handleDataDirArg(args []string, dataDir string) ([]string, error) {
	for _, arg := range args {
		if strings.HasPrefix(arg, "--datadir") {
			return nil, errors.New("invalid argument --datadir")
		}
	}
	return append([]string{"--datadir", dataDir}, args...), nil
}
