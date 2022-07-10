// file: examples/tutorial/blinkv3/blink.go

package main

import (
	"flag"
	"github.com/merliot/merle"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
	"log"
	"sync"
	"time"
)

type blink struct {
	sync.Mutex
	Msg   string
	State bool
}

func (b *blink) run(p *merle.Packet) {
	adaptor := raspi.NewAdaptor()
	adaptor.Connect()

	led := gpio.NewLedDriver(adaptor, "11")
	led.Start()

	for {
		led.Toggle()

		b.Lock()
		b.Msg = "Update"
		b.State = !b.State
//		b.State = led.State()
		p.Marshal(b)
		b.Unlock()

		p.Broadcast()

		time.Sleep(time.Second)
	}
}

func (b *blink) getState(p *merle.Packet) {
	b.Lock()
	b.Msg = merle.ReplyState
	p.Marshal(b)
	b.Unlock()
	p.Reply()
}

func (b *blink) saveState(p *merle.Packet) {
	b.Lock()
	p.Unmarshal(b)
	b.Unlock()
}

func (b *blink) update(p *merle.Packet) {
	b.saveState(p)
	p.Broadcast()
}

func (b *blink) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdInit:    nil,
		merle.CmdRun:     b.run,
		merle.GetState:   b.getState,
		merle.ReplyState: b.saveState,
		"Update":         b.update,
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
				case "Update":
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
		AssetsDir:        "examples/tutorial/blinkv5/assets",
		HtmlTemplateText: html,
	}
}

func main() {
	thing := merle.NewThing(&blink{})

	thing.Cfg.Model = "blink"
	thing.Cfg.Name = "blinky"
	thing.Cfg.User = "merle"

	thing.Cfg.PortPublic = 80
	thing.Cfg.PortPrivate = 8080

	flag.StringVar(&thing.Cfg.MotherHost, "rhost", "", "Remote host")
	flag.StringVar(&thing.Cfg.MotherUser, "ruser", "merle", "Remote user")
	flag.BoolVar(&thing.Cfg.IsPrime, "prime", false, "Run as Thing Prime")
	flag.UintVar(&thing.Cfg.PortPublicTLS, "TLS", 0, "TLS port")
	flag.Parse()

	log.Fatalln(thing.Run())
}
