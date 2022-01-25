# Overview

[Merle](https://merliot.org) is a framework for building secure web-apps for your IoT project.

Put your **Thing** on the **Internet** with Merle.

Merle uses the Go programming language.

## Set Up Your Environment

### Installing Go {#go}

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
go get github.com/scottfeldman/merle
```

# Tutorial

This tutorial is broken up into multiple steps, each step building on the
previous.  The goal is to build a working, secure web-app available anywhere on
the Internet for your Thing.  In this tutorial, your Thing is a Raspberry Pi,
an LED, a resistor, and some wires.  We're going to make the LED blink and show
and control the LED status on the web-app.

If you don't have the hardware needed for this tutorial, you can still run
through the tutorial.  There is no real LED to blink, so that's not very
exciting, but everything should work otherwise.  All that's really needed is a
system with the [Go](#go) environment installed.
