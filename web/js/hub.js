// Copyright 2021 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

function show(id) {
	var iframe = document.getElementById("device")
	iframe.src = "/?id=" + encodeURIComponent(id)
}

function updateStatus(device) {
	var img = document.getElementById(device.Id)
	var pre = document.getElementById("pre-" + device.Id)

	if (img == null) {
		addDevice(device)
	} else {
		img.src = "/web/images/" + device.Model + "/" + device.Status + ".jpg"
		pre.innerText = device.Name
		show(device.Id)
	}
}

function addDevice(device) {
	var iframe = document.getElementById("device")
	var devices = document.getElementById("devices")
	var newdiv = document.createElement("div")
	var newpre = document.createElement("pre")
	var newimg = document.createElement("img")

	newpre.innerText = device.Name
	newpre.id = "pre-" + device.Id

	newimg.src = "/web/images/" + device.Model + "/" + device.Status + ".jpg"
	newimg.onclick = function (){show(device.Id);}
	newimg.id = device.Id

	newdiv.appendChild(newpre)
	newdiv.appendChild(newimg)
	devices.appendChild(newdiv)

	if (iframe.src == "") {
		show(device.Id)
	}
}

function Run(host) {
	var conn = new WebSocket("wss://" + host + "/ws")

	conn.onopen = function(evt) {
		var cmd = {Type: "cmd", Cmd: "Devices"}
		conn.send(JSON.stringify(cmd))
	}

	conn.onclose = function(evt) {
		location.reload(true)
	}

	conn.onmessage = function(evt) {
		var msg = JSON.parse(evt.data)

		console.log('event', msg)

		switch(msg.Type) {

		case "resp":

			switch(msg.Cmd) {

			case "Devices":
				if (msg.Devices != null) {
					for (const device of msg.Devices) {
						console.log('device', device)
						addDevice(device)
					}
				}
				break
			}

			break

		case "spam":

			switch(msg.Spam) {

			case "Status":
				updateStatus(msg)
				break
			}

			break

		}
	}
}
