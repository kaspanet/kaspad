// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package winservice

import (
	"fmt"

	"github.com/btcsuite/winsvc/eventlog"
	"github.com/btcsuite/winsvc/svc"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/os/signal"
	"github.com/kaspanet/kaspad/version"
)

// Service houses the main service handler which handles all service
// updates and launching the application's main.
type Service struct {
	main        MainFunc
	description *ServiceDescription
	cfg         *config.Config
	eventLog    *eventlog.Log
}

func newService(main MainFunc, description *ServiceDescription, cfg *config.Config) *Service {
	return &Service{
		main:        main,
		description: description,
		cfg:         cfg,
	}
}

// Start starts the srevice
func (s *Service) Start() error {
	elog, err := eventlog.Open(s.description.Name)
	if err != nil {
		return err
	}
	s.eventLog = elog
	defer s.eventLog.Close()

	err = svc.Run(s.description.Name, &Service{})
	if err != nil {
		s.eventLog.Error(1, fmt.Sprintf("Service start failed: %s", err))
		return err
	}

	return nil
}

// Execute is the main entry point the winsvc package calls when receiving
// information from the Windows service control manager. It launches the
// long-running kaspadMain (which is the real meat of kaspad), handles service
// change requests, and notifies the service control manager of changes.
func (s *Service) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (bool, uint32) {
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
		err := s.main(startedChan)
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
				s.eventLog.Error(1, fmt.Sprintf("Unexpected control "+
					"request #%d.", c))
			}

		case <-startedChan:
			s.logServiceStart()

		case err := <-doneChan:
			if err != nil {
				s.eventLog.Error(1, err.Error())
			}
			break loop
		}
	}

	// Service is now stopped.
	changes <- svc.Status{State: svc.Stopped}
	return false, 0
}

// logServiceStart logs information about kaspad when the main server has
// been started to the Windows event log.
func (s *Service) logServiceStart() {
	var message string
	message += fmt.Sprintf("%s version %s\n", s.description.DisplayName, version.Version())
	message += fmt.Sprintf("Configuration file: %s\n", s.cfg.ConfigFile)
	message += fmt.Sprintf("Application directory: %s\n", s.cfg.AppDir)
	message += fmt.Sprintf("Logs directory: %s\n", s.cfg.LogDir)
}
