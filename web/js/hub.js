// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

let shadowMode

function sendForChildren() {
	conn.send(JSON.stringify({Msg: "GetChildren"}))
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

	newimg.src = "/web/images/" + msg.Model + "/" + msg.Status + ".jpg"
	newimg.onclick = function (){show(msg.Id);}
	newimg.id = msg.Id

	newdiv.appendChild(newpre)
	newdiv.appendChild(newimg)
	children.appendChild(newdiv)
}

function addChild(msg) {
	var iframe = document.getElementById("child")

	if (!shadowMode) {
		showIcon(msg)
	}
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
	if (shadowMode) {
		show(msg.Id)
		return
	}

	var img = document.getElementById(msg.Id)
	var pre = document.getElementById("pre-" + msg.Id)

	if (img == null) {
		addChild(msg)
	} else {
		img.src = "/web/images/" + msg.Model + "/" + msg.Status + ".jpg"
		pre.innerText = msg.Name
		show(msg.Id)
	}
}

function Run(scheme, host, id, max) {

	shadowMode = (max == 1)

	if (!shadowMode) {
		document.getElementById("children").style.height = "100px"
	}

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
		case "SpamStatus":
			updateStatus(msg)
			break
		}
	}
}
