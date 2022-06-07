// file: examples/gps/gps.go

package main

import (
	"flag"
	"github.com/merliot/merle"
	"github.com/merliot/merle/examples/telit"
	"log"
	"sync"
	"time"
)

type gps struct {
	sync.RWMutex
	lastPosition string
}

type msg struct {
	Msg      string
	Position string
}

func (g *gps) run(p *merle.Packet) {
	var telit telit.Telit
	msg := &msg{Msg: "Update"}

	err := telit.Init()
	if err != nil {
		log.Fatalln("Telit init failed:", err)
		return
	}

	for {
		msg.Position = telit.Location()

		g.Lock()
		if msg.Position != g.lastPosition {
			p.Marshal(&msg).Broadcast()
			g.lastPosition = msg.Position
		}
		g.Unlock()

		time.Sleep(time.Minute)
	}
}

func (g *gps) getState(p *merle.Packet) {
	g.RLock()
	defer g.RUnlock()
	msg := &msg{Msg: merle.ReplyState, Position: g.lastPosition}
	p.Marshal(&msg).Reply()
}

func (g *gps) saveState(p *merle.Packet) {
	var msg msg
	p.Unmarshal(&msg)
	g.Lock()
	g.lastPosition = msg.Position
	g.Unlock()
}

func (g *gps) update(p *merle.Packet) {
	g.saveState(p)
	p.Broadcast()
}

func (g *gps) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdRun:     g.run,
		merle.GetState:   g.getState,
		merle.ReplyState: g.saveState,
		"Update":         g.update,
	}
}

const html = `<html lang="en">
	<body>
		<pre id="position">Position:</pre>

		<script>
			position = document.getElementById("position")

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
					position.textContent = "Position: " + msg.Position
					break
				}
			}
		</script>
	</body>
</html>`

func (g *gps) Assets() *merle.ThingAssets {
	return &merle.ThingAssets{
		TemplateText: html,
	}
}

func main() {
	cfg := merle.FlagThingConfig("", "gps", "gypsy", "merle")
	flag.Parse()

	thing := merle.NewThing(&gps{}, cfg)
	log.Fatalln(thing.Run())
}
