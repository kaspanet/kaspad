package winservice

import (
	"os"
	"path/filepath"
	"time"

	"github.com/btcsuite/winsvc/eventlog"

	"github.com/btcsuite/winsvc/mgr"

	"github.com/btcsuite/winsvc/svc"
	"github.com/pkg/errors"
)

// performServiceCommand attempts to run one of the supported service commands
// provided on the command line via the service command flag. An appropriate
// error is returned if an invalid command is specified.
func (s *Service) performServiceCommand() error {
	var err error
	command := s.cfg.ServiceOptions.ServiceCommand
	switch command {
	case "install":
		err = s.installService()

	case "remove":
		err = s.removeService()

	case "start":
		err = s.startService()

	case "stop":
		err = s.controlService(svc.Stop, svc.Stopped)

	default:
		err = errors.Errorf("invalid service command [%s]", command)
	}

	return err
}

// installService attempts to install the kaspad service. Typically this should
// be done by the msi installer, but it is provided here since it can be useful
// for development.
func (s *Service) installService() error {
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
	service, err := serviceManager.OpenService(s.description.Name)
	if err == nil {
		service.Close()
		return errors.Errorf("service %s already exists", s.description.Name)
	}

	// Install the service.
	service, err = serviceManager.CreateService(s.description.Name, exePath, mgr.Config{
		DisplayName: s.description.DisplayName,
		Description: s.description.Description,
	})
	if err != nil {
		return err
	}
	defer service.Close()

	// Support events to the event log using the standard "standard" Windows
	// EventCreate.exe message file. This allows easy logging of custom
	// messges instead of needing to create our own message catalog.
	err = eventlog.Remove(s.description.Name)
	if err != nil {
		return err
	}
	eventsSupported := uint32(eventlog.Error | eventlog.Warning | eventlog.Info)
	return eventlog.InstallAsEventCreate(s.description.Name, eventsSupported)
}

// removeService attempts to uninstall the kaspad service. Typically this should
// be done by the msi uninstaller, but it is provided here since it can be
// useful for development. Not the eventlog entry is intentionally not removed
// since it would invalidate any existing event log messages.
func (s *Service) removeService() error {
	// Connect to the windows service manager.
	serviceManager, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer serviceManager.Disconnect()

	// Ensure the service exists.
	service, err := serviceManager.OpenService(s.description.Name)
	if err != nil {
		return errors.Errorf("service %s is not installed", s.description.Name)
	}
	defer service.Close()

	// Remove the service.
	return service.Delete()
}

// startService attempts to Start the kaspad service.
func (s *Service) startService() error {
	// Connect to the windows service manager.
	serviceManager, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer serviceManager.Disconnect()

	service, err := serviceManager.OpenService(s.description.Name)
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
func (s *Service) controlService(c svc.Cmd, to svc.State) error {
	// Connect to the windows service manager.
	serviceManager, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer serviceManager.Disconnect()

	service, err := serviceManager.OpenService(s.description.Name)
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
