package merle

import (
	"github.com/gorilla/websocket"
	"log"
)

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
