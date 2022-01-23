// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

var hubId

function sendForChildren() {
	conn.send(JSON.stringify({Msg: "_GetChildren"}))
}

function show(id) {
	var iframe = document.getElementById("child")
	iframe.src = "/" + encodeURIComponent(id)
}

function showIcon(msg) {
	var iframe = document.getElementById("child")
	var children = document.getElementById("children")
	var newdiv = document.createElement("div")
	var newpre = document.createElement("pre")
	var newimg = document.createElement("img")

	newpre.innerText = msg.Name
	newpre.id = "pre-" + msg.Id

	newimg.src = "/" + hubId + "/assets/images/" + msg.Status + ".jpg"
	newimg.onclick = function (){show(msg.Id);}
	newimg.id = msg.Id

	newdiv.appendChild(newpre)
	newdiv.appendChild(newimg)
	children.appendChild(newdiv)
}

function addChild(msg) {
	var iframe = document.getElementById("child")

	showIcon(msg)

	if (iframe.src == "") {
		show(msg.Id)
	}
}

function saveChildren(msg) {
	if (msg.Children != null) {
		for (const child of msg.Children) {
			addChild(child)
		}
	}
}

function updateStatus(msg) {
	var img = document.getElementById(msg.Id)
	var pre = document.getElementById("pre-" + msg.Id)

	if (img == null) {
		addChild(msg)
	} else {
		img.src = "/" + hubId + "/assets/images/" + msg.Status + ".jpg"
		pre.innerText = msg.Name
		show(msg.Id)
	}
}

function Run(scheme, host, id) {

	hubId = id

	conn = new WebSocket(scheme + host + "/ws/" + id)

	conn.onopen = function(evt) {
		sendForChildren()
	}

	conn.onclose = function(evt) {
		location.reload(true)
	}

	conn.onmessage = function(evt) {
		var msg = JSON.parse(evt.data)

		console.log('hub msg', msg)

		switch(msg.Msg) {
		case "ReplyChildren":
			saveChildren(msg)
			break
		case "_SpamStatus":
			updateStatus(msg)
			break
		}
	}
}
