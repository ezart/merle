// file: examples/gps/gps.go

package main

import (
	"github.com/scottfeldman/merle"
	"github.com/scottfeldman/merle/examples/telit"
	"time"
)

type gps struct {
	telit Telit
	last  string
}

func (g *gps) run(p *merle.Packet) {
	update := struct {
		Msg      string
		Position string
	}{Msg: "update"}

	err := telit.Init()
	if err != nil {
		log.Println("Telit init failed:", err)
		return
	}

	for {
		update.Position = telit.Location()
		if update.Position != last {
			p.Marshal(&update).Broadcast()
			last = update.Position
		}
		time.Sleep(time.Second)
	}
}

func (g *gps) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdRun: b.run,
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
