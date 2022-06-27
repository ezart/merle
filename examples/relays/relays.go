// file: examples/relays/relays.go

package relays

import (
	"github.com/merliot/merle"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
	"sync"
)

type relay struct {
	driver *gpio.RelayDriver
	state bool
}

type thing struct {
	sync.RWMutex
	relays [4]relay
}

func NewThing() merle.Thinger {
	return &thing{}
}

type msg struct {
	Msg   string
	Connected bool
	State [4]bool
}

func (t *thing) run(p *merle.Packet) {
	adaptor := raspi.NewAdaptor()
	adaptor.Connect()

	t.relays = [4]relay{
		{driver: gpio.NewRelayDriver(adaptor, "31")}, // GPIO 6
		{driver: gpio.NewRelayDriver(adaptor, "33")}, // GPIO 13
		{driver: gpio.NewRelayDriver(adaptor, "35")}, // GPIO 19
		{driver: gpio.NewRelayDriver(adaptor, "37")}, // GPIO 26
	}

	for i, _ := range t.relays {
		t.relays[i].driver.Start()
		t.relays[i].driver.Off()
		t.relays[i].state = false
	}

	select{}
}

func (t *thing) getState(p *merle.Packet) {
	t.RLock()
	defer t.RUnlock()

	msg := &msg{Msg: merle.ReplyState}
	msg.Connected = p.IsConnected()
	for i, _ := range t.relays {
		msg.State[i] = t.relays[i].state
	}

	p.Marshal(&msg).Reply()
}

func (t *thing) saveState(p *merle.Packet) {
	t.Lock()
	defer t.Unlock()

	var msg msg
	p.Unmarshal(&msg)

	for i, _ := range t.relays {
		t.relays[i].state = msg.State[i]
	}
}

type clickMsg struct {
	Msg   string
	Relay int
	State bool
}

func (t *thing) click(p *merle.Packet) {
	t.Lock()
	defer t.Unlock()

	var msg clickMsg
	p.Unmarshal(&msg)

	t.relays[msg.Relay].state = msg.State

	if p.IsThing() {
		if msg.State {
			t.relays[msg.Relay].driver.On()
		} else {
			t.relays[msg.Relay].driver.Off()
		}
	}

	p.Broadcast()
}

func (t *thing) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdRun:     t.run,
		merle.GetState:   t.getState,
		merle.ReplyState: t.saveState,
		"Click":          t.click,
	}
}

const html = `
<!DOCTYPE html>
<html lang="en">
	<head>
		<meta name="viewport" content="width=device-width, initial-scale=1">
	</head>
	<body>
		<div id="buttons" style="display: none;">
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

			relays = []
			for (var i = 0; i < 4; i++) {
				relays[i] = document.getElementById("relay" + i)
			}
			buttons = document.getElementById("buttons")

			function getState() {
				conn.send(JSON.stringify({Msg: "_GetState"}))
			}

			function saveState(states) {
				for (var i = 0; i < relays.length; i++) {
					relays[i].checked = states[i]
				}
			}

			function enable(connected) {
				for (var i = 0; i < relays.length; i++) {
					relays[i].disabled = !connected
				}
			}

			function showAll() {
				buttons.style.display = "block"
			}

			function clearAll() {
				buttons.style.display = "none"
			}

			function sendClick(relay, num) {
				conn.send(JSON.stringify({Msg: "Click", Relay: num,
					State: relay.checked}))
			}

			function connect() {
				conn = new WebSocket("{{.WebSocket}}")

				conn.onopen = function(evt) {
					clearAll()
					getState()
				}

				conn.onclose = function(evt) {
					enable(false)
					setTimeout(connect, 1000)
				}

				conn.onerror = function(err) {
					conn.close()
				}

				conn.onmessage = function(evt) {
					msg = JSON.parse(evt.data)
					console.log('msg', msg)

					switch(msg.Msg) {
					case "_ReplyState":
						saveState(msg.State)
						enable(msg.Connected)
						showAll()
						break
					case "_EventConnect":
						getState()
						break
					case "_EventDisconnect":
						enable(false)
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

func (t *thing) Assets() *merle.ThingAssets {
	return &merle.ThingAssets{
		HtmlTemplateText: html,
	}
}
