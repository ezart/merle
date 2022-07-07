// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

var thermoId

function clear() {
}

function refresh(msg) {
}

function Run(ws, id) {

	thermoId = id

	var conn

	function connect() {
		conn = new WebSocket(ws)

		conn.onopen = function(evt) {
			clear()
			conn.send(JSON.stringify({Msg: "_GetState"}))
		}

		conn.onclose = function(evt) {
			clear()
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
				refresh(msg)
				break
			}
		}
	}

	connect()
}
