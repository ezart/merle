// Copyright 2021 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

let status

function sendForThings() {
	conn.send(JSON.stringify({Msg: "GetThings"}))
}

function show(id) {
	var iframe = document.getElementById("thing")
	iframe.src = "/" + encodeURIComponent(id)
}

function addThing(thing) {
	var iframe = document.getElementById("thing")
	var things = document.getElementById("things")
	var newdiv = document.createElement("div")
	var newpre = document.createElement("pre")
	var newimg = document.createElement("img")

	newpre.innerText = thing.Name
	newpre.id = "pre-" + thing.Id

	newimg.src = "/web/images/" + thing.Model + "/" + thing.Status + ".jpg"
	newimg.onclick = function (){show(thing.Id);}
	newimg.id = thing.Id

	newdiv.appendChild(newpre)
	newdiv.appendChild(newimg)
	things.appendChild(newdiv)

	if (iframe.src == "") {
		show(thing.Id)
	}
}

function saveThings(msg) {
	if (msg.Things != null) {
		for (const thing of msg.Things) {
			addThing(thing)
		}
	}
}

function updateStatus(msg) {
	status = msg.Status
}

function Run(scheme, host, id) {

	conn = new WebSocket(scheme + host + "/ws/" + id)

	conn.onopen = function(evt) {
		sendForThings()
	}

	conn.onclose = function(evt) {
		location.reload(true)
	}

	conn.onmessage = function(evt) {
		var msg = JSON.parse(evt.data)

		console.log('event', msg)

		switch(msg.Msg) {
		case "ReplyThings":
			saveThings(msg)
			break
		case "SpamStatus":
			updateStatus(msg)
			break
		}
	}
}
