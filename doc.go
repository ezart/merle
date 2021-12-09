// Copyright 2021 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

/*
Merle is a "shortest stack" IoT framework.  The stack spans hardware access at
the bottom to html presentation at the top.  Merle uses websockets for
messaging.


Status

Alpha quatility of code here...

Installation

Merle is two packages: core and devices.  Install the core package from here:

	go get github.com/scottfeldman/merle

The devices package contains a library of devices.  Install the merle devices
package from here:

	go get github.com/scottfeldman/merle_devices

Quickstart

See Quickstart in github.com/scottfeldman/merle_devices for building sample
devices for common IoT device hardware configurations.

Overview

A device in Merle is the "Thing" in IoT.  A device is modeled with an IModel
interface.

	type IModel interface {
		Init(*Device) error
		Run()
		Receive(*Packet)
		HomePage(http.ResponseWriter, *http.Request)
	}

An implementation of IModel is the "device driver" for the device.  Init()
initializes the device hardware and structures.  Run() runs the main loop for
the device.  The device stops when Run() exits.  Receive() handles incoming
Packets, where a Packet is a container for a []byte message.  The IModel
implemenation defines the contents of the messages and their symantics.
HomePage() is the device's html home page.

Use merle.NewDevice to create a new device from an IModel.

	model := ... // an IModel
	device := NewDevice(model, ...)

Use Device.Run to start running the device.  Device.Run will

1) Init() and Run() the IModel.
2) Start an optional http server to serve up the device's home page and a
websocket interface to the device.
3) Optionally, create an SSH tunnel to a hub.

	device.Run(...)

The hub is an aggregator of devices.  Zero of more devices connect to the hub
over SSH tunnels, one SSH tunnel per device.  The hub runs it's own http server
and serves up the devices' home pages.  A hub can also aggregate other hubs.
To create a new hub, use merle.NewHub:

	hub := NewHub(...)

And to run the hub, use:

	hub.Run()

Messaging

Merle uses websockets for messaging.  The device's http server serves up a
websocket interface.  A client opens a websocket connection to the device.  The
websocket connection persists until the client disconnects and allows
bi-directional messaging from/to the device and to/from the client.  Since
websocket is built on TCP, the connection is reliable.  A client could be a
hub, or a client could be the device's own home page, using Javascript to open
a websocket back to the device.

The device's http server serves the websocket interface on a public port and a
private port.  Regardless of the port, the websocket address is:

	ws://<host>:<port>/ws

Public port access is gated by http Basic Authentication.  More about this in
the security section.

Merle defines a few constraints on the message format, but otherwise the
message content is defined by the device's IModel.  Although websockets
supports both binary and text messages, merle uses text message with a JSON
enconding.  A JSON message in merle uses the base structure MsgType:

	type MsgType struct {
		Type string
		// payload
	}

Where Type is one of:

	const (
		MsgTypeCmd     = "cmd"
		MsgTypeCmdResp = "resp"
		MsgTypeSpam    = "spam"
	)

A device can make new message structures building on MsgType:

	type msgMyCmd struct {
		Type  string
		Cmd   string
		Value int
	}

The device receives messages on IModel:Receive(p *Packet).  The message is
contained in p.Msg.  The device will JSON Unmarshal the p.Msg message, process
the message, and optionally send a reply.  For example, using the msgMyCmd
structure, handle getting or setting a device value:

	func (m *MyModel) Receive(p *Packet) {
		var msg msgMyCmd
		json.Unmarshal(p.Msg, &msg)
		switch msg.Cmd {
		case "Put":
			// save msg.Value to device
		case "Get":
			msg.Value = // get from device
			p.Msg, _ = json.Marshal(&msg)
			m.device.Reply(p)
		}
	}

The device can broadcast a message using device.Broadcast(&p).  This will
broadcast the message to all of the websocket clients.

Security

blah, blah, blah

*/
package merle
