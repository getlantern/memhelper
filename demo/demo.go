package main

import (
	"time"

	"github.com/getlantern/golog"
	"github.com/getlantern/memhelper"
)

var (
	log = golog.LoggerFor("demo")

	retained [][]byte
)

func main() {
	log.Debug("This program will print memory stats a few times and then exit")
	memhelper.Track(1*time.Second, 1*time.Second)
	for i := 0; i < 200; i++ {
		retained = append(retained, make([]byte, 1024768))
		time.Sleep(10 * time.Millisecond)
	}
	// Clear retained to let memory get released
	retained = nil
	time.Sleep(15 * time.Second)
	log.Debug("Finished")
}
