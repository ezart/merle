// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

// Hub is a Thing that connects other Things, including other Hubs.  Hub keeps
// track of Thing connection status and displays each Thing.
package chat

import (
	"github.com/scottfeldman/merle"
	"html/template"
	"net/http"
)

var templ *template.Template

func init() {
	templ = template.Must(template.ParseFiles("web/templates/chat.html"))
}

type chat struct {
	merle.Thing
}

func (c *chat) init() error {
	return nil
}

func (c *chat) run() {
	for {
	}
}

func (c *chat) home(w http.ResponseWriter, r *http.Request) {
	templ.Execute(w, c.HomeParams(r, nil))
}

func NewChat(id, model, name string) *merle.Thing {
	c := &chat{}

	t := c.InitThing(id, model, name)
	if t == nil {
		return nil
	}

	t.Init = c.init
	t.Run = c.run
	t.Home = c.home

	t.Subscribe("CmdNewUser", c.Broadcast)
	t.Subscribe("CmdText", c.Broadcast)

	return t
}
