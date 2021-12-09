# merle

Merle is a "shortest stack" IoT framework.  The stack spans hardware access at
the bottom to html/javascript presentation at the top.

## Status

Pre-alpha quatility of code here...

## Installation

Install merle:

```go
go get github.com/scottfeldman/merle
```

Install merle devices:

```go
go get github.com/scottfeldman/merle_devices
```

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
Packets, where a Packet is a container for a []byte message payload.  The
IModel implemenation defines the contents of the messages and their symantics.
HomePage() is the device's html home page.

Use merle.NewDevice to create a new device from an IModel.

```go
model := ... // an IModel
device := merle.NewDevice(model, ...)
```

Use Device.Run to start running the device.  Device.Run will 1) init and run
the IModel; 2) start an optional web server to serve up the device's home page;
and 3) optionally, create an SSH tunnel to a hub.

```go
device.Run(...)
```

The hub is an aggregator of devices.  Zero of more devices connect to the hub
over SSH tunnels, one SSH tunnel per device.  The hub runs it's own web server
and serves up the devices' home pages.  A hub can also aggregate other hubs.

## Constants

```golang
const (
    MsgTypeCmd     = "cmd"
    MsgTypeCmdResp = "resp"
    MsgTypeSpam    = "spam"
)
```

```golang
const (
    CmdIdentify = "Identify"
    CmdStart    = "Start"
    CmdDevices  = "Devices"
)
```

## Functions

### func [DefaultId](/device.go#L333)

`func DefaultId() string`

DefaultId returns a default ID based on the device's MAC address

## Types

### type [Device](/device.go#L69)

`type Device struct { ... }`

Device runs an IModel, either in device-mode or hub-mode.  In device-mode,
the Device runs locally, with direct access to the device hardware.  In
hub-mode, the Device runs inside a Hub, connecting to the device-mode Device
over a tunnel using websockets.  The IModel implements both modes.

## Device-mode

In device-mode, the Device runs the IModel, starts two web servers, and
(optionally) creates a tunnel to a Hub.

The first web server listens on port :80 for HTTP and serves up the Device
homepage on [http://localhost/](http://localhost/) and a websocket connection to the Device on
ws://localhost/ws.  Basic Authentication protects access to both http:// and
ws://.  The only allowed Basic Authentication user is authUser passed to
Device.Run(...).

The second web server listens on port :8080 for HTTP and serves up websocket
connection to the Device on ws://localhost:8080/ws.

The optional tunnel is a SSH remote port forwarding tunnel where the Device
is the ssh client and the Hub is the SSH server.  To create the tunnel,
first the Device requests, over SSH, a remote port from the Hub.  The Device
then creates a SSH remote port forwarding tunnel mapping <remote
port>:localhost:8080.  The Hub can now create a websocket connection back to
the device-mode Device using ws://localhost:<remote port>/ws.

## Hub-mode

In hub-mode, the Device runs the IModel inside a Hub.  The Hub will
websocket connect back to the device-mode Device, also running the same
IModel.  See type Hub for more information.

#### func [NewDevice](/device.go#L94)

`func NewDevice(m IModel, inHub bool, id, model, name, status string,
    startupTime time.Time) *Device`

NewDevice returns a new Device.

```go
m is IModel instance.
inHub is true if Device is running in hub-mode.
id is ID of Device.  id is unique for each Device in a Hub.  If id is "",
a default ID is assigned.
model is the name of the model.
name is the name of the Device.
status is status of the Device, e.g. "online", "offline".
startupTime is the Device's startup time.
```

#### func (*Device) [Broadcast](/device.go#L225)

`func (d *Device) Broadcast(p *Packet)`

Broadcast packet to all websocket connections on the Device, except self.

#### func (*Device) [HomeParams](/device.go#L146)

`func (d *Device) HomeParams(r *http.Request) *homeParams`

HomeParams returns useful parameters to be passed to IModel.HomePage's
html template.

#### func (*Device) [Id](/device.go#L116)

`func (d *Device) Id() string`

Return the Device ID

#### func (*Device) [InHub](/device.go#L131)

`func (d *Device) InHub() bool`

Return true if Device running in Hub

#### func (*Device) [Model](/device.go#L121)

`func (d *Device) Model() string`

Return the Device model

#### func (*Device) [Name](/device.go#L126)

`func (d *Device) Name() string`

Return the Device name

#### func (*Device) [Reply](/device.go#L176)

`func (d *Device) Reply(p *Packet)`

Reply sends Packet back to originator of a Packet received with
IModel.Receive.

#### func (*Device) [Run](/device.go#L312)

`func (d *Device) Run(authUser, hubHost, hubUser, hubKey string,
    publicPort, privatePort int) error`

Run the Device.  Run should not be called on a Device in Hub.  Run will
initialize the IModel, create a tunnel, start the http servers, and then run
the IModel.

```go
authUser is the valid user for Basic Authentication of the public http
server.
hubHost is URL for the Hub host.  If blank, Device will not connect to Hub.
hubUser is the Hub SSH user.
hubKey is the Hub SSH key.
publicPort is the public http server listening port
privatePort is the private http server listening port
```

#### func (*Device) [Sink](/device.go#L191)

`func (d *Device) Sink(p *Packet)`

Sink sends Packet towards device-mode Device.  The Packet is not sunk if
Device is not in hub-mode.  The Packet is not sunk if it came thru the port
from the device-mode Device.

### type [Hub](/hub.go#L18)

`type Hub struct { ... }`

#### func [NewHub](/hub.go#L27)

`func NewHub(modelGen func(model string) IModel, templ string) *Hub`

#### func (*Hub) [Run](/hub.go#L306)

`func (h *Hub) Run()`

### type [IModel](/device.go#L19)

`type IModel interface { ... }`

IModel is the business logic of a Device, specifying a "device driver" interface.

### type [MsgCmd](/msg.go#L21)

`type MsgCmd struct { ... }`

### type [MsgCmdResp](/msg.go#L27)

`type MsgCmdResp struct { ... }`

### type [MsgDevicesDevice](/msg.go#L55)

`type MsgDevicesDevice struct { ... }`

### type [MsgDevicesResp](/msg.go#L62)

`type MsgDevicesResp struct { ... }`

### type [MsgIdentifyResp](/msg.go#L45)

`type MsgIdentifyResp struct { ... }`

### type [MsgSpam](/msg.go#L33)

`type MsgSpam struct { ... }`

### type [MsgStatusSpam](/msg.go#L68)

`type MsgStatusSpam struct { ... }`

### type [MsgType](/msg.go#L17)

`type MsgType struct { ... }`

### type [Packet](/packet.go#L13)

`type Packet struct { ... }`

A Packet contains a message and a (hidden) source.

---
Readme created from Go doc with [goreadme](https://github.com/posener/goreadme)
