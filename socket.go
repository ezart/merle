package merle

import (
	"github.com/gorilla/websocket"
	"time"
)

type socketer interface {
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
