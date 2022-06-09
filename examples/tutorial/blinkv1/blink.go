// file: examples/tutorial/blinkv1/blink.go

package main

import (
	"github.com/merliot/merle"
	"log"
)

type blink struct {
}

func (b *blink) Subscribers() merle.Subscribers {
	return merle.Subscribers{}
}

func main() {
	thing := merle.NewThing(&blink{})
	log.Fatalln(thing.Run())
}
