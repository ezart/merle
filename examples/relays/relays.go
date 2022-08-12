// file: examples/relays/relays.go

package relays

import (
	"sync"

	"github.com/merliot/merle"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
)

type Relays struct {
	sync.Mutex
	drivers [4]*gpio.RelayDriver
	Msg     string
	States  [4]bool
}

func NewRelays() merle.Thinger {
	return &Relays{}
}

func (r *Relays) run(p *merle.Packet) {
	adaptor := raspi.NewAdaptor()
	adaptor.Connect()

	r.drivers[0] = gpio.NewRelayDriver(adaptor, "31") // GPIO 6
	r.drivers[1] = gpio.NewRelayDriver(adaptor, "33") // GPIO 13
	r.drivers[2] = gpio.NewRelayDriver(adaptor, "35") // GPIO 19
	r.drivers[3] = gpio.NewRelayDriver(adaptor, "37") // GPIO 26

	for _, driver := range r.drivers {
		driver.Start()
		driver.Off()
	}

	select {}
}

func (r *Relays) getState(p *merle.Packet) {
	r.Lock()
	r.Msg = merle.ReplyState
	p.Marshal(r)
	r.Unlock()
	p.Reply()
}

func (r *Relays) saveState(p *merle.Packet) {
	r.Lock()
	p.Unmarshal(r)
	r.Unlock()
}

type MsgClick struct {
	Msg   string
	Relay int
	State bool
}

func (r *Relays) click(p *merle.Packet) {
	var msg MsgClick
	p.Unmarshal(&msg)

	r.Lock()
	r.States[msg.Relay] = msg.State
	r.Unlock()

	if p.IsThing() {
		if msg.State {
			r.drivers[msg.Relay].On()
		} else {
			r.drivers[msg.Relay].Off()
		}
	}

	p.Broadcast()
}

func (r *Relays) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdRun:     r.run,
		merle.GetState:   r.getState,
		merle.ReplyState: r.saveState,
		"Click":          r.click,
	}
}

const html = `
<!DOCTYPE html>
<html lang="en">
	<head>
		<meta name="viewport" content="width=device-width, initial-scale=1">
	</head>
	<body style="background-color:orange">
		<div id="buttons" style="display: none">
			<input type="checkbox" id="relay0" disabled=true onclick='sendClick(this, 0)'>
			<label for="relay0"> Relay 0 </label>
			<input type="checkbox" id="relay1" disabled=true onclick='sendClick(this, 1)'>
			<label for="relay1"> Relay 1 </label>
			<input type="checkbox" id="relay2" disabled=true onclick='sendClick(this, 2)'>
			<label for="relay2"> Relay 2 </label>
			<input type="checkbox" id="relay3" disabled=true onclick='sendClick(this, 3)'>
			<label for="relay3"> Relay 3 </label>
		</div>

		<script>
			var conn
			var online = false

			relays = []
			for (var i = 0; i < 4; i++) {
				relays[i] = document.getElementById("relay" + i)
			}
			buttons = document.getElementById("buttons")

			function getState() {
				conn.send(JSON.stringify({Msg: "_GetState"}))
			}

			function getIdentity() {
				conn.send(JSON.stringify({Msg: "_GetIdentity"}))
			}

			function saveState(msg) {
				for (var i = 0; i < relays.length; i++) {
					relays[i].checked = msg.States[i]
				}
			}

			function showAll() {
				for (var i = 0; i < relays.length; i++) {
					relays[i].disabled = !online
				}
				buttons.style.display = "block"
			}

			function sendClick(relay, num) {
				conn.send(JSON.stringify({Msg: "Click", Relay: num,
					State: relay.checked}))
			}

			function connect() {
				conn = new WebSocket("{{.WebSocket}}")

				conn.onopen = function(evt) {
					getIdentity()
				}

				conn.onclose = function(evt) {
					online = false
					showAll()
					setTimeout(connect, 1000)
				}

				conn.onerror = function(err) {
					conn.close()
				}

				conn.onmessage = function(evt) {
					msg = JSON.parse(evt.data)
					console.log('relays', msg)

					switch(msg.Msg) {
					case "_ReplyIdentity":
					case "_EventStatus":
						online = msg.Online
						getState()
						break
					case "_ReplyState":
						saveState(msg)
						showAll()
						break
					case "Click":
						relays[msg.Relay].checked = msg.State
						break
					}
				}
			}

			connect()
		</script>
	</body>
</html>`

func (r *Relays) Assets() *merle.ThingAssets {
	return &merle.ThingAssets{
		HtmlTemplateText: html,
	}
}
