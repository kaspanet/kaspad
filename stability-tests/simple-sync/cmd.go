package main

import (
	"os/exec"
	"syscall"
)

func killWithSigterm(cmd *exec.Cmd, commandName string) {
	err := cmd.Process.Signal(syscall.SIGTERM)
	if err != nil {
		log.Criticalf("error sending SIGKILL to %s", commandName)
	}
}
