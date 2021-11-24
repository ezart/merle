package merle

import (
	"database/sql"
	"encoding/json"
	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"sync"
	"time"
)

type Hub struct {
	sync.Mutex
	modelGen         func(model string) IModel
	devices          map[string]*Device
	conns            map[*websocket.Conn]bool
	db               *sql.DB
}

func NewHub(modelGen func(model string) IModel) *Hub {
	return &Hub{
		modelGen: modelGen,
		devices:  make(map[string]*Device),
		conns:    make(map[*websocket.Conn]bool),
	}
}

func (h *Hub) connAdd(c *websocket.Conn) {
	h.Lock()
	defer h.Unlock()
	h.conns[c] = true
}

func (h *Hub) connDelete(c *websocket.Conn) {
	h.Lock()
	defer h.Unlock()
	delete(h.conns, c)
}

func (h *Hub) broadcast(msg []byte) {
	var p = &Packet{
		Msg: msg,
	}

	h.Lock()
	defer h.Unlock()

	if len(h.conns) == 0 {
		log.Printf("Hub would broadcast: %s\n", msg)
		return
	}

	log.Printf("Hub broadcast: %s", msg)

	// TODO Perf optimization: use websocket.NewPreparedMessage
	// TODO to prepare msg once, and then sent on each socket

	for c := range h.conns {
		p.conn = c
		p.writeMessage()
	}
}

func (h *Hub) sendPacket(p *Packet) {
	log.Printf("Hub sendPacket: %s", p.Msg)
	err := p.writeMessage()
	if err != nil {
		log.Println("Hub sendPacket error:", err)
	}
}

func (h *Hub) deviceList() []MsgDevicesDevice {
	var devices []MsgDevicesDevice

	query := `SELECT id, model, name FROM devices ORDER BY firstseen`
	rows, err := h.db.Query(query)
	if err != nil {
		log.Println(err)
		return devices
	}
	defer rows.Close()

	for rows.Next() {
		var id, model, name string

		err = rows.Scan(&id, &model, &name)
		if err != nil {
			log.Println(err)
			return devices
		}

		d := h.getDevice(id)
		if d == nil {
			continue
		}

		devices = append(devices, MsgDevicesDevice{id,
			model, name, d.status})
	}

	return devices
}

func (h *Hub) receiveCmd(p *Packet) {
	var cmd MsgCmd

	json.Unmarshal(p.Msg, &cmd)

	switch cmd.Cmd {

	case CmdDevices:
		var resp = MsgDevicesResp{
			Type:    MsgTypeCmdResp,
			Cmd:     cmd.Cmd,
			Devices: h.deviceList(),
		}
		p.Msg, _ = json.Marshal(&resp)
		h.sendPacket(p)

	default:
		log.Printf("Unknown command \"%s\", skipping", cmd.Cmd)
	}
}

func (h *Hub) receivePacket(p *Packet) {
	var msg MsgType

	log.Printf("Hub receive: %s", p.Msg)
	json.Unmarshal(p.Msg, &msg)

	switch msg.Type {
	case MsgTypeCmd:
		h.receiveCmd(p)
	default:
		log.Printf("Unknown command type \"%s\", skipping", msg.Type)
	}
}

func (h *Hub) newDevice(id, model, name string, startupTime time.Time) *Device {
	m := h.modelGen(id)
	if m == nil {
		log.Printf("No device model named '%s'", model)
		return nil
	}

	d := NewDevice(m, id, model, name, "offline", startupTime)
	if d == nil {
		log.Printf("Device creation failed, model '%s'", model)
		return nil
	}

	h.devices[id] = d
	return d
}

func (h *Hub) getDevice(id string) *Device {
	if d, ok := h.devices[id]; ok {
		return d
	}
	return nil
}

func (h *Hub) saveDevice(d *Device) error {
	var err error
	var count int

	// TODO is there an SQL oneliner to do all the below?

	query := `SELECT COUNT(*) FROM devices WHERE id=?;`
	err = h.db.QueryRow(query, d.id).Scan(&count)
	if err != nil {
		return err
	}

	switch count {
	case 0:
		insert := `INSERT INTO devices (id, model, name, firstseen) VALUES(?, ?, ?, ?);`
		_, err = h.db.Exec(insert, d.id, d.model, d.name, time.Now())
	case 1:
		update := `UPDATE devices SET model=?, name=? WHERE id=?;`
		_, err = h.db.Exec(update, d.model, d.name, d.id)
	}

	return err
}

func (h *Hub) createTable(table string, spec string) error {
	create := `CREATE TABLE IF NOT EXISTS ` + table + ` ` + spec + `;`
	_, err := h.db.Exec(create)
	return err
}

var tableSpecs = map[string]string{
	"devices": `(id TEXT PRIMARY KEY, model TEXT, name TEXT, firstseen TIMESTAMP)`,
}

func (h *Hub) openDB() error {
	var err error

	h.db, err = sql.Open("sqlite3", "./dbs/hub.db")
	if err != nil {
		return err
	}

	for table, spec := range tableSpecs {
		err := h.createTable(table, spec)
		if err != nil {
			return err
		}
	}

	query := `SELECT id, model, name FROM devices ORDER BY firstseen`
	rows, err := h.db.Query(query)
	if err != nil {
		log.Printf("Hub query failed: %#v", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id, model, name string

		err = rows.Scan(&id, &model, &name)
		if err != nil {
			log.Println(err)
			return err
		}
		d := h.newDevice(id, model, name, time.Time{})
		if d == nil {
			continue
		}
		log.Printf("Hub new device %s, %s, %s", id, model, name)
	}

	return nil
}

/*
func (h *Hub) changeStatus(d *Device, status string) {
	d.status = status

	spam := MsgStatusSpam{
		Type:   MsgTypeSpam,
		Spam:   "Status",
		Id:     d.id,
		Model:  d.model,
		Name:   d.name,
		Status: d.status,
	}

	msg, _ := json.Marshal(&spam)
	h.broadcast(msg)
}
*/

func (h *Hub) portRun(p *Port) {
	var d *Device

	resp, err := p.connect()
	if err != nil {
		goto disconnect
	}

	d = h.getDevice(resp.Id)
	if d == nil {
		d = h.newDevice(resp.Id, resp.Model, resp.Name, resp.StartupTime)
		if d == nil {
			goto disconnect
		}
	}

	err = h.saveDevice(d)
	if err != nil {
		goto disconnect
	}

	//h.changeStatus(d, "online")
	//p.run(d.IDevice)
	//h.changeStatus(d, "offline")

disconnect:
	p.disconnect()
}

func (h *Hub) Run() {
	err := h.openDB()
	if err != nil {
		return
	}
	go h.http()
	h.portScan()
}
