package memhelper

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/getlantern/golog"
)

var (
	log = golog.LoggerFor("memhelper")

	mem atomic.Value // *runtime.MemStats

	runOnce sync.Once
)

// Track refreshes memory stats every refreshInterval and logs them every logPeriod.
func Track(refreshInterval time.Duration, logPeriod time.Duration) {
	runOnce.Do(func() {
		go trackMemStats(refreshInterval)
		go logMemStats(logPeriod)
	})
}

func setMem(_mem *runtime.MemStats) {
	mem.Store(_mem)
}

func getMem() *runtime.MemStats {
	_mem := mem.Load()
	if _mem == nil {
		return nil
	}
	return _mem.(*runtime.MemStats)
}

func trackMemStats(period time.Duration) {
	memstats := &runtime.MemStats{}
	for {
		runtime.ReadMemStats(memstats)
		setMem(memstats)
		time.Sleep(period)
	}
}

func logMemStats(period time.Duration) {
	for {
		memstats := getMem()
		if memstats != nil {
			log.Debugf("Memory InUse: %v    Alloc: %v    Sys: %v",
				humanize.Bytes(memstats.HeapInuse),
				humanize.Bytes(memstats.Alloc),
				humanize.Bytes(memstats.Sys),
			)
		}
		time.Sleep(period)
	}
}
