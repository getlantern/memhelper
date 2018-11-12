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
	for i := 0; i < 1000; i++ {
		retained = append(retained, make([]byte, 1024768))
		time.Sleep(10 * time.Millisecond)
	}
	log.Debug("Finished")
}
