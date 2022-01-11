package merle

import (
	"github.com/gorilla/websocket"
	"time"
)

// Socketer is an interface to a socket.  A socket plugs into a bus.
type socketer interface {
	// Send the packet on bus the socket is connected to
	Send(*Packet) error
	// Close the socket
	Close()
	// Name of the socket
	Name() string
}

// Websocket socket
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

// Wire socket
type wireSocket struct {
	name string
	bus  *bus
	opposite *wireSocket
}

func newWireSocket(name string, bus *bus, opposite *wireSocket) *wireSocket {
	return &wireSocket{name: name, bus: bus, opposite: opposite}
}

func (s *wireSocket) Send(p *Packet) error {
	s.bus.receive(p.clone(s.bus, s.opposite))
	return nil
}

func (s *wireSocket) Close() {
}

func (s *wireSocket) Name() string {
	return s.name
}
