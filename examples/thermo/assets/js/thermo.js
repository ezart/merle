// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

var thermoId

function clearScreen() {
}

function saveState(msg) {
}

function update(child) {
}

function Run(ws, id) {

	thermoId = id

	var conn

	function connect() {
		conn = new WebSocket(ws)

		conn.onopen = function(evt) {
			clearScreen()
			conn.send(JSON.stringify({Msg: "_GetState"}))
		}

		conn.onclose = function(evt) {
			clearScreen()
			setTimeout(connect, 1000)
		}

		conn.onerror = function(err) {
			conn.close()
		}

		conn.onmessage = function(evt) {
			var msg = JSON.parse(evt.data)

			console.log('thermo', msg)

			switch(msg.Msg) {
			case "_ReplyState":
				saveState(msg)
				break
			case "_EventStatus":
				update(msg)
				break
			}
		}
	}

	connect()
}
