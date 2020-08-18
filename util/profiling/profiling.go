package profiling

import (
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"net"
	"net/http"

	// Required for profiling
	_ "net/http/pprof"

	"github.com/kaspanet/kaspad/util/panics"
)

// Start starts the profiling server
func Start(port string, log *logger.Logger) {
	spawn := panics.GoroutineWrapperFunc(log)
	spawn("profiling.Start", func() {
		listenAddr := net.JoinHostPort("", port)
		log.Infof("Profile server listening on %s", listenAddr)
		profileRedirect := http.RedirectHandler("/debug/pprof", http.StatusSeeOther)
		http.Handle("/", profileRedirect)
		log.Error(http.ListenAndServe(listenAddr, nil))
	})
}
