// file: examples/tutorial/blinkv3/blink.go

package main

import (
	"github.com/merliot/merle"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
	"log"
	"sync"
	"time"
)

type blink struct {
	sync.RWMutex
	Msg   string
	State bool
}

func (b *blink) run(p *merle.Packet) {
	adaptor := raspi.NewAdaptor()
	adaptor.Connect()

	led := gpio.NewLedDriver(adaptor, "11")
	led.Start()

	b.Msg = merle.ReplyState

	for {
		led.Toggle()

		b.Lock()
		b.State = led.State()
		p.Marshal(b)
		b.Unlock()

		p.Broadcast()

		time.Sleep(time.Second)
	}
}

func (b *blink) getState(p *merle.Packet) {
	b.RLock()
	p.Marshal(b)
	b.RUnlock()
	p.Reply()
}

func (b *blink) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdInit:  nil,
		merle.CmdRun:   b.run,
		merle.GetState: b.getState,
	}
}

const html = `
<!DOCTYPE html>
<html lang="en">
	<body>
		<img id="LED" style="width: 400px">

		<script>
			image = document.getElementById("LED")

			conn = new WebSocket("{{.WebSocket}}")

			conn.onopen = function(evt) {
				conn.send(JSON.stringify({Msg: "_GetState"}))
			}

			conn.onmessage = function(evt) {
				msg = JSON.parse(evt.data)
				console.log('blink', msg)

				switch(msg.Msg) {
				case "_ReplyState":
					image.src = "/{{.AssetsDir}}/images/led-" +
						msg.State + ".png"
					break
				}
			}
		</script>
	</body>
</html>`

func (b *blink) Assets() *merle.ThingAssets {
	return &merle.ThingAssets{
		AssetsDir:        "examples/tutorial/blinkv4/assets",
		HtmlTemplateText: html,
	}
}

func main() {
	thing := merle.NewThing(&blink{})
	thing.Cfg.PortPublic = 80
	log.Fatalln(thing.Run())
}
