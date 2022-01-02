# Merle

[![Go Reference](https://pkg.go.dev/badge/pkg.dev.go/github.com/scottfeldman/merle.svg)](https://pkg.go.dev/github.com/scottfeldman/merle)

Merle is an IoT framework.  With Merle, you can build a web app for your device
(the "Thing" in IoT).  A Web browser runs the app for monitoring and control of
your device, either over the internet or locally on the device itself.  To
build the app, a device model is created which plugs into the Merle framework.

The device model describes the device-specific view (web page) and behavior.
To model a device, you have two choices: write a new model for your device, or
re-use one of the existing models already created.  There is a library of
models written for Merle in a companion project merle_devices:

[Merle Devices](https://github.com/scottfeldman/merle_devices)













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


