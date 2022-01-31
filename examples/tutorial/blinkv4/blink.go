// file: examples/tutorial/blinkv3/blink.go

package main

import (
	"flag"
	"github.com/scottfeldman/merle"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
	"sync"
	"time"
)

type blink struct {
	sync.Mutex
	led   *gpio.LedDriver
	state bool
}

type msg struct {
	Msg   string
	State bool
}

func (b *blink) run(p *merle.Packet) {
	msg := &msg{Msg: "Update"}

	adaptor := raspi.NewAdaptor()
	adaptor.Connect()

	b.led = gpio.NewLedDriver(adaptor, "11")
	b.led.Start()

	for {
		b.Lock()
		b.led.Toggle()
		b.state = b.led.State()
		msg.State = b.state
		p.Marshal(&msg).Broadcast()
		b.Unlock()

		time.Sleep(time.Second)
	}
}

func (b *blink) getState(p *merle.Packet) {
	b.Lock()
	defer b.Unlock()
	msg := &msg{Msg: merle.ReplyState, State: b.state}
	p.Marshal(&msg).Reply()
}

func (b *blink) saveState(p *merle.Packet) {
	b.Lock()
	defer b.Unlock()
	var msg msg
	p.Unmarshal(&msg)
	b.state = msg.State
}

func (b *blink) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdRun: b.run,
		merle.GetState: b.getState,
		merle.ReplyState: b.saveState,
		"Update": merle.Broadcast,
	}
}

const html = `<html lang="en">
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
				console.log('msg', msg)

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
		Dir: "examples/tutorial/blinkv4/assets",
		TemplateText: html,
	}
}

func main() {
	var cfg merle.ThingConfig

	prime := flag.Bool("prime", false, "Run Thing-prime")
	flag.Parse()

	cfg.Thing.PortPublic = 80
	cfg.Thing.PortPrivate = 8080
	cfg.Thing.User = "admin"

	if *prime {
		cfg.Thing.Prime = true
		cfg.Thing.PortPrime = 8000
		cfg.Thing.PortPublic = 90
		cfg.Thing.PortPrivate = 9080
//		cfg.Thing.PortPublicTLS = 443
	} else {
		cfg.Mother.Host = "localhost"
		cfg.Mother.User = "admin"
		cfg.Mother.Key = "/home/admin/.ssh/id_rsa"
		cfg.Mother.PortPrivate = 9080
	}

	merle.NewThing(&blink{}, &cfg).Run()
}
