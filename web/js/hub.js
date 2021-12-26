// Copyright 2021 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

function sendForDevices() {
	conn.send(JSON.stringify({Msg: "devices"}))
}

function updateStatus(device) {
}

function Run(scheme, host, id) {

	conn = new WebSocket(scheme + host + "/ws/" + id)

	conn.onopen = function(evt) {
		sendForDevices()
	}

	conn.onclose = function(evt) {
		location.reload(true)
	}

	conn.onmessage = function(evt) {
		var msg = JSON.parse(evt.data)

		console.log('event', msg)

		switch(msg.Msg) {
		case "devicesResp":
			break
		case "status":
			updateStatus(msg)
			break
		}
	}
}
