// Copyright 2021 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

function sendForThings() {
	conn.send(JSON.stringify({Msg: "GetThings"}))
}

function show(id) {
	var iframe = document.getElementById("thing")
	iframe.src = "/" + encodeURIComponent(id)
}

function getThing(msg) {
	if (msg.Things != null) {
		thing = msg.Things[0]
		show(thing.Id)
	}
}

function Run(scheme, host, id) {

	conn = new WebSocket(scheme + host + "/ws/" + id)

	conn.onopen = function(evt) {
		sendForThings()
	}

	conn.onclose = function(evt) {
		location.reload(true)
	}

	conn.onmessage = function(evt) {
		var msg = JSON.parse(evt.data)

		console.log('event', msg)

		switch(msg.Msg) {
		case "ReplyThings":
			getThing(msg)
			break
		case "SpamStatus":
			show(msg.Id)
			break
		}
	}
}
