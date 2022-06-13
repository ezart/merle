// file: examples/xmas/xmas.go

package main

import (
	"flag"
	"github.com/merliot/merle"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
	"log"
	"sync"
)

type relay struct {
	driver *gpio.RelayDriver
	state bool
}

type xmas struct {
	sync.RWMutex
	relays [4]relay
}

type msg struct {
	Msg   string
	State [4]bool
}

func (x *xmas) run(p *merle.Packet) {
	adaptor := raspi.NewAdaptor()
	adaptor.Connect()

	x.relays = [4]relay{
		{driver: gpio.NewRelayDriver(adaptor, "31")}, // GPIO 6
		{driver: gpio.NewRelayDriver(adaptor, "33")}, // GPIO 13
		{driver: gpio.NewRelayDriver(adaptor, "35")}, // GPIO 19
		{driver: gpio.NewRelayDriver(adaptor, "37")}, // GPIO 26
	}

	for _, relay := range x.relays {
		relay.driver.Start()
		relay.driver.Off()
		relay.state = false
	}

	select{}
}

func (x *xmas) getState(p *merle.Packet) {
	x.RLock()
	defer x.RUnlock()

	msg := &msg{Msg: merle.ReplyState}
	for i, relay := range x.relays {
		msg.State[i] = relay.state
	}

	p.Marshal(&msg).Reply()
}

func (x *xmas) saveState(p *merle.Packet) {
	x.Lock()
	defer x.Unlock()

	var msg msg
	p.Unmarshal(&msg)

	for i, relay := range x.relays {
		relay.state = msg.State[i]
	}
}

type clickMsg struct {
	Msg   string
	Relay int
	State bool
}

func (x *xmas) click(p *merle.Packet) {
	x.Lock()
	defer x.Unlock()

	var msg clickMsg
	p.Unmarshal(&msg)

	x.relays[msg.Relay].state = msg.State

	if p.IsThing() {
		if msg.State {
			x.relays[msg.Relay].driver.On()
		} else {
			x.relays[msg.Relay].driver.Off()
		}
	}

	p.Broadcast()
}

func (x *xmas) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdRun:     x.run,
		merle.GetState:   x.getState,
		merle.ReplyState: x.saveState,
		"Click":          x.click,
	}
}

const html = `<html lang="en">
	<body>
		<div>
			<input type="checkbox" id="relay0" onclick='relayClick(this, 0)'>
			<label for="relay0"> Relay 0 </label>
			<input type="checkbox" id="relay1" onclick='relayClick(this, 1)'>
			<label for="relay1"> Relay 1 </label>
			<input type="checkbox" id="relay2" onclick='relayClick(this, 2)'>
			<label for="relay2"> Relay 2 </label>
			<input type="checkbox" id="relay3" onclick='relayClick(this, 3)'>
			<label for="relay3"> Relay 3 </label>
		</div>

		<script>
			relay = [4]
			relay[0] = document.getElementById("relay0")
			relay[1] = document.getElementById("relay1")
			relay[2] = document.getElementById("relay2")
			relay[3] = document.getElementById("relay3")

			conn = new WebSocket("{{.WebSocket}}")

			conn.onopen = function(evt) {
				conn.send(JSON.stringify({Msg: "_GetState"}))
			}

			conn.onmessage = function(evt) {
				msg = JSON.parse(evt.data)
				console.log('msg', msg)

				switch(msg.Msg) {
				case "_ReplyState":
					relay[0].checked = msg.State[0]
					relay[1].checked = msg.State[1]
					relay[2].checked = msg.State[2]
					relay[3].checked = msg.State[3]
					break
				case "Click":
					relay[msg.Relay].checked = msg.State
					break
				}
			}

			function relayClick(relay, num) {
				conn.send(JSON.stringify({Msg: "Click", Relay: num,
					State: relay.checked}))
			}
		</script>
	</body>
</html>`

func (x *xmas) Assets() *merle.ThingAssets {
	return &merle.ThingAssets{
		TemplateText: html,
	}
}

func main() {
	thing := merle.NewThing(&xmas{})

	thing.Cfg.Model = "xmas"
	thing.Cfg.Name = "xmas0"
	thing.Cfg.User = "merle"

	flag.StringVar(&thing.Cfg.MotherHost, "rhost", "", "Remote host")
	flag.StringVar(&thing.Cfg.MotherUser, "ruser", "merle", "Remote user")
	flag.BoolVar(&thing.Cfg.IsPrime, "prime", false, "Run as Thing Prime")
	flag.UintVar(&thing.Cfg.PortPublicTLS, "TLS", 0, "TLS port")

	flag.Parse()

	log.Fatalln(thing.Run())
}
