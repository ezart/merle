package merle

import (
	"log"
	"net"
)

func defaultId(id string) string {
	if id == "" {
		// Use the MAC address of the first non-lo interface
		ifaces, err := net.Interfaces()
		if err == nil {
			for _, iface := range ifaces {
				if iface.Name != "lo" {
					id = iface.HardwareAddr.String()
					log.Println("Defaulting ID to", id)
					break
				}
			}
		}
	}
	return id
}

func must(err error) error {
	if err != nil {
		log.Println(err)
	}
	return err
}


