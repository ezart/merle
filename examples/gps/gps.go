// file: examples/gps/gps.go

package gps

import (
	"log"
	"sync"
	"time"

	"github.com/merliot/merle"
	"github.com/merliot/merle/examples/telit"
)

type gps struct {
	sync.Mutex
	lastLat  float64
	lastLong float64
	Demo     bool
}

func NewGps() *gps {
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
		changed := false
		if msg.Lat != g.lastLat || msg.Long != g.lastLong {
			g.lastLat = msg.Lat
			g.lastLong = msg.Long
			p.Marshal(&msg)
			changed = true
		}
		g.Unlock()

		if changed {
			p.Broadcast()
		}

		time.Sleep(time.Minute)
	}
}

type place struct {
	lat  float64
	long float64
}

var places = [...]place{
	{57.75, 12},
	{35.0064, 135.8674},
	{56.495, 84.975},
	{35.6, 103.2},
	{33.8455, 132.7658},
	{49.4304, 1.08},
	{22.5804, 113.08},
	{17.0827, -96.6699},
	{-19.82, 34.87},
	{16.33, 80.45},
	{34.42, 35.87},
	{34.796, 48.515},
	{38.3204, 116.87},
	{5.98, 116.11},
	{-28.0815, 153.4482},
	{27.1304, 115},
	{-23.3, -51.18},
	{54.62, 39.72},
	{30.32, 112.23},
	{6.33, -75.57},
	{57.14, 65.53},
	{52.62, 39.64},
	{26.7204, 88.455},
	{39.795, 30.53},
	{5.55, 95.32},
	{23.1904, 75.79},
	{-24.7834, -65.4166},
	{53.18, 45},
	{36.4203, 2.83},
	{46.9677, 31.9843},
	{32.6149, 44.0245},
	{30.005, 32.5499},
	{50.3304, 18.67},
	{-0.3031, 100.3615},
	{42.9, 125.13},
	{6.12, 102.23},
	{-23.2, -46.88},
	{55.9483, -3.2191},
	{19.32, -98.23},
	{40.2458, -111.6457},
	{34.0804, 49.7},
	{14.47, 75.92},
	{-33.03, -71.54},
	{22.6817, 120.4817},
	{36.92, 7.76},
	{20.71, 77.01},
	{50.8303, -0.17},
	{46.3487, 48.055},
	{53.8, -1.75},
	{41.1142, 16.8728},
	{29.0171, -110.1333},
	{6.2104, 7.07},
	{-25.34, -57.52},
	{24.9889, 121.3111},
	{23.4755, 120.4351},
	{25.1333, 121.7333},
	{21.6, 105.83},
	{46.6704, 131.35},
	{28.09, 30.75},
	{31.0504, 30.47},
	{-7.6296, 112.9},
	{34.7171, 136.5167},
	{-8.5795, 116.135},
	{0.033, -51.05},
	{42.2705, -71.8079},
	{26.08, -98.3},
	{39.0618, 66.8315},
	{33.4017, -111.7181},
	{33.5833, 36.4},
	{42, 21.4335},
	{-2.52, 32.93},
	{37.928, 102.641},
	{27.9861, -80.6628},
	{19.6158, 37.2164},
	{-31.6239, -60.69},
	{54.2, 37.6299},
	{29.0804, 31.09},
	{42.8823, 129.5128},
	{41.6639, -83.5822},
	{44.5004, 11.34},
	{29.97, 77.55},
	{33.5719, -117.1909},
	{17.35, 76.82},
	{22.8504, 88.52},
	{37.6897, -97.3442},
	{7.4804, 4.56},
	{-12.25, -38.97},
	{3.0667, 101.55},
	{47.0962, 37.5562},
	{41.5725, -93.6104},
	{16.75, -93.15},
	{34.33, 62.17},
	{52.43, 31},
	{23.0504, 112.45},
	{-22.7499, -47.33},
	{20.9, 74.77},
	{49.8304, 18.25},
	{31.9201, 54.37},
	{32.52, 74.56},
	{55.34, 86.09},
	{8.55, 39.27},
	{40.5834, -74.1496},
	{28.6804, 121.45},
	{41.5504, 120.42},
	{-21.77, -43.375},
	{24.6, 73.73},
}

func (g *gps) runDemo(p *merle.Packet) {
	msg := &msg{Msg: "Update"}
	p.Marshal(&msg).Broadcast()

	i := 0
	for {
		msg.Lat = places[i].lat
		msg.Long = places[i].long

		g.Lock()
		g.lastLat = places[i].lat
		g.lastLong = places[i].long
		p.Marshal(&msg)
		g.Unlock()

		p.Broadcast()
		time.Sleep(time.Minute)
		i = (i + 1) % len(places)
	}
}

func (g *gps) getState(p *merle.Packet) {
	g.Lock()
	msg := &msg{Msg: merle.ReplyState, Lat: g.lastLat, Long: g.lastLong}
	p.Marshal(&msg)
	g.Unlock()
	p.Reply()
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
	subs := merle.Subscribers{
		merle.CmdRun:     g.run,
		merle.GetState:   g.getState,
		merle.ReplyState: g.saveState,
		"Update":         g.update,
	}

	if g.Demo {
		subs[merle.CmdRun] = g.runDemo
	}

	return subs
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
