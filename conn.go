package merle

import (
	"github.com/gorilla/websocket"
	"time"
)

type IConn interface {
	Send(*Packet) error
	Close()
	Name() string
}

type WsConn struct {
	conn *websocket.Conn
	name string
}

func NewWsConn(name string, conn *websocket.Conn) *WsConn {
	return &WsConn{name: name, conn: conn}
}

func (c *WsConn) Send(p *Packet) error {
	return c.conn.WriteMessage(websocket.TextMessage, p.msg)
}

func (c *WsConn) Close() {
	c.conn.WriteControl(websocket.CloseMessage, nil, time.Now())
}

func (c *WsConn) Name() string {
	return c.name
}

type ChConn struct {
	conn chan *Packet
	name string
}

func NewChConn(name string, conn chan *Packet) *ChConn {
	return &ChConn{name: name, conn: conn}
}

func (c *ChConn) Send(p *Packet) error {
	c.conn <- p
	return nil
}

func (c *ChConn) Close() {
	close(c.conn)
}

func (c *ChConn) Name() string {
	return c.name
}
