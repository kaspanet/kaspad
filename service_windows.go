// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"

	"github.com/btcsuite/winsvc/eventlog"
	"github.com/btcsuite/winsvc/mgr"
	"github.com/btcsuite/winsvc/svc"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/os/signal"
	"github.com/kaspanet/kaspad/version"
)

const (
	// svcName is the name of kaspad service.
	svcName = "kaspadsvc"

	// svcDisplayName is the service name that will be shown in the windows
	// services list. Not the svcName is the "real" name which is used
	// to control the service. This is only for display purposes.
	svcDisplayName = "Kaspad Service"

	// svcDesc is the description of the service.
	svcDesc = "Downloads and stays synchronized with the Kaspa block " +
		"DAG and provides DAG services to applications."
)

// elog is used to send messages to the Windows event log.
var elog *eventlog.Log

// logServiceStart logs information about kaspad when the main server has
// been started to the Windows event log.
func logServiceStart() {
	var message string
	message += fmt.Sprintf("Version %s\n", version.Version())
	message += fmt.Sprintf("Configuration directory: %s\n", config.DefaultHomeDir)
	message += fmt.Sprintf("Configuration file: %s\n", cfg.ConfigFile)
	message += fmt.Sprintf("Data directory: %s\n", cfg.DataDir)

	elog.Info(1, message)
}

// kaspadService houses the main service handler which handles all service
// updates and launching kaspadMain.
type kaspadService struct{}

// Execute is the main entry point the winsvc package calls when receiving
// information from the Windows service control manager. It launches the
// long-running kaspadMain (which is the real meat of kaspad), handles service
// change requests, and notifies the service control manager of changes.
func (s *kaspadService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (bool, uint32) {
	// Service start is pending.
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}

	// Start kaspadMain in a separate goroutine so the service can start
	// quickly. Shutdown (along with a potential error) is reported via
	// doneChan. startedChan is notified once kaspad is started so this can
	// be properly logged
	doneChan := make(chan error)
	startedChan := make(chan struct{})
	spawn("kaspadMain-windows", func() {
		err := kaspadMain(startedChan)
		doneChan <- err
	})

	// Service is now started.
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus

			case svc.Stop, svc.Shutdown:
				// Service stop is pending. Don't accept any
				// more commands while pending.
				changes <- svc.Status{State: svc.StopPending}

				// Signal the main function to exit.
				signal.ShutdownRequestChannel <- struct{}{}

			default:
				elog.Error(1, fmt.Sprintf("Unexpected control "+
					"request #%d.", c))
			}

		case <-startedChan:
			logServiceStart()

		case err := <-doneChan:
			if err != nil {
				elog.Error(1, err.Error())
			}
			break loop
		}
	}

	// Service is now stopped.
	changes <- svc.Status{State: svc.Stopped}
	return false, 0
}

// installService attempts to install the kaspad service. Typically this should
// be done by the msi installer, but it is provided here since it can be useful
// for development.
func installService() error {
	// Get the path of the current executable. This is needed because
	// os.Args[0] can vary depending on how the application was launched.
	// For example, under cmd.exe it will only be the name of the app
	// without the path or extension, but under mingw it will be the full
	// path including the extension.
	exePath, err := filepath.Abs(os.Args[0])
	if err != nil {
		return err
	}
	if filepath.Ext(exePath) == "" {
		exePath += ".exe"
	}

	// Connect to the windows service manager.
	serviceManager, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer serviceManager.Disconnect()

	// Ensure the service doesn't already exist.
	service, err := serviceManager.OpenService(svcName)
	if err == nil {
		service.Close()
		return errors.Errorf("service %s already exists", svcName)
	}

	// Install the service.
	service, err = serviceManager.CreateService(svcName, exePath, mgr.Config{
		DisplayName: svcDisplayName,
		Description: svcDesc,
	})
	if err != nil {
		return err
	}
	defer service.Close()

	// Support events to the event log using the standard "standard" Windows
	// EventCreate.exe message file. This allows easy logging of custom
	// messges instead of needing to create our own message catalog.
	eventlog.Remove(svcName)
	eventsSupported := uint32(eventlog.Error | eventlog.Warning | eventlog.Info)
	return eventlog.InstallAsEventCreate(svcName, eventsSupported)
}

// removeService attempts to uninstall the kaspad service. Typically this should
// be done by the msi uninstaller, but it is provided here since it can be
// useful for development. Not the eventlog entry is intentionally not removed
// since it would invalidate any existing event log messages.
func removeService() error {
	// Connect to the windows service manager.
	serviceManager, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer serviceManager.Disconnect()

	// Ensure the service exists.
	service, err := serviceManager.OpenService(svcName)
	if err != nil {
		return errors.Errorf("service %s is not installed", svcName)
	}
	defer service.Close()

	// Remove the service.
	return service.Delete()
}

// startService attempts to Start the kaspad service.
func startService() error {
	// Connect to the windows service manager.
	serviceManager, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer serviceManager.Disconnect()

	service, err := serviceManager.OpenService(svcName)
	if err != nil {
		return errors.Errorf("could not access service: %s", err)
	}
	defer service.Close()

	err = service.Start(os.Args)
	if err != nil {
		return errors.Errorf("could not start service: %s", err)
	}

	return nil
}

// controlService allows commands which change the status of the service. It
// also waits for up to 10 seconds for the service to change to the passed
// state.
func controlService(c svc.Cmd, to svc.State) error {
	// Connect to the windows service manager.
	serviceManager, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer serviceManager.Disconnect()

	service, err := serviceManager.OpenService(svcName)
	if err != nil {
		return errors.Errorf("could not access service: %s", err)
	}
	defer service.Close()

	status, err := service.Control(c)
	if err != nil {
		return errors.Errorf("could not send control=%d: %s", c, err)
	}

	// Send the control message.
	timeout := time.Now().Add(10 * time.Second)
	for status.State != to {
		if timeout.Before(time.Now()) {
			return errors.Errorf("timeout waiting for service to go "+
				"to state=%d", to)
		}
		time.Sleep(300 * time.Millisecond)
		status, err = service.Query()
		if err != nil {
			return errors.Errorf("could not retrieve service "+
				"status: %s", err)
		}
	}

	return nil
}

// performServiceCommand attempts to run one of the supported service commands
// provided on the command line via the service command flag. An appropriate
// error is returned if an invalid command is specified.
func performServiceCommand(command string) error {
	var err error
	switch command {
	case "install":
		err = installService()

	case "remove":
		err = removeService()

	case "start":
		err = startService()

	case "stop":
		err = controlService(svc.Stop, svc.Stopped)

	default:
		err = errors.Errorf("invalid service command [%s]", command)
	}

	return err
}

// serviceMain checks whether we're being invoked as a service, and if so uses
// the service control manager to start the long-running server. A flag is
// returned to the caller so the application can determine whether to exit (when
// running as a service) or launch in normal interactive mode.
func serviceMain() (bool, error) {
	// Don't run as a service if we're running interactively (or that can't
	// be determined due to an error).
	isInteractive, err := svc.IsAnInteractiveSession()
	if err != nil {
		return false, err
	}
	if isInteractive {
		return false, nil
	}

	elog, err = eventlog.Open(svcName)
	if err != nil {
		return false, err
	}
	defer elog.Close()

	err = svc.Run(svcName, &kaspadService{})
	if err != nil {
		elog.Error(1, fmt.Sprintf("Service start failed: %s", err))
		return true, err
	}

	return true, nil
}

// Set windows specific functions to real functions.
func init() {
	config.RunServiceCommand = performServiceCommand
	winServiceMain = serviceMain
}
