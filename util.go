// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package merle

import (
	"net"
	"strings"
)

// Make up an id using the MAC address of the first non-lo interface
func defaultId() string {
	ifaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range ifaces {
			if iface.Name != "lo" {
				addr := iface.HardwareAddr.String()
				return strings.Replace(addr, ":", "_", -1)
			}
		}
	}
	return "unknown"
}

// A valid ID is a string with only [a-z], [A-Z], [0-9], or underscore
// characters.
func validId(s string) bool {
	for _, r := range s {
		if (r < 'a' || r > 'z') &&
			(r < 'A' || r > 'Z') &&
			(r < '0' || r > '9') &&
			(r != '_') {
			return false
		}
	}
	return true
}

func validModel(s string) bool { return validId(s) }
func validName(s string) bool  { return validId(s) }
