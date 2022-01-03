package merle

import (
	"github.com/gorilla/websocket"
	"time"
)

type IConn interface {
	Send([]byte) error
	Close()
}

type WsConn struct {
	conn *websocket.Conn
}

func NewWsConn(conn *websocket.Conn) *WsConn {
	return &WsConn{conn: conn}
}

func (w *WsConn) Send(msg []byte) error {
	return w.conn.WriteMessage(websocket.TextMessage, msg)
}

func (w *WsConn) Close() {
	w.conn.WriteControl(websocket.CloseMessage, nil, time.Now())
}
