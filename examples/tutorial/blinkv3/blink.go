// file: examples/tutorial/blinkv3/blink.go

package main

import (
	"github.com/merliot/merle"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
	"log"
	"time"
)

type blink struct {
}

type msg struct {
	Msg   string
	State bool
}

func (b *blink) run(p *merle.Packet) {
	msg := &msg{Msg: "Update"}

	adaptor := raspi.NewAdaptor()
	adaptor.Connect()

	led := gpio.NewLedDriver(adaptor, "11")
	led.Start()

	for {
		led.Toggle()

		msg.State = led.State()
		p.Marshal(&msg).Broadcast()

		time.Sleep(time.Second)
	}
}

func (b *blink) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdRun: b.run,
	}
}

const html = `<html lang="en">
	<body>
		<img id="LED" style="width: 400px">

		<script>
			image = document.getElementById("LED")

			conn = new WebSocket("{{.WebSocket}}")

			conn.onmessage = function(evt) {
				msg = JSON.parse(evt.data)
				console.log('msg', msg)

				switch(msg.Msg) {
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
		Dir:          "examples/tutorial/blinkv3/assets",
		TemplateText: html,
	}
}

func main() {
	thing := merle.NewThing(&blink{})
	log.Fatalln(thing.Run())
}
