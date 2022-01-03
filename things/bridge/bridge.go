// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package bridge

import (
	"github.com/scottfeldman/merle"
	"github.com/scottfeldman/merle/config"
	"html/template"
	"net/http"
	"log"
)

var templ *template.Template

func init() {
	templ = template.Must(template.ParseFiles("web/templates/bridge.html"))
}

var cfg struct {
	Bridge struct {
		Max   uint   `yaml:"Max"`
		Match string `yaml:"Match"`
	} `yaml:"Bridge"`
}

type bridge struct {
	merle.Thing
	bus    chan *merle.Packet
	online map[*merle.Thing]merle.IConn
}

func (b *bridge) init(soft bool) error {
	err := config.ParseFile(b.ConfigFile(), &cfg)
	if err != nil {
		return err
	}

	//b.bus = make(chan *merle.Packet, cfg.Bridge.Max)
	b.bus = make(chan *merle.Packet)
	b.online = make(map[*merle.Thing]merle.IConn)

	return b.ListenForThings(cfg.Bridge.Max, cfg.Bridge.Match)
}

func (b *bridge) run() {
	for {
		select {
		case p := <-b.bus:
			log.Println("GOT PACKET", p.String())
		}
	}
}

func (b *bridge) home(w http.ResponseWriter, r *http.Request) {
	templ.Execute(w, b.HomeParams(r, nil))
}

func (b *bridge) connect(child *merle.Thing) {
	if child.Status() == "online" {
		conn := merle.NewChConn("uc:" + child.Id(), b.bus)
		child.ConnAdd(conn)
		b.online[child] = conn
	} else {
		child.ConnDel(b.online[child])
		delete(b.online, child)
	}
}

func NewThing(id, model, name string) *merle.Thing {
	b := &bridge{}

	b.Init = b.init
	b.Run = b.run
	b.Home = b.home
	b.Connect = b.connect

	return b.InitThing(id, model, name)
}
