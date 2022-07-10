// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package merle

// Socket flags
const (
	sock_flag_bcast uint32 = 1 << iota
)

// socketer is an interface to a socket.  A socket plugs into a bus.
type socketer interface {
	// Send the Packet
	Send(*Packet) error
	// Close the socket
	Close()
	// Name of the socket
	Name() string
	// Socket flags
	Flags() uint32
	SetFlags(uint32)
	Src() string
}
