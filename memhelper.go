package memhelper

import (
	"os"
	"runtime"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/getlantern/golog"
	"github.com/shirou/gopsutil/process"
)

var (
	log = golog.LoggerFor("memhelper")

	mem atomic.Value

	runOnce sync.Once
)

type memoryInfo struct {
	mi       *process.MemoryInfoStat
	memstats *runtime.MemStats
}

// Track refreshes memory stats every refreshInterval and logs them every logPeriod.
func Track(refreshInterval time.Duration, logPeriod time.Duration) {
	runOnce.Do(func() {
		go trackMemStats(refreshInterval, logPeriod)
	})
}

// TrackAndLimit tracks memory usage like Track and also tries to limit resident
// size (physical memory usage) to the given limitInBytes, applying the limit
// every limitPeriod.
func TrackAndLimit(refreshInterval time.Duration, logPeriod time.Duration, limitPeriod time.Duration, limitInBytes int) {
	runOnce.Do(func() {
		go trackMemStats(refreshInterval, logPeriod)
		go limitRSS(limitPeriod, uint64(limitInBytes))
	})
}

func setMem(_mem *memoryInfo) {
	mem.Store(_mem)
}

func getMem() *memoryInfo {
	_mem := mem.Load()
	if _mem == nil {
		return nil
	}
	return _mem.(*memoryInfo)
}

func trackMemStats(refreshInterval time.Duration, logPeriod time.Duration) {
	var logOnce sync.Once
	for {
		logOnce.Do(func() {
			go logMemStats(logPeriod)
		})
		updateMemStats()
		time.Sleep(refreshInterval)
	}
}

func updateMemStats() {
	p, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		log.Errorf("Unable to get process info: %v", err)
		return
	}
	mi, err := p.MemoryInfo()
	if err != nil {
		log.Errorf("Unable to get memory info for process: %v", err)
		return
	}
	memstats := &runtime.MemStats{}
	runtime.ReadMemStats(memstats)
	setMem(&memoryInfo{
		mi:       mi,
		memstats: memstats,
	})
}

func logMemStats(period time.Duration) {
	t := time.NewTicker(period)
	defer t.Stop()

	for range t.C {
		mem := getMem()
		if mem == nil {
			continue
		}
		mi := mem.mi
		memstats := mem.memstats
		log.Debugf("Memory InUse: %v    Alloc: %v    Sys: %v     RSS: %v",
			humanize.Bytes(memstats.HeapInuse),
			humanize.Bytes(memstats.Alloc),
			humanize.Bytes(memstats.Sys),
			humanize.Bytes(mi.RSS))
	}
}

func limitRSS(period time.Duration, limit uint64) {
	log.Debugf("Will attempt to limit RSS to %v", humanize.Bytes(limit))
	t := time.NewTicker(period)
	defer t.Stop()

	for range t.C {
		mem := getMem()
		if mem == nil {
			continue
		}
		if mem.mi.RSS > limit {
			log.Debugf("Resident size of %v exceeds limit of %v, attempting to free OS memory", humanize.Bytes(mem.mi.RSS), humanize.Bytes(limit))
			runtime.GC()
			debug.FreeOSMemory()
		}
	}
}
