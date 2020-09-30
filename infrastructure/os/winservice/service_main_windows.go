package winservice

import (
	"github.com/btcsuite/winsvc/svc"
	"github.com/kaspanet/kaspad/infrastructure/config"
)

// serviceMain checks whether we're being invoked as a service, and if so uses
// the service control manager to start the long-running server. A flag is
// returned to the caller so the application can determine whether to exit (when
// running as a service) or launch in normal interactive mode.
func serviceMain(main MainFunc, description *ServiceDescription, cfg *config.Config) (bool, error) {
	service := newService(main, description, cfg)

	if cfg.ServiceOptions.ServiceCommand != "" {
		return true, service.performServiceCommand()
	}

	// Don't run as a service if we're running interactively (or that can't
	// be determined due to an error).
	isInteractive, err := svc.IsAnInteractiveSession()
	if err != nil {
		return false, err
	}
	if isInteractive {
		return false, nil
	}

	err = service.Start()
	if err != nil {
		return true, err
	}

	return true, nil
}

// Set windows specific functions to real functions.
func init() {
	WinServiceMain = serviceMain
}
