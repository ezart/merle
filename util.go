// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package merle

import (
	"bytes"
	"encoding/json"
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

func prettyPrintJSON(msg []byte) string {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, msg, "", "    "); err != nil {
		return ""
	}
	return prettyJSON.String()
}

