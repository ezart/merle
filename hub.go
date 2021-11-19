package merle

import (
	"encoding/json"
	"sync"
)

type Hub struct {
	sync.Mutex
	supportedDevices map[string]DeviceGenerator
	devices map[string]IDevice
}

func NewHub(supportedDevices map[string]DeviceGenerator) *Hub {
	return &Hub{
		supportedDevices: supportedDevices,
		devices: make(map[string]IDevice),
	}
}

func (h *Hub) getDevice(resp *MsgIdentifyResp) IDevice {
	h.Lock()
	defer h.Unlock()

	if d, ok := h.devices[resp.Id]; ok {
		return d
	}

	if gen, ok := h.supportedDevices[resp.Model]; ok {
		d := gen(resp.Id, resp.Model, resp.Name, resp.StartupTime)
		h.devices[resp.Id] = d
		return d
	}

	return nil
}

func (p *Port) run(d IDevice) {
	var pkt = &Packet{
		conn: p.ws,
	}
	var msg = MsgCmd{
		Type: MsgTypeCmd,
		Cmd: CmdStart,
	}
	var err error

	pkt.Msg, _ = json.Marshal(&msg)
	d.ReceivePacket(pkt)

	for {
		pkt.Msg, err = p.readMessage()
		if err != nil {
			break
		}
		d.ReceivePacket(pkt)
	}
}

func (h *Hub) portRun(p *Port) {
	var d IDevice

	resp, err := p.connect()
	if err != nil {
		goto disconnect
	}

	d = h.getDevice(resp)
	if d == nil {
		goto disconnect
	}

	// TODO save device to hub db

	p.run(d)

disconnect:
	p.disconnect()
}

func (h *Hub) Run() {
	go h.http()
	h.portScan()
}
