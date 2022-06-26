// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

var hubId

function show(id) {
	var iframe = document.getElementById("child")
	iframe.src = "/" + encodeURIComponent(id)
}

function iconName(child) {
	if (child.Connected) {
		return "connected"
	} else {
		return "disconnected"
	}
}

function showIcon(child) {
	var iframe = document.getElementById("child")
	var children = document.getElementById("children")
	var newdiv = document.createElement("div")
	var newpre = document.createElement("pre")
	var newimg = document.createElement("img")

	newpre.innerText = child.Id
	newpre.id = "pre-" + child.Id

	newimg.src = "/" + hubId + "/assets/images/" + iconName(child) + ".jpg"
	newimg.onclick = function (){show(child.Id);}
	newimg.id = child.Id

	newdiv.appendChild(newpre)
	newdiv.appendChild(newimg)
	children.appendChild(newdiv)
}

function addChild(child) {
	showIcon(child)
	show(child.Id)
}

function clearScreen() {
	var children = document.getElementById("children")
	var iframe = document.getElementById("child")

	iframe.src = ""
	while (children.firstChild) {
		children.removeChild(children.firstChild)
	}
}

function saveState(msg) {
	for (const id in msg.Children) {
		child = msg.Children[id]
		addChild(child)
	}
}

function update(msg) {
	var child = msg.Child
	var img = document.getElementById(child.Id)
	var pre = document.getElementById("pre-" + child.Id)

	if (img == null) {
		addChild(child)
	} else {
		img.src = "/" + hubId + "/assets/images/" + iconName(child) + ".jpg"
		show(child.Id)
	}
}

function Run(ws, id) {

	hubId = id

	var conn

	function connect() {
		conn = new WebSocket(ws)

		conn.onopen = function(evt) {
			clearScreen()
			conn.send(JSON.stringify({Msg: "_GetState"}))
		}

		conn.onclose = function(evt) {
			clearScreen()
			console.log('websocket close', evt.reason)
			setTimeout(connect, 1000)
		}

		conn.onerror = function(err) {
			console.log('websocket error', err.message)
			conn.close()
		}

		conn.onmessage = function(evt) {
			var msg = JSON.parse(evt.data)

			console.log('msg', msg)

			switch(msg.Msg) {
			case "_ReplyState":
				saveState(msg)
				break
			case "Update":
				update(msg)
				break
			}
		}
	}

	connect()
}
