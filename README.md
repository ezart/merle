# Merle

[![Go Reference](https://pkg.go.dev/badge/pkg.dev.go/github.com/scottfeldman/merle.svg)](https://pkg.go.dev/github.com/scottfeldman/merle)

Merle is a framework for building secure web-apps for your IoT project.

Put your **Thing** on the **Internet** with Merle.

Merle uses the Go programming language and javascript.

## Installation

```sh
go get github.com/scottfeldman/merle
```

## Architecture

2000 words

<img src="https://docs.google.com/drawings/d/e/2PACX-1vSkx75Ta5MePFXAM_O1C5voMNJ8aguUg8ahdgCNCw9MTpOkI3wgeFrcEUpYfoN0-_OFyQe37uAmVnRk/pub?w=1400&amp;h=580">

## Quick Start, Part I

Merle includes a library of [Things](things/README.md) already built and tested.  For this quick start, let's pick a quintessential one: a Raspberry Pi LED blinker[^1].  Here's the hardware setup:

![raspi_blink](web/images/raspi_blink/led-gpio17-off-small.png)

Hardware list:
- Rapsberry Pi (any model except Pico)
- An LED
- A 120ohm resistor
- some wire.

Wire the LED and resistor to GPIO pin 17 and ground as shown.

**Don't worry if you don't have the hardware on hand; we can run the Thing in demo-mode to similate the hardare.  All that's need for demo-mode is a system with [Go](https://go.dev/) installed.**

Install Merle, if you haven't already:

```sh
go get github.com/scottfeldman/merle
```

Build Merle:

```sh
go install ./...
```

Before we run Merle on your Thing, we need to configure Merle for your Thing.  Merle gets the Thing configuration from /etc/merle/thing.yml.  As sudo, let's create and edit the /etc/merle/thing.yml file.

```sh
sudo mkdir /etc/merle
sudo vi /etc/merle/thing.yml
```

Add this to /etc/merle/thing.yml:

```yaml
# Thing configuration
Thing:
  Model: raspi_blink
  Name: quickstart
  PortPublic: 80
```

Now start Merle on your Thing:

````sh
sudo ../go/bin/merle-thing
````

Or, for demo mode, add --demo:


````sh
sudo ../go/bin/merle-thing --demo
````

The hardware LED should blink on/off every second.

Open a web browser to localhost and see your Thing running!  Click the button to pause and resume the LED blinking.  

![raspi_blink](web/images/raspi_blink/led-gpio17-animation.gif?raw=true)

Notice the LED state is always synced between the hardware LED and the LED shown in the browser.  Open another browser window to localhost.  Now both browsers and the hardware LEDs are synced.  This is the first principle of Merle:

### Principle #1: The Thing is the truth and all views of the Thing hold this truth.

Code for this Raspberry Pi LED blinker is in two parts:
  - Back-end: [Thing code](things/raspi_blink/raspi_blink.go)
  - Front-end: [HTML](web/templates/raspi_blink.html), [Javascript](web/js/raspi_blink.js), and [CSS](web/css/raspi_blink.css)

Wait, localhost is not the Internet!  How do I get my Thing on the Internet?

Continue on to [Quick Start, Part II](README-QS2.md) to learn how put your device on the Internet.

## Documentation

Find documentation [here](https://pkg.go.dev/github.com/scottfeldman/merle)

## Need help?
* Issues: [https://github.com/scottfeldman/merle/issues](https://github.com/scottfeldman/merle/issues)
* Mailing list: [https://groups.google.com/g/merle-io](https://groups.google.com/g/merle-io)

## Contributing
For contribution guidelines, please go to [here](https://github.com/scottfeldman/merle/blob/main/CONTRIBUTING.md).

## License
Copyright (c) 2021-2022 Scott Feldman (sfeldma@gmail.com).  Licensed under the [BSD 3-Clause License](https://github.com/scottfeldman/merle/blob/main/LICENSE)

[^1]: This Thing was built using the excellent robotics library [GoBot](https://gobot.io) for hardware access.
