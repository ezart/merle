// file: examples/gps/gps.go

package main

import (
	"github.com/scottfeldman/merle"
	"github.com/scottfeldman/merle/examples/telit"
	"time"
	"log"
)

type gps struct {
	telit telit.Telit
	last  string
}

func (g *gps) run(p *merle.Packet) {
	update := struct {
		Msg      string
		Position string
	}{Msg: "update"}

	err := g.telit.Init()
	if err != nil {
		log.Fatalln("Telit init failed:", err)
	}

	for {
		update.Position = g.telit.Location()
		if update.Position != g.last {
			p.Marshal(&update).Broadcast()
			g.last = update.Position
		}
		time.Sleep(time.Minute)
	}
}

func (g *gps) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdRun: g.run,
	}
}

const html = `<html lang="en">
	<body>
		<pre id="position">Position:</pre>

		<script>
			position = document.getElementById("position")

			conn = new WebSocket("ws://localhost:8080/ws/{{.Id}}")

			conn.onmessage = function(evt) {
				msg = JSON.parse(evt.data)
				console.log('msg', msg)

				switch(msg.Msg) {
				case "update":
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
	var cfg merle.ThingConfig

	cfg.Thing.PortPublic = 8080

	merle.NewThing(&gps{}, &cfg).Run()
}
