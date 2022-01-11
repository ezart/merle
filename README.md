# Merle

[![Go Reference](https://pkg.go.dev/badge/pkg.dev.go/github.com/scottfeldman/merle.svg)](https://pkg.go.dev/github.com/scottfeldman/merle)

Merle is a framework for building web-apps for your IoT project.  Put your Thing on the Internet using Merle.


Merle uses the Go programming language and javascript.

## Architecture

<img src="https://docs.google.com/drawings/d/e/2PACX-1vSkx75Ta5MePFXAM_O1C5voMNJ8aguUg8ahdgCNCw9MTpOkI3wgeFrcEUpYfoN0-_OFyQe37uAmVnRk/pub?w=419&amp;h=424">

## Installation

```sh
go get github.com/scottfeldman/merle
```

## Quick Start

Merle includes a library of Things already built and tested.  Let's pick a quintessential one for the Quickstart: a Raspberry Pi LED blinker.  Here's the hardware setup:

![foo](web/images/raspi_blink/led_gpio_off.png?raw=true)

Don't worry if you don't have the hardware setup on hand; we'll run the Thing in demo-mode first to similate the hardare accesses.  You can turn off demo mode and run on the real hardware setup if you have that ready. 

## Documentation

Find documentation [here](https://pkg.go.dev/github.com/scottfeldman/merle)

## Need help?
* Issues: [https://github.com/scottfeldman/merle/issues](https://github.com/scottfeldman/merle/issues)
* Mailing list: [https://groups.google.com/g/merle-io](https://groups.google.com/g/merle-io)

## Contributing
For contribution guidelines, please go to [here](https://github.com/scottfeldman/merle/blob/main/CONTRIBUTING.md).

## License
Copyright (c) 2021-2022 Scott Feldman (sfeldma@gmail.com).  Licensed under the [BSD 3-Clause License](https://github.com/scottfeldman/merle/blob/main/LICENSE)
