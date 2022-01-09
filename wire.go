package merle

type wireSocket struct {
	name string
	bus *bus
}

func newWireSocket(name string, bus *bus) *wireSocket {
	return &wireSocket{name: name, bus: bus}
}

func (s *wireSocket) Send(p *Packet) error {
	s.bus.receive(p.clone(s.bus, s))
	return nil
}

func (s *wireSocket) Close() {
}

func (s *wireSocket) Name() string {
	return s.name
}

type wire struct {
	name string
	aSock *wireSocket
	bSock *wireSocket
}

func newWire(name string, a *bus, b *bus) *wire {
	return &wire{
		name: name,
		aSock: newWireSocket(name + ":a", a),
		bSock: newWireSocket(name + ":b", b),
	}
}
