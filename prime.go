// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package merle

import (
	"fmt"
	"log"
	"os"
)

func (t *Thing) changeStatus(status string) {
	t.status = status

	spam := MsgSpamStatus{
		Msg:    SpamStatus,
		Id:     t.id,
		Model:  t.model,
		Name:   t.name,
		Status: t.status,
	}
	newPacket(t.bus, nil, &spam).Broadcast()
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

	t.changeStatus("online")
	err := t.runOnPort(p)
	t.changeStatus("offline")

	return err
}

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

func (t *Thing) runPrime() error {
	t.web.start()
	t.tunnel.start()
	return t.primePort.run()
}
