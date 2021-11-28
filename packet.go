package merle

import (
	"github.com/gorilla/websocket"
	"log"
)

// A Packet contains a message and a (hidden) source.
type Packet struct {
	conn *websocket.Conn
	Msg  []byte
}

func (p *Packet) writeMessage() error {
	err := p.conn.WriteMessage(websocket.TextMessage, p.Msg)
	if err != nil {
		log.Println("Packet writeMessage error:", err)
	}
	return err
}
