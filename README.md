# Merle

[![Go Reference](https://pkg.go.dev/badge/pkg.dev.go/github.com/merliot/merle.svg)](https://pkg.go.dev/github.com/merliot/merle)
[![Go Report Card](https://goreportcard.com/badge/github.com/merliot/merle)](https://goreportcard.com/report/github.com/merliot/merle)

[Merle](https://merliot.org) is a framework for building secure web-apps for your IoT project.

Put your **Thing** on the **Internet** with Merle.

Merle uses the Go programming language.

<img src="https://docs.google.com/drawings/d/e/2PACX-1vSkx75Ta5MePFXAM_O1C5voMNJ8aguUg8ahdgCNCw9MTpOkI3wgeFrcEUpYfoN0-_OFyQe37uAmVnRk/pub?w=1400&amp;h=580">

## Documentation

Find documentation [here](https://pkg.go.dev/github.com/scottfeldman/merle)

## Need help?
* Issues: [https://github.com/scottfeldman/merle/issues](https://github.com/scottfeldman/merle/issues)
* Mailing list: [https://groups.google.com/g/merle-io](https://groups.google.com/g/merle-io)

## Contributing
For contribution guidelines, please go to [here](https://github.com/scottfeldman/merle/blob/main/CONTRIBUTING.md).

## License
Licensed under the [BSD 3-Clause License](https://github.com/scottfeldman/merle/blob/main/LICENSE)

Copyright (c) 2021-2022 Scott Feldman (sfeldma@gmail.com).

## TODO

My TODO list, not in any particular order.

 - Thing-wide setting for a websocket ping/pong messages to close half-open sockets
 	(send ping from Thing side of websocket)
 - Can we use something smaller like a container to run Thing Prime rather than a full VM?
 - favicon.ico support?  just add a /favicon.ico file?
 - Investigate if Merle framework could be written in other languages (Rust?).
   Assests (js/html/etc) wouldn't need to change.  Thing code would be rewritten in new language.
   A Thing written in one language should interoperate with another Thing written in another language?
 - [tunnel.go] Need to use golang.org/x/crypto/ssh instead of os/exec'ing ssh calls.  Also, look
   into using golang.org/x/crypto/ssh on hub-side of merle for bespoke ssh server.
 - More tests!
