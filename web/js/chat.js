// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

let conn
let chatId

function sendText(text) {
	conn.send(JSON.stringify({Msg: "CmdText", Text: text}))
}

function showText(text) {
	var newpre = document.createElement("pre")
	var history = document.getElementById("history")

	newpre.innerText = text
	history.appendChild(newpre)
}

function enter(event) {
	if (event.keyCode == 13)
		document.getElementById('send').click()
}

function send() {
	var button = document.getElementById("send")
	var input = document.getElementById("input")

	if (input.value == "") {
		return
	}

	text = "[" + chatId + "] " + input.value
	sendText(text)
	showText(text)
}

function Run(scheme, host, id) {

	chatId = id
	conn = new WebSocket(scheme + host + "/ws/" + id)

	conn.onclose = function(evt) {
		location.reload(true)
	}

	conn.onmessage = function(evt) {
		var msg = JSON.parse(evt.data)

		console.log('thing msg', msg)

		switch(msg.Msg) {
		case "CmdText":
			showText(msg.Text)
			break
		}
	}
}
