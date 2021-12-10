# Merle

[![Go Reference](https://pkg.go.dev/badge/pkg.dev.go/github.com/scottfeldman/merle.svg)](https://pkg.go.dev/github.com/scottfeldman/merle)

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

See Quickstart in https://github.com/scottfeldman/merle_devices for building sample
devices for common IoT device hardware configurations.
