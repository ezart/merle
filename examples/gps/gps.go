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
	lastLat  float64
	lastLong float64
}

type msg struct {
	Msg  string
	Lat  float64
	Long float64
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
		msg.Lat, msg.Long = telit.Location()

		g.Lock()
		if msg.Lat != g.lastLat || msg.Long != g.lastLong {
			g.lastLat = msg.Lat
			g.lastLong = msg.Long
			p.Marshal(&msg).Broadcast()
		}
		g.Unlock()

		time.Sleep(time.Minute)
	}
}

func (g *gps) getState(p *merle.Packet) {
	g.RLock()
	defer g.RUnlock()
	msg := &msg{Msg: merle.ReplyState, Lat: g.lastLat, Long: g.lastLong}
	p.Marshal(&msg).Reply()
}

func (g *gps) saveState(p *merle.Packet) {
	g.Lock()
	defer g.Unlock()
	var msg msg
	p.Unmarshal(&msg)
	g.lastLat = msg.Lat
	g.lastLong = msg.Long
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

const html = `
<!DOCTYPE html>
<html lang="en">
	<head>
		<!-- Leaflet's CSS -->
		<link rel="stylesheet" href="https://unpkg.com/leaflet@1.8.0/dist/leaflet.css"
		integrity="sha512-hoalWLoI8r4UszCkZ5kL8vayOGVae1oxXe/2A4AO6J9+580uKHDO3JdHb7NzwwzK5xr/Fs0W40kiNHxM9vyTtQ=="
		crossorigin=""/>

		<!-- Leaflet's JavaScript -->
		<script src="https://unpkg.com/leaflet@1.8.0/dist/leaflet.js"
		integrity="sha512-BB3hKbKWOc9Ez/TAwyWxNXeoV9c1v6FIeYiBieIWkpLjauysF18NzgR1MBNBXf8/KABdlkX68nAhlwcDFLGPCQ=="
		crossorigin=""></script>
	</head>
	<body>
		<div id="map" style="height:100%"></div>

		<script>
			<!-- Create a Leaflet map using OpenStreetMap -->
			map = L.map('map').setZoom(13)
			L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
			    maxZoom: 19,
			    attribution: '© OpenStreetMap'
			}).addTo(map)

			<!-- Create a map marker with popup that has [Id, Model, Name] -- !>
			popup = "ID: {{.Id}}<br>Model: {{.Model}}<br>Name: {{.Name}}"
			marker = L.marker([0, 0]).addTo(map).bindPopup(popup);

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
					marker.setLatLng([msg.Lat, msg.Long])
					map.panTo([msg.Lat, msg.Long])
					break
				}
			}
		</script>
	</body>
</html>`

func (g *gps) Assets() *merle.ThingAssets {
	return &merle.ThingAssets{
		HtmlTemplateText: html,
	}
}

func main() {
	thing := merle.NewThing(&gps{})

	thing.Cfg.Model = "gps"
	thing.Cfg.Name = "gypsy"
	thing.Cfg.User = "merle"

	thing.Cfg.PortPublic = 80
	thing.Cfg.PortPrivate = 8080

	flag.StringVar(&thing.Cfg.MotherHost, "rhost", "", "Remote host")
	flag.StringVar(&thing.Cfg.MotherUser, "ruser", "merle", "Remote user")
	flag.BoolVar(&thing.Cfg.IsPrime, "prime", false, "Run as Thing Prime")
	flag.UintVar(&thing.Cfg.PortPublicTLS, "TLS", 0, "TLS port")

	flag.Parse()

	log.Fatalln(thing.Run())
}
