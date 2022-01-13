// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

function sendForOnlyChild() {
	conn.send(JSON.stringify({Msg: "GetOnlyChild"}))
}

function showOnlyChild(msg) {
	var iframe = document.getElementById("child")
	// force refresh of child by using getTime() trick
	iframe.src = "/" + encodeURIComponent(msg.Id) + "?t=" + new Date().getTime()
}

function Run(scheme, host, id) {

	conn = new WebSocket(scheme + host + "/ws/" + id)

	conn.onopen = function(evt) {
		sendForOnlyChild()
	}

	conn.onclose = function(evt) {
		location.reload(true)
	}

	conn.onmessage = function(evt) {
		var msg = JSON.parse(evt.data)

		console.log('prime msg', msg)

		switch(msg.Msg) {
		case "ReplyOnlyChild":
		case "SpamStatus":
			showOnlyChild(msg)
			break
		}
	}
}
