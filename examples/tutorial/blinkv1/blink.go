// file: examples/tutorial/blinkv1/blink.go

package main

import (
	"github.com/scottfeldman/merle"
	"log"
)

type blink struct {
}

func (b *blink) Subscribers() merle.Subscribers {
	return merle.Subscribers{}
}

func main() {
	var cfg merle.ThingConfig

	thing := merle.NewThing(&blink{}, &cfg)
	log.Fatalln(thing.Run())
}
