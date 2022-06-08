// Copyright 2021 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package telit

import (
	"fmt"
	"github.com/tarm/serial"
	"log"
	"strconv"
	"strings"
	"time"
)

type Telit struct {
	modem *serial.Port
}

func (t *Telit) modemCmd(cmd string) (string, error) {
	var buf = make([]byte, 128)
	var res []byte
	var err error

	t.modem.Flush()

	_, err = t.modem.Write([]byte(cmd))
	if err != nil {
		return "", err
	}

	for {
		var n int

		n, err = t.modem.Read(buf)
		if n == 0 { // timed-out; no more to read
			err = nil
			break
		}
		if err != nil {
			return "", err
		}
		res = append(res, buf[:n]...)
	}

	fields := strings.Fields(string(res))
	log.Printf("Telit modem response %q", fields)

	if len(fields) < 2 {
		return "", fmt.Errorf("Telit modem not enough fields returned: %s", fields)
	}

	if cmd[:len(cmd)-1] != fields[0] {
		return "", fmt.Errorf("Telit modem cmd not echo'ed: %s", fields)
	}

	if "OK" != fields[len(fields)-1] {
		return "", fmt.Errorf("Telit modem expected OK: %s", fields)
	}

	response := fields[len(fields)-2]

	return response, err
}

func (t *Telit) Init() error {
	var err error

	usb3 := &serial.Config{Name: "/dev/ttyUSB3", Baud: 115200,
		ReadTimeout: time.Second / 2}
	t.modem, err = serial.OpenPort(usb3)
	if err != nil {
		return err
	}

	// Wake up
	_, err = t.modemCmd("AT\r")
	if err != nil {
		return err
	}

	// Reset the GNSS parameters to "Factory Default" configuration
	_, err = t.modemCmd("AT$GPSRST\r")
	if err != nil {
		return err
	}

	// Delete the GPS information stored in NVM
	_, err = t.modemCmd("AT$GPSNVRAM=15,0\r")
	if err != nil {
		return err
	}

	// Start the GNSS receiver in standalone mode
	_, err = t.modemCmd("AT$GPSP=1\r")

	return err
}

func parseLatLong(loc string) float64 {
	dot := strings.Index(loc, ".")
	if dot == -1 {
		return 0.0
	}

	// TODO warning: probably fragile code below
	min := loc[dot-2 : len(loc)-1]
	deg := loc[0 : dot-2]
	dir := loc[len(loc)-1]

	minf, _ := strconv.ParseFloat(min, 64)
	degf, _ := strconv.ParseFloat(deg, 64)

	locf := degf + minf/60.0

	if dir == 'S' || dir == 'W' {
		locf = -locf
	}

	return locf
}

func (t *Telit) Location() (float64, float64) {
	acp, err := t.modemCmd("AT$GPSACP\r")
	if err != nil {
		log.Println(err)
		return 0, 0
	}
	loc := strings.Split(acp, ",")
	if len(loc) == 12 {
		lat := parseLatLong(loc[1])
		long := parseLatLong(loc[2])
		if lat != 0.0 {
			return lat, long
		}
	}
	return 0, 0
}
