package merle

import (
	"fmt"
)

type Bridge struct {
	Thing
	// TODO need R/W lock for t.children[] map
	children    map[string]*Thing
	//childStatus func(*Thing)
}

type msgChild struct {
	Id     string
	Model  string
	Name   string
	Status string
}

type msgChildren struct {
	Msg      string
	Children []msgChild
}

func (b *Bridge) getChildren(p *Packet) {
	resp := msgChildren{ Msg: "ReplyChildren" }
	for _, child := range b.children {
		resp.Children = append(resp.Children, msgChild{child.id,
			child.model, child.name, child.status})
	}
	b.Reply(p.Marshal(&resp))
}

func (b *Bridge) getChild(id string) *Thing {
	if child, ok := b.children[id]; ok {
		return child
	}
	return nil
}

type SpamStatus struct {
	Msg    string
	Id     string
	Model  string
	Name   string
	Status string
}

func (b *Bridge) changeStatus(child *Thing, status string) {
	child.status = status

	spam := SpamStatus{
		Msg:    "SpamStatus",
		Id:     child.id,
		Model:  child.model,
		Name:   child.name,
		Status: child.status,
	}
	b.Broadcast(b.NewPacket(&spam))

	//if t.childStatus != nil {
	//	t.childStatus(child)
	//}
}

func (b *Bridge) attach(p *port, spec *msgIdentity) error {
	var err error

	if b.id == spec.Id {
		return fmt.Errorf("Sorry, you can't be your own Mother")
	}

	child := b.getChild(spec.Id)

	if child == nil {
		child, err = b.stork(spec.Id, spec.Model, spec.Name)
		if err != nil {
			return fmt.Errorf("Creating new Thing failed: %s", err)
		}
		b.children[spec.Id] = child
	} else {
		if child.model != spec.Model {
			return fmt.Errorf("Model mismatch")
		}
		if child.name != spec.Name {
			return fmt.Errorf("Name mismatch")
		}
	}

	child.startupTime = spec.StartupTime

	b.changeStatus(child, "online")
	child.run(p)
	b.changeStatus(child, "offline")

	return nil
}

func (b *Bridge) InitBridge(id, model, name string, max uint, match string) (*Thing, error) {

	t, err := b.InitThing(id, model, name)
	if err != nil {
		return nil, err
	}

	b.children = make(map[string]*Thing)
	b.childFromId = b.getChild

	//b.childStatus = status
	b.Subscribe("GetChildren", b.getChildren)
	b.muxPrivate.HandleFunc("/port/{id}", getPort)

	b.log.Println("Listening for Children...")

	if err := portScan(max, match, b.attach); err != nil {
		return nil, err
	}

	return t, nil
}
