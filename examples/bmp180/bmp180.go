// file: examples/bmp180/bmp180.go

package main

import (
	"github.com/merliot/merle"
	"gobot.io/x/gobot/drivers/i2c"
	"gobot.io/x/gobot/platforms/raspi"
	"log"
	"sync"
	"time"
)

type bmp180 struct {
	sync.RWMutex
	driver *i2c.BMP180Driver
	lastTemperature float32
	lastPressure float32
}

type msg struct {
	Msg         string
	Temperature float32
	Pressure    float32
}

func (b *bmp180) init(p *merle.Packet) {
	adaptor := raspi.NewAdaptor()
	adaptor.Connect()
	b.driver = i2c.NewBMP180Driver(adaptor)
	b.driver.Start()
}

func (b *bmp180) run(p *merle.Packet) {
	msg := msg{Msg: "Update"}

	for {
		msg.Temperature, _ = b.driver.Temperature()
		msg.Pressure, _ = b.driver.Pressure()

		b.Lock()
		if msg.Temperature != b.lastTemperature ||
		   msg.Pressure != b.lastPressure {
			b.lastTemperature = msg.Temperature
			b.lastPressure = msg.Pressure
			p.Marshal(&msg).Broadcast()
		}
		b.Unlock()

		time.Sleep(time.Second)
	}
}

func (b *bmp180) getState(p *merle.Packet) {
	b.RLock()
	defer b.RUnlock()
	msg := &msg{
		Msg: merle.ReplyState,
		Pressure: b.lastPressure,
		Temperature: b.lastTemperature,
	}
	p.Marshal(&msg).Reply()
}

func (b *bmp180) saveState(p *merle.Packet) {
	b.Lock()
	defer b.Unlock()
	var msg msg
	p.Unmarshal(&msg)
	b.lastPressure = msg.Pressure
	b.lastTemperature = msg.Temperature
}

func (b *bmp180) update(p *merle.Packet) {
	b.saveState(p)
	p.Broadcast()
}

func (b *bmp180) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdInit:    b.init,
		merle.CmdRun:     b.run,
		merle.GetState:   b.getState,
		merle.ReplyState: b.saveState,
		"Update":         b.update,
	}
}

const html = `
<!DOCTYPE html>
<html lang="en">
	<head>
		<meta name="viewport" content="width=device-width, initial-scale=1">
	</head>
	<body>
		<canvas id="temp_gauge"></canvas>
		<canvas id="pres_gauge"></canvas>

		<script src="//cdn.rawgit.com/Mikhus/canvas-gauges/gh-pages/download/2.1.7/radial/gauge.min.js"></script>

		<script>
			var conn
			var online = false

			temp_gauge = document.getElementById("temp_gauge")
			pres_gauge = document.getElementById("pres_gauge")

			var tempGauge = new RadialGauge({renderTo: temp_gauge})
			var presGauge = new RadialGauge({renderTo: pres_gauge})

			function getState() {
				conn.send(JSON.stringify({Msg: "_GetState"}))
			}

			function getIdentity() {
				conn.send(JSON.stringify({Msg: "_GetIdentity"}))
			}

			function save(msg) {
				tempGauge.value = (msg.Temperature * 1.8) + 32.0
				presGauge.value = msg.Pressure / 1000.0
			}

			function show() {
				tempGauge.draw()
				presGauge.draw()
			}

			function connect() {
				conn = new WebSocket("{{.WebSocket}}")

				conn.onopen = function(evt) {
					getIdentity()
				}

				conn.onclose = function(evt) {
					online = false
					show()
					setTimeout(connect, 1000)
				}

				conn.onerror = function(err) {
					conn.close()
				}

				conn.onmessage = function(evt) {
					msg = JSON.parse(evt.data)
					console.log('bmp180', msg)

					switch(msg.Msg) {
					case "_ReplyIdentity":
					case "_EventStatus":
						online = msg.Online
						getState()
						break
					case "_ReplyState":
					case "Update":
						save(msg)
						show()
						break
					}
				}
			}

			connect()
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
