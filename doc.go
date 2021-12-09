// Copyright 2021 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

/*
Merle is a "shortest stack" IoT framework.  The stack spans hardware access at
the bottom to html/javascript presentation at the top.

Status

Pre-alpha quatility of code here...

Installation

Install merle:

	go get github.com/scottfeldman/merle

Install merle devices:

	go get github.com/scottfeldman/merle_devices

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
Packets, where a Packet is a container for a []byte message payload.  The
IModel implemenation defines the contents of the messages and their symantics.
HomePage() is the device's html home page.

Use merle.NewDevice to create a new device from an IModel.

	model := ... // an IModel
	device := merle.NewDevice(model, ...)

Use Device.Run to start running the device.  Device.Run will 1) init and run
the IModel; 2) start an optional web server to serve up the device's home page;
and 3) optionally, create an SSH tunnel to a hub.

	device.Run(...)

The hub is an aggregator of devices.  Zero of more devices connect to the hub
over SSH tunnels, one SSH tunnel per device.  The hub runs it's own web server
and serves up the devices' home pages.  A hub can also aggregate other hubs.

*/
package merle
