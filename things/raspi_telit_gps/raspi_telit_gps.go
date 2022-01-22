package raspi_telit_gps

import (
	"github.com/scottfeldman/merle"
	"github.com/scottfeldman/merle/things/telit"
	"log"
	"time"
)

type thing struct {
	log          *log.Logger
	demo         bool
	gps          *telit.Gps
	ticker       *time.Ticker
	location     string
}

func NewModel(log *log.Logger, demo bool) merle.Thinger {
	return &thing{
		log: log,
		demo: demo,
		gps: &telit.Gps{},
		ticker: time.NewTicker(time.Minute),
	}
}

type msgLocation struct {
	Msg      string
	Location string
}

func (t *thing) everyMinute(p *merle.Packet) {
	var loc string

	if t.demo {
		loc = "34.134306N,118.321556W"
	} else {
		loc = t.gps.Location()
	}
	if loc == t.location {
		return
	}
	t.location = loc

	msg := msgLocation{
		Msg: "Location",
		Location: t.location,
	}

	p.Marshal(&msg).Broadcast()
}

func (t *thing) run(p *merle.Packet) {
	var err error

	if t.demo {
		goto demo
	}

	err = t.gps.Init()
	if err != nil {
		t.log.Println("GPS init failed:", err)
		return
	}

demo:
	t.log.Printf("GPS initialized")

	t.everyMinute(p)

	for {
		select{
		case <-t.ticker.C:
			t.everyMinute(p)
		}
	}
}

func (t *thing) getLocation(p *merle.Packet) {
	msg := msgLocation{
		Msg: "Location",
		Location: t.location,
	}

	p.Marshal(&msg).Reply()
}

func (t *thing) saveLocation(p *merle.Packet) {
	var msg msgLocation
	p.Unmarshal(&msg)
	t.location = msg.Location
	p.Broadcast()
}

func (t *thing) start(p *merle.Packet) {
	msg := struct{ Msg string }{Msg: "GetLocation"}
	p.Marshal(&msg).Reply()
}

func (t *thing) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		"CmdRun": t.run,
		"CmdStart": t.start,
		"GetLocation": t.getLocation,
		"Location": t.saveLocation,
	}
}

func (t *thing) Config(config merle.Configurator) error {
	return nil
}

func (t *thing) Template() string {
	return "web/templates/raspi_telit_gps.html"
}
