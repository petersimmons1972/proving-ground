package orchestrator

import (
	"io"
	"log"
	"os"
)

func captureLog(f *os.File) func() {
	prev := log.Writer()
	log.SetOutput(io.MultiWriter(prev, f))
	return func() { log.SetOutput(prev) }
}
