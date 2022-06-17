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

type readyCb func(*Thing)
type cleanupCb func(*Thing)

func (t *Thing) runOnPort(p *port, ready readyCb, cleanup cleanupCb) error {
	var name = fmt.Sprintf("port:%d", p.port)
	var sock = newWebSocket(name, p.ws)
	var pkt = newPacket(t.bus, sock, nil)
	var msg = Msg{Msg: GetState}
	var err error

	t.log.Printf("Websocket opened [%s]", name)

	t.bus.plugin(sock)

	// Send GetState msg to Thing
	t.log.Println("Sending:", msg)
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

		// Receiving ReplyState is a special case.  The socket is
		// disabled for broadcasts until ReplyState is received.  This
		// ensures the other end doesn't receive unsolicited broadcast
		// messages before ReplyState. Also, It's safe now to handle
		// html and ws requests on Thing Prime.

		if msg.Msg == ReplyState {
			sock.SetFlags(sock.Flags() | bcast)
			ready(t)
		}
	}

	cleanup(t)

	t.bus.unplug(sock)

	return err
}

func (t *Thing) primeReady(self *Thing) {
	t.web.public.activate()
}

func (t *Thing) primeCleanup(self *Thing) {
	t.web.public.deactivate()
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

func (t *Thing) runPrime() error {
	t.web.start()
	t.tunnel.start()
	return t.primePort.run()
}
