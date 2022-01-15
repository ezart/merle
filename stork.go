package merle

import (
	"log"
)

// Stork is an interface to something which delivers things
type Storker interface {
	// Deliver a new thing.  The model is the thing model.  Demo is tells
	// the thing to run in demo-mode.  In demo-mode, the thing runs with
	// simulated or synthetized data, not relying on an external source for
	// the data.
	NewThinger(log *log.Logger, model string, demo bool) (Thinger, error)
	// Return a list of models support by stork
	Models() ([]string)
}
