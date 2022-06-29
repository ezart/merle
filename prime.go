// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package merle

import (
	"fmt"
	"log"
	"os"
)

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

	t.log.Printf("Websocket opened [%s]", name)

	t.bus.plugin(sock)

	// Send GetState msg to Thing
	t.log.Println("PRIME SENDING:", msg)
	sock.Send(pkt.Marshal(&msg))

	for {
		// new pkt for each rcv
		var pkt = newPacket(t.bus, sock, nil)

		pkt.msg, err = p.readMessage()
		if err != nil {
			t.log.Printf("Websocket closed [%s]", name)
			break
		}

		pkt.Unmarshal(&msg)

		t.bus.receive(pkt)

		if msg.Msg == ReplyState {
			t.log.Println("PRIME READY")
			ready(t)
		}
	}

	t.bus.unplug(sock)

	t.log.Println("PRIME CLEANUP")
	cleanup(t)

	return nil
}

func (t *Thing) sendConnect() {
	msg := MsgId{Msg: EventConnect, Id: t.id}
	newPacket(t.bus, nil, &msg).Broadcast()
}

func (t *Thing) sendDisconnect() {
	msg := MsgId{Msg: EventDisconnect, Id: t.id}
	newPacket(t.bus, nil, &msg).Broadcast()
}

func (t *Thing) primeReady(self *Thing) {
	t.connected = true
	t.web.public.start()
	t.sendConnect()
}

func (t *Thing) primeCleanup(self *Thing) {
	t.connected = false
	t.sendDisconnect()
}

func (t *Thing) primeAttach(p *port, msg *MsgIdentity) error {
	if msg.Model != t.Cfg.Model {
		return fmt.Errorf("Model mis-match: want %s, got %s",
			t.Cfg.Model, msg.Model)
	}

	t.id = msg.Id
	t.model = msg.Model
	t.name = msg.Name
	t.startupTime = msg.StartupTime
	t.primeId = t.id

	prefix := "[" + t.id + "] "
	t.log = log.New(os.Stderr, prefix, 0)

	t.setAssetsDir(t)

	return t.runOnPort(p, t.primeReady, t.primeCleanup)
}

func (t *Thing) primeRun() error {
	t.web.private.start()
	return t.primePort.run()
}
