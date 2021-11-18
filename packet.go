package merle

import (
	"github.com/gorilla/websocket"
	"log"
	"sync"
)

type Packet struct {
	sync.Mutex
	conn *websocket.Conn
	Msg []byte
}

func (p *Packet) writeMessage() error {
	// TODO what is this lock protecting?
	p.Lock()
	defer p.Unlock()
	err := p.conn.WriteMessage(websocket.TextMessage, p.Msg)
	if err != nil {
		log.Println("Packet writeMessage error:", err)
	}
	return err
}
