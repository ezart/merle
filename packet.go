// Copyright 2021 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package merle

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
)

// A Packet contains a message and a (hidden) source.
type Packet struct {
	tap bool
	conn *websocket.Conn
	msg  []byte
}

func NewPacket(msg interface{}) *Packet {
	var p Packet
	p.msg, _ = json.Marshal(msg)
	return &p
}

func (p *Packet) SetTap() {
	p.tap = true
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

func (p *Packet) write() error {
	err := p.conn.WriteMessage(websocket.TextMessage, p.msg)
	if err != nil {
		log.Println("Packet writeMessage error:", err)
	}
	return err
}
