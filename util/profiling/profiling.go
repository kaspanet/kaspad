package profiling

import (
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	// Required for profiling
	_ "net/http/pprof"

	"github.com/kaspanet/kaspad/util/panics"
	"runtime"
	"runtime/pprof"
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

// TrackHeap tracks the size of the heap and dumps a profile if it passes a limit
func TrackHeap(appDir string, log *logger.Logger) {
	spawn := panics.GoroutineWrapperFunc(log)
	spawn("profiling.TrackHeap", func() {
		dumpFolder := filepath.Join(appDir, "dumps")
		err := os.MkdirAll(dumpFolder, 0700)
		if err != nil {
			log.Errorf("Could not create heap dumps folder at %s", dumpFolder)
			return
		}
		const limitInGigabytes = 8
		trackHeapSize(limitInGigabytes*1024*1024*1024, dumpFolder, log)
	})
}

func trackHeapSize(heapLimit uint64, dumpFolder string, log *logger.Logger) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		memStats := &runtime.MemStats{}
		runtime.ReadMemStats(memStats)
		// If we passed the expected heap limit, dump the heap profile to a file
		if memStats.HeapAlloc > heapLimit {
			dumpHeapProfile(heapLimit, dumpFolder, memStats, log)
		}
	}
}

func dumpHeapProfile(heapLimit uint64, dumpFolder string, memStats *runtime.MemStats, log *logger.Logger) {
	heapFile := filepath.Join(dumpFolder, "heap.out") // Should we keep a few recent files or override each time?
	log.Infof("Saving heap statistics into %s (HeapAlloc=%d > %d=heapLimit)", heapFile, memStats.HeapAlloc, heapLimit)
	f, err := os.Create(heapFile)
	defer f.Close()
	if err != nil {
		log.Infof("Could not create heap profile: %s", err)
		return
	}
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Infof("Could not write heap profile: %s", err)
	}
}
