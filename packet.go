// Copyright 2021 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package merle

import (
	"encoding/json"
)

// A Packet contains a JSON message and a source connection.
type Packet struct {
	src IConn
	msg []byte
}

func NewPacket(msg interface{}) *Packet {
	var p Packet
	p.msg, _ = json.Marshal(msg)
	return &p
}

func (p *Packet) Marshal(msg interface{}) *Packet {
	p.msg, _ = json.Marshal(msg)
	return p
}

func (p *Packet) Unmarshal(msg interface{}) {
	json.Unmarshal(p.msg, msg)
}

func (p *Packet) String() string {
	return string(p.msg)
}

func (p *Packet) send(dst IConn) error {
	return dst.Send(p)
}
