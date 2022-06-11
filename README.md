# Merle

[![Go Reference](https://pkg.go.dev/badge/pkg.dev.go/github.com/merliot/merle.svg)](https://pkg.go.dev/github.com/merliot/merle)
[![Go Report Card](https://goreportcard.com/badge/github.com/merliot/merle)](https://goreportcard.com/report/github.com/merliot/merle)

[Merle](https://merliot.org) is a framework for building secure web-apps for your IoT project.

Put your **Thing** on the **Internet** with Merle.

Merle uses the Go programming language.

## TODO

 - Thing-wide setting for a websocket ping/pong messages to close half-open sockets
 	(send ping from Thing side of websocket)
 - Can we use something smaller like a container to run Thing Prime rather than a full VM?

## Set Up Your Environment

### Installing Go

Go is an open-source programming language that makes it easy to develop simple, reliable, and efficient software.

Go may already be installed on your distribution.  Try running ```go version``` to verify.

```sh
$ go version
go version x.xx.x
```

If Go is not installed, follow the [official installation instructions](https://go.dev/doc/install) to get started.

## Installing Merle

With Go installed, the ```go get``` tool will help you install Merle and its required dependencies:

```sh
$ go get github.com/scottfeldman/merle
```

## Writing Your First Thing

Once you have the Merle package installed, you're ready to start writing your own code. The first program we are going to create is the "Hello, World" of things, which is a web-app that shows "Hello, World!" when viewed with a browser.

### Hello, World!

```go
// file: hello_world.go

package main

import (
	"github.com/scottfeldman/merle"
)

type hello struct {
}

func (h *hello) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdRun: merle.RunForever,
	}
}

func (h *hello) Assets() *merle.ThingAssets {
	return &merle.ThingAssets{
		TemplateText: "Hello, World!\n",
	}
}

func main() {
	var cfg merle.ThingConfig
	
	cfg.Thing.PortPublic = 8080

	merle.NewThing(&hello{}, &cfg).Run()
}
```

Let's make a new directory:

```sh
$ mkdir hello_world
$ cd hello_world
```

Copy the above Thing code to hello_world.go and initialize the go module:

```sh
$ go mod init hello_world
$ go mod tidy
```

Now run hello_world:

```sh
$ go run hello_world.go
Defaulting ID to 00_16_3e_30_e5_f5
Skipping private HTTP server; port is zero
Public HTTP server listening on :80
Skipping public HTTPS server; port is zero
Skipping tunnel; missing host
[00_16_3e_30_e5_f5] Received: {"Msg":"_CmdRun"}
```

In another shell, view the Thing's web output:

```sh
$ curl localhost:8080
Hello, World!
```

## Architecture

2000 words

<img src="https://docs.google.com/drawings/d/e/2PACX-1vSkx75Ta5MePFXAM_O1C5voMNJ8aguUg8ahdgCNCw9MTpOkI3wgeFrcEUpYfoN0-_OFyQe37uAmVnRk/pub?w=1400&amp;h=580">

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
