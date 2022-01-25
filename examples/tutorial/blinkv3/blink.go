// file: examples/tutorial/blinkv3/blink.go

package main

import (
	"github.com/scottfeldman/merle"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
	"time"
)

type blink struct {
}

func (b *blink) run(p *merle.Packet) {
	update := struct {
		Msg   string
		State bool
	}{Msg: "update"}

	adaptor := raspi.NewAdaptor()
	adaptor.Connect()

	led := gpio.NewLedDriver(adaptor, "11")
	led.Start()

	for {
		led.Toggle()

		update.State = led.State()
		p.Marshal(&update).Broadcast()

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

			conn = new WebSocket("ws://localhost:8080/ws/{{.Id}}")

			conn.onmessage = function(evt) {
				msg = JSON.parse(evt.data)
				console.log('msg', msg)

				switch(msg.Msg) {
				case "update":
					image.src = "/{{.Id}}/assets/images/led-" +
						msg.State + ".png"
					break
				}
			}
		</script>
	</body>
</html>`

func (b *blink) Assets() *merle.ThingAssets {
	return &merle.ThingAssets{
		Dir: "examples/tutorial/blinkv3/assets",
		TemplateText: html,
	}
}

func main() {
	var cfg merle.ThingConfig

	cfg.Thing.PortPublic = 8080

	merle.NewThing(&blink{}, &cfg).Run()
}
