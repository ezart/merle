// Copyright 2021 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

let status

function sendForThings() {
	conn.send(JSON.stringify({Msg: "GetThings"}))
}

function addThing(thing) {
}

function saveThings(msg) {
	if (msg.Things != null) {
		for (const thing of msg.Things) {
			console.log('thing', thing)
			addThing(thing)
		}
	}
}

function updateStatus(msg) {
	status = msg.Status
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
		case "RespThings":
			saveThings(msg)
			break
		case "SpamStatus":
			updateStatus(msg)
			break
		}
	}
}
