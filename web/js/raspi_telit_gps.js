// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

let conn

function sendForLocation() {
	conn.send(JSON.stringify({Msg: "GetLocation"}))
}

function updateLocation(msg) {
	var preLoc = document.getElementById("location")
	preLoc.textContent = msg.Location
}

function Run(scheme, host, id) {

	conn = new WebSocket(scheme + host + "/ws/" + id)

	conn.onopen = function(evt) {
		sendForLocation()
	}

	conn.onclose = function(evt) {
		location.reload(true)
	}

	conn.onmessage = function(evt) {
		var msg = JSON.parse(evt.data)

		console.log('thing msg', msg)

		switch(msg.Msg) {
		case "Location":
			updateLocation(msg)
			break
		}
	}
}
