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
		if msg.Lat != g.lastLat || msg.Long != g.lastLong {
			g.lastLat = msg.Lat
			g.lastLong = msg.Long
			p.Marshal(&msg).Broadcast()
		}
		g.Unlock()

		time.Sleep(time.Minute)
	}
}

type place struct {
	lat  float64
	long float64
}

var places = [...]place{
	{38.890591, -77.004745},
	{38.364895, -98.774300},
	{37.838108, -94.710022},
	{51.522121, -0.129420},
	{51.503052, -0.118391},
	{51.505028, -0.115013},
	{51.505444, -0.091249},
	{51.512409, -0.125146},
	{51.511227, -0.119470},
	{29.199438, -99.788361},
	{42.314812, -71.147903},
	{51.037434, -114.052261},
	{26.158476, -80.325958},
	{38.405903, -96.188339},
	{37.753098, -100.024872},
	{38.661697, -96.492599},
	{42.812351, -73.941849},
	{37.679878, -95.459778},
	{39.563606, -95.125549},
	{41.579830, -93.791328},
	{42.464397, -93.829056},
	{42.499504, -92.358665},
	{42.495132, -96.400070},
	{34.014465, -118.449669},
	{53.342247, -6.258232},
	{41.016621, -92.430550},
	{41.299023, -92.653198},
	{41.703957, -93.054817},
	{40.969242, -91.555420},
	{40.407204, -91.410805},
	{41.746952, -92.729362},
	{42.512745, -94.188148},
	{38.844112, -77.309181},
	{43.403500, -94.843323},
	{42.516033, -90.718506},
	{41.250854, -95.882042},
	{41.827965, -90.249619},
	{42.754681, -95.557831},
	{43.066807, -92.683464},
	{40.810947, -91.131844},
	{42.060650, -93.885490},
	{41.800617, -91.869537},
	{40.800262, -85.830887},
	{38.678299, -87.522491},
	{51.503162, -0.086852},
	{43.596413, -79.637047},
	{41.483845, -87.063965},
	{39.472298, -87.401917},
	{39.524483, -85.786476},
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
		g.Unlock()

		p.Marshal(&msg).Broadcast()
		time.Sleep(time.Minute)
		i = (i + 1) % len(places)
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
