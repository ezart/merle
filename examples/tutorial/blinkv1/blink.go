// file: examples/tutorial/blinkv1/blink.go

package main

import (
	"github.com/scottfeldman/merle"
)

type blink struct {
}

func (b *blink) Subscribers() merle.Subscribers {
	return merle.Subscribers{}
}

func main() {
	var cfg merle.ThingConfig

	merle.NewThing(&blink{}, &cfg).Run()
}
