# Merle

[![Go Reference](https://pkg.go.dev/badge/pkg.dev.go/github.com/merliot/merle.svg)](https://pkg.go.dev/github.com/merliot/merle)
[![Go Report Card](https://goreportcard.com/badge/github.com/merliot/merle)](https://goreportcard.com/report/github.com/merliot/merle)

[Merle](https://merliot.org) is a framework for building secure web-apps for your IoT project.

Put your **Thing** on the **Internet** with Merle.

Merle uses the Go programming language.

***Warning: Early-Beta Software - not for production use***

![Gopher Thing](gopher_cloud.png)

## Documentation

[Project Web Page](https://merliot.org)

- Find Getting Started, Tutorial, Examples and useful Guides.

[Code Reference](https://pkg.go.dev/github.com/merliot/merle)

## Hello, world!

```
package main

import (
        "log"

        "github.com/merliot/merle"
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
                HtmlTemplateText: "Hello, world!\n",
        }
}

func main() {
        thing := merle.NewThing(&hello{})
        thing.Cfg.PortPublic = 80
        log.Fatalln(thing.Run())
}
```

## Need help?
* Issues: [https://github.com/merliot/merle/issues](https://github.com/merliot/merle/issues)
* Gophers Slack channel: [https://gophers.slack.com/messages/merliot/](https://gophers.slack.com/messages/merliot/)

## Contributing
For contribution guidelines, please go to [here](https://github.com/merliot/merle/blob/main/CONTRIBUTING.md).

## License
Licensed under the [BSD 3-Clause License](https://github.com/merliot/merle/blob/main/LICENSE)

Copyright (c) 2021-2022 Scott Feldman (sfeldma@gmail.com).

## TODO

My TODO list, not in any particular order.  (Help would be much appreciated).

 - Thing-wide setting for a websocket ping/pong messages to close half-open sockets
 	(send ping from Thing side of websocket)
 - Can we use something smaller like a container to run Thing Prime rather than a full VM?
 - favicon.ico support?  just add a /favicon.ico file?
 - Investigate if Merle framework could be written in other languages (Rust?).
   Assests (js/html/etc) wouldn't need to change.  Thing code would be rewritten in new language.
   A Thing written in one language should interoperate with another Thing written in another language?
 - More tests!
 - I'm not a Security expert.  Need review of Merle by some who are.
 - I'm not a JavaScript/HTML expert.  Need review of Merle by experts.
