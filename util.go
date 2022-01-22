package merle

import (
	"net"
)

// Make up an id using the MAC address of the first non-lo interface
func defaultId() string {
	ifaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range ifaces {
			if iface.Name != "lo" {
				return iface.HardwareAddr.String()
			}
		}
	}
	return "unknown"
}
