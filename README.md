# merle

Merle is a "shortest stack" IoT framework.  The stack spans hardware access at
the bottom to html presentation at the top.  Merle uses websockets for
messaging.

## Status

Alpha quatility of code here...

## Installation

Merle comprises two packages: core and devices.  Install the core package from here:

```go
go get github.com/scottfeldman/merle
```

The devices package contains a library of devices.  Install the merle devices
package from here:

```go
go get github.com/scottfeldman/merle_devices
```

## Quickstart

See Quickstart in github.com/scottfeldman/merle_devices for building sample
devices for common IoT device hardware configurations.

## Overview

A device in Merle is the "Thing" in IoT.  A device is modeled with an IModel
interface.

```go
type IModel interface {
	Init(*Device) error
	Run()
	Receive(*Packet)
	HomePage(http.ResponseWriter, *http.Request)
}
```

An implementation of IModel is the "device driver" for the device.  Init()
initializes the device hardware and structures.  Run() runs the main loop for
the device.  The device stops when Run() exits.  Receive() handles incoming
Packets, where a Packet is a container for a []byte message.  The IModel
implemenation defines the contents of the messages and their symantics.
HomePage() is the device's html home page.

Use merle.NewDevice to create a new device from an IModel.

```go
model := ... // an IModel
device := NewDevice(model, ...)
```

Use Device.Run to start running the device.

```go
device.Run(...)
```

Device.Run will:

```go
1) Init() and Run() the IModel.
2) Start an optional http server to serve up the device's home page and a
websocket interface to the device.
3) Optionally, create an SSH tunnel to a hub.
```

The hub is an aggregator of devices.  Zero of more devices connect to the hub
over SSH tunnels, one SSH tunnel per device.  The hub runs it's own http server
and serves up the devices' home pages.  A hub can also aggregate other hubs.
To create a new hub, use merle.NewHub:

```go
hub := NewHub(...)
```

And to run the hub, use:

```go
hub.Run()
```

## Messaging

Merle uses websockets for messaging.  The device's http server serves up a
websocket interface.  A client opens a websocket connection to the device to
communicate with the device.  The websocket connection persists until the
client disconnects.  The connection allows bi-directional messaging from/to the
device and to/from the client.  Since websockets is built on TCP, the
connection is reliable.

One client is the device's own home page, using Javascript to open a websocket
back to the device.  Another client is the device running inside a hub (more
about this in a bit), opening a websocket back to the device running on real
hardware.  Yet another client is the device running inside a hub's own home
page, using the same Javascript as before to open a websocket back to the
device (running inside the hub).  Confused?  A picture would be worth 1000
words here...

The device's http server serves the websocket interface on a public port and a
private port.  Regardless of the port, the websocket address is:

```go
ws://<host>:<port>/ws
```

Public port access is gated by http Basic Authentication.  More about this in
the security section.

Merle defines a few constraints on the message format, but otherwise the
message content is defined by the device's IModel.  Although websockets
supports both binary and text messages, merle uses text message with a JSON
enconding.  A JSON message in merle uses the base structure MsgType:

```go
type MsgType struct {
	Type string
	// payload
}
```

Where Type is one of:

```go
const (
	MsgTypeCmd     = "cmd"
	MsgTypeCmdResp = "resp"
	MsgTypeSpam    = "spam"
)
```

A device can make new message structures building on MsgType:

```go
type msgMyCmd struct {
	Type  string
	Cmd   string
	Value int
}
```

The device receives messages on IModel:Receive(p *Packet).  The message is
contained in p.Msg.  The device will JSON Unmarshal the p.Msg message, process
the message, and optionally send a reply.  For example, using the msgMyCmd
structure, handle getting or setting a device value:

```go
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
```

The device can broadcast a message using device.Broadcast(&p).  This will
broadcast the message to all of the websocket clients.

## Device Duality

As hinted to above, the device can run in two different modes.  The first mode
is when the device runs on the real hardware, accessing hardware and serving up
a websocket interface.  The second mode is when the device runs inside a hub.
This is same device code as before, but instead of direct hardware accesses,
the hub device is a proxy for the real device, talking to the real device over
a private websocket connection.  So here the hub device becomes a websocket
client of the real device.

The hub device serves up it's own websocket interface as well, running the same
"protocol" as the real device.  The device's home page can invoke a javascript
websocket, which connects back to the device, regardless of the device's mode
(real or hub).  In this case, the javascript websocket is the websocket client.

Hub mode is used again when we have a hub of hubs.  Or a hub of hubs of hubs.
Here, the hub device on the top-most hub is a proxy for the hub device below,
and so on, until we have a proxy for the real device.

To summerize, the different websocket client connections are:

```go
1) javascript websocket connection to real device
2) javascript websocket connection to hub device
3) hub device to real device
4) hub device to hub device
```

Each connection should be consider when designing the device's protocol.

## Defining the Protocol

The rules that determine what and how data is transfered between the device and
a client is the device's protocol.  Each device modeled with IModel defines a
protocol.  The protocol uses the messaging format described above, transmitted
over the websocket between the device and the client.

By example, we can look at the protocol for the Model60 device from the
merle_devices project.  A Model60 device has a temperature and humidity sensor
as well as GPS locator.  The device's home page shows a history graph over the
last 60 minutes of temperature and humidity, and a GPS location in lat/long
coordinates.  On startup, a client will ask for the current GPS location and
the last 60 minutes worth of temperature/humidty data.  Periodically, the
device will send asynchronous messages to update the current GPS location or
temperature/humidity.  The asynchromous messages are called spam.
Additionally, the user can press a restart button on the home page which will
generate a message down to the device to restart the device.

We have to consider all clients when defining the protocol (see above), so
let's start with

1) javascript websocket connection to real device

```go
Javascript (model60.js)		Real Device (model60.go)
===============================================================

START:
--------

cmd:Identify   ------------>
                                get identity
               <------------    resp:Identify
save identity
cmd:Location   ------------>
                                get location from store
               <------------    resp:Location
save location
cmd:Temps      ------------>
                                get temps from store
               <------------    resp:Temps
save temps

RUNNING:
--------

                                new temp/humidity from device
                                save temp/humidity to store
                <------------   spam:Temp (Broadcast)
update temps

                                new location from device
                                save location to store
                <------------   spam:Temp (Broadcast)
update location

presses restart button
cmd:Restart     ------------>
                                restart device
```

The spam broadcasts broadcast to all the connected clients.  For example,
multiple web clients viewing the device's home page would open a websocket
connection back to the device for each client.

## Security

blah, blah, blah

---
Readme created from Go doc with [goreadme](https://github.com/posener/goreadme)
