// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package merle

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

const testId = "HS30_01132" // sorry if that's your car
const testModel = "240z"
const testName = "Fairlady"
const helloWorld = "Hello World!"

type sparse struct {
}

func (s *sparse) Subscribers() Subscribers {
	return Subscribers{}
}

func (s *sparse) Assets() *ThingAssets {
	return &ThingAssets{}
}

func TestBogusRun(t *testing.T) {
	var cfg ThingConfig
	var thinger sparse

	cfg.Thing.Id = testId

	thing := NewThing(&thinger, &cfg)
	if thing == nil {
		t.Errorf("Create with non-nil thinger/cfg failed")
	}

	err := thing.Run()
	if err == nil {
		t.Errorf("Run should have errored out")
	}
}

type simple struct {
	done chan bool
}

func (s *simple) run(p *Packet) {
	s.done = make(chan bool)

	for {
		select {
		case <-s.done:
			return
		}
	}
}

func (s *simple) quit(p *Packet) {
	s.done <- true
}

func (s *simple) Subscribers() Subscribers {
	return Subscribers{
		CmdRun: s.run,
		"quit": s.quit,
	}
}

func (s *simple) Assets() *ThingAssets {
	return &ThingAssets{
		TemplateText: helloWorld,
	}
}

func checkIdentityResp(r *MsgIdentity) error {
	if r.Id != testId ||
		r.Model != testModel ||
		r.Name != testName {
		return fmt.Errorf("Identify not matching")
	}
	return nil
}

func testIdentify(t *testing.T, thing *Thing, httpPort uint) {
	var p = newPort(thing, httpPort, nil)

	err := p.wsOpen()
	if err != nil {
		t.Errorf("ws open failed: %s", err)
	}

	err = p.wsIdentity()
	if err != nil {
		t.Errorf("ws identify failed: %s", err)
	}

	resp, err := p.wsReplyIdentity()
	if err != nil {
		t.Errorf("ws identify response failed: %s", err)
	}

	err = checkIdentityResp(resp)
	if err != nil {
		t.Errorf("Unexpected identify response: %s", err)
	}

	p.ws.Close()
}

func testDone(t *testing.T, thing *Thing, httpPort uint) {
	var p = newPort(thing, httpPort, nil)

	err := p.wsOpen()
	if err != nil {
		t.Errorf("ws open failed: %s", err)
	}

	// Send msg to shutdown device
	var msg = struct{ Msg string }{Msg: "quit"}

	err = p.ws.WriteJSON(msg)
	if err != nil {
		t.Errorf("ws writeJSON failed: %s", err)
	}

	p.ws.Close()
}

func testHomePage(t *testing.T, httpPort uint) {
	url := fmt.Sprintf("http://localhost:%d", httpPort)

	get, err := http.Get(url)
	if err != nil {
		t.Errorf("Get %s failed: %s", url, err)
	}

	body, err := io.ReadAll(get.Body)
	get.Body.Close()

	if err != nil {
		t.Errorf("Get %s failed: %s", url, err)
	}

	contents := strings.TrimSpace(string(body))
	if contents != helloWorld {
		t.Errorf("Get %s body failed.  Got: %s, wanted %s",
			url, contents, helloWorld)
	}
}

func testSimple(t *testing.T, thing *Thing, publicPort, privatePort uint) {
	// sleep a second for http servers to start
	time.Sleep(time.Second)
	testHomePage(t, publicPort)
	testIdentify(t, thing, privatePort)
	testDone(t, thing, privatePort)
}

func TestRun(t *testing.T) {
	var cfg ThingConfig
	var thinger simple

	cfg.Thing.Id = testId
	cfg.Thing.Model = testModel
	cfg.Thing.Name = testName

	cfg.Thing.PortPublic = 8080
	cfg.Thing.PortPrivate = 8081

	thing := NewThing(&thinger, &cfg)
	if thing == nil {
		t.Errorf("Create with non-nil thinger/cfg failed")
	}

	go testSimple(t, thing, cfg.Thing.PortPublic, cfg.Thing.PortPrivate)

	err := thing.Run()
	if err == nil {
		t.Errorf("Run should have errored out")
	}
}
