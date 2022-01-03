// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

function Run(scheme, host, id) {

	conn = new WebSocket(scheme + host + "/ws/" + id)

	conn.onopen = function(evt) {
	}

	conn.onclose = function(evt) {
		location.reload(true)
	}

	conn.onmessage = function(evt) {
		var msg = JSON.parse(evt.data)

		console.log('bridge msg', msg)

		trace = document.getElementById("trace")
		trace.innerHtml += msg
	}
}
