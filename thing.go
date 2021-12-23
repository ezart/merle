package merle

import (
	"log"
	"time"
	"net"
)

type Thing struct {
	Init func() error
	Run func()

	Status        string
	Id            string
	Model         string
	Name          string
	StartupTime   time.Time
}

func (t *Thing) prefix() string {
	return "["+t.Id+","+t.Model+","+t.Name+"]"
}

func (t *Thing) Start() {
	if t.Init != nil {
		log.Println(t.prefix(), "Init...")
		if err := t.Init(); err != nil {
			log.Fatalln(t.prefix(), "Init failed:", err)
		}
	}
	if t.Run != nil {
		log.Println(t.prefix(), "Run...")
		t.Run()
	}
	log.Fatalln(t.prefix(), "Run() didn't run forever")
}

// DefaultId returns a default ID based on the device's MAC address
func DefaultId_() string {

	// Use the MAC address of the first non-lo interface
	// as the default device ID

	ifaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range ifaces {
			if iface.Name != "lo" {
				return iface.HardwareAddr.String()
			}
		}
	}

	return ""
}
