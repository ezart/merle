# Merle

[![Go Reference](https://pkg.go.dev/badge/pkg.dev.go/github.com/scottfeldman/merle.svg)](https://pkg.go.dev/github.com/scottfeldman/merle)

Merle is a "shortest stack" IoT framework.  An IoT device ("thing") plugs into
merle using a driver written specifically for the device.  Merle presents a
secure HTML user interface to the device, both locally at the device and
remotely at a device hub.  A device hub is an aggregator of devices.  Merle
uses websockets for all messaging between the device and the framework.
Websocket messaging between the device and hub is through a SSH tunnel.

Merle has a companion project called merle_devices.  Merle_devices is a libary
of device drivers for common IoT hardware configurations.

	https://github.com/scottfeldman/merle_devices

## Status

Alpha quatility of code here...

## Quickstart

See Quickstart in https://github.com/scottfeldman/merle_devices for building sample
devices for common IoT device hardware configurations.
