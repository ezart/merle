# Merle

[Merle](https://merliot.org) is a framework for building secure web-apps for your IoT project.

Put your **Thing** on the **Internet** with Merle.

Merle uses the Go programming language.

## Set Up Your Environment

### Installing Go

Go is an open-source programming language that makes it easy to develop simple, reliable, and efficient software.

Go may already be installed on your distribution.  Try running ```go version``` to verify.

```sh
$ go version
go version go1.17.3 linux/amd64
```

If Go is not installed, follow the [official installation instructions](https://go.dev/doc/install) to get started.

## Installing Merle

With Go installed, the ```go get``` tool will help you install Merle and its required dependencies:

```sh
$ go get github.com/scottfeldman/merle
```

# Tutorial

This tutorial is broken up into multiple steps, each step building on the
previous.  The goal is to build a working, secure web-app available anywhere on
the Internet for your Thing.  In this tutorial, your Thing is a Raspberry Pi,
an LED, a resistor, and some wires.  We're going to make the LED blink and show
and control the LED status on the web-app.

![LED](examples/blink/assets/images/led-pgio17-off-small.png)
*Caption*

If you don't have the hardware needed for this tutorial, you can still run
through the tutorial.  There is no real LED to blink, so that's not very
exciting, but everything else should work.  All that's really needed is a
system with the Go environment installed.

## Step 1: Minimal Thing

This is the start of our Thing.  We'll call it blink.go.  It's basically the
smallest Thing you can make in Merle, but it will compile and run.  It doesn't
do anything, yet.

```go
// file: examples/tutorial/blinkv0/blink.go

package main

import (
	"github.com/scottfeldman/merle"
)

type blink struct {
}

func (b *blink) Subscribers() merle.Subscribers {
	return merle.Subscribers{}
}

func (b *blink) Assets() *merle.ThingAssets {
	return &merle.ThingAssets{}
}

func main() {
	var cfg merle.ThingConfig

	merle.NewThing(&blink{}, &cfg).Run()
}
```

A Thing in Merle is a Go program which implements and runs the Thinger
interface.  The Thinger interface has two methods: Subscribers and Assets.

```go
type Thinger interface {
	Subscribers() Subscribers
	Assets() *ThingAssets
}
```

Subscribers is a list of message handlers for your Thing.  We'll see later in
this tutorial that everything is a message in Merle, and Subscribers is the
message dispatcher.

Assets are the Thing's web assets, things like HTML and Javascript files.
These assets make up the front-end of your Thing (the side you see with a web
browser).

In our minimalist Thing, we don't (yet) subscribe to any messages and we don't
have any web assets.

Let's run our Thing and see what happens.  First, build Merle at the top level
to build the tutorial.

```sh
$ go install ./...
```

Then run our Thing:

```sh
$ ../go/bin/blinkv0
2022/01/24 17:57:26 Defaulting ID to 00:16:3e:30:e5:f5
2022/01/24 17:57:26 Skipping private HTTP server; port is zero
2022/01/24 17:57:26 Skipping public HTTP server; port is zero
2022/01/24 17:57:26 Skipping tunnel; missing host
[00:16:3e:30:e5:f5] Not handled: {"Msg":"_CmdRun"}
```

Ignore the "Skipping..." log messages for now.  Those are features we'll enable
in future steps.  The first thing to notice is the Thing was assigned an ID of
00:16:3e:30:e5:f5.  If that looks like a MAC address, you're right.  Every
Thing has an ID and since one wasn't given in the program, a default is
assigned, made up from a MAC address of one of the network interfaces on your
system.

The second thing to notice is the program quit.  It should not quit.  In this
case, the message CmdRun was not handled.  In the next step on this tutorial,
we'll handle the CmdRun message to blink the LED.

## Step 2: Blink the LED

Let's add a handler for CmdRun.  Every Thing should handle CmdRun.

```go
func (b *blink) run(p *merle.Packet) {
	select {}
}

func (b *blink) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdRun: b.run,
	}
}
```

The CmdRun handler should not exit unless some error occurs.  Otherwise, CmdRun
handler is the main loop for the Thing and is designed to run forever.  CmdRun
handler must not block.  select{} will not block or exit.  But we need to do
more than sleep so let's initialize the hardware and blink the LED.

We're using the excellent [GoBot](https://gobot.io) package to blink the LED
from the Raspberry Pi.  The CmdRun handler calls GoBot in Metal mode.


```go
import (
	"github.com/scottfeldman/merle"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
	"time"
)

func (b *blink) run(p *merle.Packet) {
	adaptor := raspi.NewAdaptor()
	adaptor.Connect()

	led := gpio.NewLedDriver(adaptor, "11")
	led.Start()

	for {
		led.Toggle()
		time.Sleep(time.Second)
	}
}
```
