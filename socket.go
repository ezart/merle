package merle

import (
	"github.com/gorilla/websocket"
	"time"
)

type ISocket interface {
	Send(*Packet) error
	Close()
	Name() string
}

type webSocket struct {
	conn *websocket.Conn
	name string
}

func newWebSocket(name string, conn *websocket.Conn) *webSocket {
	return &webSocket{name: name, conn: conn}
}

func (ws *webSocket) Send(p *Packet) error {
	return ws.conn.WriteMessage(websocket.TextMessage, p.msg)
}

func (ws *webSocket) Close() {
	ws.conn.WriteControl(websocket.CloseMessage, nil, time.Now())
}

func (ws *webSocket) Name() string {
	return ws.name
}

type chanSocket struct {
	conn chan *Packet
	name string
}

func newChanSocket(name string, conn chan *Packet) *chanSocket {
	return &chanSocket{name: name, conn: conn}
}

func (c *chanSocket) Send(p *Packet) error {
	c.conn <- p
	return nil
}

func (c *chanSocket) Close() {
	close(c.conn)
}

func (c *chanSocket) Name() string {
	return c.name
}
