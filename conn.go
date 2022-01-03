package merle

import (
	"github.com/gorilla/websocket"
	"time"
)

type IConn interface {
	Send([]byte) error
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

func (w *WsConn) Send(msg []byte) error {
	return w.conn.WriteMessage(websocket.TextMessage, msg)
}

func (w *WsConn) Close() {
	w.conn.WriteControl(websocket.CloseMessage, nil, time.Now())
}

func (w *WsConn) Name() string {
	return w.name
}
