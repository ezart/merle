// file: examples/bmp180/bmp180.go

package main

import (
	"github.com/merliot/merle"
	"gobot.io/x/gobot/drivers/i2c"
	"gobot.io/x/gobot/platforms/raspi"
	"log"
	"time"
)

type bmp180 struct {
}

func (b *bmp180) run(p *merle.Packet) {
	update := struct {
		Msg         string
		Pressure    float32
		Temperature float32
	}{Msg: "update"}

	adaptor := raspi.NewAdaptor()
	adaptor.Connect()

	bmp180 := i2c.NewBMP180Driver(adaptor)
	bmp180.Start()

	for {
		update.Pressure, _ = bmp180.Pressure()
		update.Temperature, _ = bmp180.Temperature()
		p.Marshal(&update).Broadcast()
		time.Sleep(time.Second)
	}
}

func (b *bmp180) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdRun: b.run,
	}
}

const html = `
<!DOCTYPE html>
<html lang="en">
	<body>
		<pre id="pressure">Pressure:</pre>
		<pre id="temperature">Temperature:</pre>

		<script>
			pressure = document.getElementById("pressure")
			temperature = document.getElementById("temperature")

			conn = new WebSocket("{{.Scheme}}{{.Host}}/ws/{{.Id}}")

			conn.onmessage = function(evt) {
				msg = JSON.parse(evt.data)
				console.log('bmp180', msg)

				switch(msg.Msg) {
				case "update":
					pressure.textContent =
						"Pressure: " + msg.Pressure
					temperature.textContent =
						"Temperature: " + msg.Temperature
					break
				}
			}
		</script>
	</body>
</html>`

func (b *bmp180) Assets() *merle.ThingAssets {
	return &merle.ThingAssets{
		HtmlTemplateText: html,
	}
}

func main() {
	thing := merle.NewThing(&bmp180{})
	thing.Cfg.PortPublic = 80
	log.Fatalln(thing.Run())
}
