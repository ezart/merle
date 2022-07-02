// file: examples/gps/gps.go

package gps

import (
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

func NewGps() merle.Thinger {
	return &gps{}
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
		merle.CmdRun:      g.run,
		merle.GetState:    g.getState,
		merle.ReplyState:  g.saveState,
		"Update":          g.update,
	}
}

const html = `
<html lang="en">
	<head>
		<meta name="viewport" content="width=device-width, initial-scale=1">

		<!-- Leaflet's CSS -->
		<link rel="stylesheet" href="https://unpkg.com/leaflet@1.8.0/dist/leaflet.css"
		integrity="sha512-hoalWLoI8r4UszCkZ5kL8vayOGVae1oxXe/2A4AO6J9+580uKHDO3JdHb7NzwwzK5xr/Fs0W40kiNHxM9vyTtQ=="
		crossorigin=""/>

		<!-- Leaflet's JavaScript -->
		<script src="https://unpkg.com/leaflet@1.8.0/dist/leaflet.js"
		integrity="sha512-BB3hKbKWOc9Ez/TAwyWxNXeoV9c1v6FIeYiBieIWkpLjauysF18NzgR1MBNBXf8/KABdlkX68nAhlwcDFLGPCQ=="
		crossorigin=""></script>

		<style>
		#overlay {
			position: fixed;
			display: none;
			width: 100%;
			height: 100%;
			top: 0;
			left: 0;
			right: 0;
			bottom: 0;
			background-color: rgba(0,0,0,0.5);
			z-index: 2000;
			cursor: wait;
		}
		#offline {
			position: absolute;
			top: 50%;
			left: 50%;
			font-size: 50px;
			color: white;
			transform: translate(-50%,-50%);
		}
		</style>
	</head>
	<body style="margin: 0">
		<div id="map" style="height:100%"></div>
		<div id="overlay">
			<div id="offline">Offline</div>
		</div>

		<script>
			var conn
			var online = false

			<!-- Create a Leaflet map using OpenStreetMap -->
			map = L.map('map').setZoom(13)
			L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
			    maxZoom: 19,
			    attribution: '© OpenStreetMap'
			}).addTo(map)

			<!-- Create a map marker with popup that has [Id, Model, Name] -- !>
			popup = "ID: {{.Id}}<br>Model: {{.Model}}<br>Name: {{.Name}}"
			marker = L.marker([0, 0]).addTo(map).bindPopup(popup);

			function getState() {
				conn.send(JSON.stringify({Msg: "_GetState"}))
			}

			function getIdentity() {
				conn.send(JSON.stringify({Msg: "_GetIdentity"}))
			}

			function show() {
				overlay = document.getElementById("overlay")
				if (online) {
					overlay.style.display = "none"
				} else {
					overlay.style.display = "block"
				}
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
					console.log('gps', msg)

					switch(msg.Msg) {
					case "_ReplyIdentity":
					case "_EventStatus":
						online = msg.Online
						getState()
						break
					case "_ReplyState":
					case "Update":
						marker.setLatLng([msg.Lat, msg.Long])
						map.panTo([msg.Lat, msg.Long])
						show()
						break
					}
				}
			}

			connect()
		</script>
	</body>
</html>`

func (g *gps) Assets() *merle.ThingAssets {
	return &merle.ThingAssets{
		HtmlTemplateText: html,
	}
}
