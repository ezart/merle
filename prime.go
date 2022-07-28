// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

//go:build !tinygo
// +build !tinygo

package merle

import "fmt"

func (t *Thing) getPrimePort(id string) string {
	t.primePort.Lock()
	defer t.primePort.Unlock()

	if t.primePort.tunnelConnected {
		return "port busy"
	}

	if t.primeId != "" && t.primeId != id {
		return "no ports available"
	}

	return fmt.Sprintf("%d", t.primePort.port)
}

func (t *Thing) runOnPort(p *port, ready func(*Thing), cleanup func(*Thing)) error {
	var name = fmt.Sprintf("port:%d", p.port)
	var sock = newWebSocket(t, name, p.ws)
	var pkt = newPacket(t.bus, sock, nil)
	var msg = Msg{Msg: GetState}
	var err error

	t.log.printf("Websocket opened [%s]", name)

	t.primeSock = sock
	t.bus.plugin(sock)

	// Send GetState msg to Thing
	t.log.println("PRIME SENDING:", msg)
	sock.Send(pkt.Marshal(&msg))

	for {
		// new pkt for each rcv
		var pkt = newPacket(t.bus, sock, nil)

		pkt.msg, err = p.readMessage()
		if err != nil {
			t.log.printf("Websocket closed [%s]", name)
			break
		}

		pkt.Unmarshal(&msg)

		t.bus.receive(pkt)

		if msg.Msg == ReplyState {
			t.log.println("PRIME READY")
			ready(t)
		}
	}

	t.bus.unplug(sock)

	t.log.println("PRIME CLEANUP")
	cleanup(t)

	return nil
}

func (t *Thing) sendStatus() {
	msg := MsgEventStatus{Msg: EventStatus, Id: t.id, Online: t.online}
	newPacket(t.bus, t.primeSock, &msg).Broadcast()
}

func (t *Thing) primeReady(self *Thing) {
	t.online = true
	t.web.public.start()
	t.sendStatus()
}

func (t *Thing) primeCleanup(self *Thing) {
	t.online = false
	t.sendStatus()
}

func (t *Thing) primeAttach(p *port, msg *MsgIdentity) error {
	if msg.Model != t.Cfg.Model {
		return fmt.Errorf("Model mis-match: want %s, got %s",
			t.Cfg.Model, msg.Model)
	}

	t.id = msg.Id
	t.model = msg.Model
	t.name = msg.Name
	t.online = msg.Online
	t.startupTime = msg.StartupTime
	t.primeId = t.id

	prefix := "[" + t.id + "] "
	t.log = NewLogger(prefix, t.Cfg.LoggingEnabled)

	t.setAssetsDir(t)

	return t.runOnPort(p, t.primeReady, t.primeCleanup)
}

func (t *Thing) primeRun() error {
	t.web.private.start()
	return t.primePort.run()
}
