package memhelper

import (
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/getlantern/golog"
	"github.com/shirou/gopsutil/process"
)

var (
	log = golog.LoggerFor("memhelper")

	memory        uint64
	logMemStatsCh = make(chan *memoryInfo)

	runOnce sync.Once
)

// Track refreshes memory stats every refreshInterval and logs them every logPeriod.
func Track(refreshInterval time.Duration, logPeriod time.Duration) {
	runOnce.Do(func() {
		go trackMemStats(refreshInterval, logPeriod)
	})
}

type memoryInfo struct {
	mi       *process.MemoryInfoStat
	memstats *runtime.MemStats
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
	atomic.StoreUint64(&memory, memstats.Alloc)
	mem := &memoryInfo{
		mi:       mi,
		memstats: memstats,
	}
	select {
	case logMemStatsCh <- mem:
		// will get logged
	default:
		// won't get logged because we're busy
	}
}

// log the most recent available memstats every 10 seconds
func logMemStats(period time.Duration) {
	t := time.NewTicker(period)
	defer t.Stop()

	var mem *memoryInfo
	var more bool
	for {
		select {
		case mem, more = <-logMemStatsCh:
			if !more {
				return
			}
		case <-t.C:
			mi := mem.mi
			memstats := mem.memstats
			log.Debugf("Memory InUse: %v    Alloc: %v    Sys: %v     RSS: %v",
				humanize.Bytes(memstats.HeapInuse),
				humanize.Bytes(memstats.Alloc),
				humanize.Bytes(memstats.Sys),
				humanize.Bytes(mi.RSS))
		}
	}
}
