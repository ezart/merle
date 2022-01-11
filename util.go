package merle

import (
	"log"
	"net"
)

// If no id, make up an id using the MAC address of the first non-lo interface
func defaultId(id string) string {
	if id == "" {
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
