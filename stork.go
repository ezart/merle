package merle

import (
	"log"
)

type Storker interface {
	NewThinger(*log.Logger, string, bool) (Thinger, error)
}
