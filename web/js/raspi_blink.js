// Copyright 2021 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

let conn

let id
let model
let name
let status
let ledState

function sendForIdentity() {
	var cmd = {Type: "identify"}
	conn.send(JSON.stringify(cmd))
}

function saveIdentity(msg) {
	status = msg.Status
	id = msg.Id
	model = msg.Model
	name = msg.Name
}

function saveLedState(msg) {
	ledState = msg.State
}

function refreshBase() {
	var labels = document.getElementById("lables")
	var preId = document.getElementById("id")
	var preModel = document.getElementById("model")
	var preName = document.getElementById("name")

	labels.className = "labels"
	preId.textContent = id
	preModel.textContent = model
	preName.textContent = name
}

function refreshLed() {
	var image = document.getElementById("raspi")
	image.style.visibility = "visible"
	on = "off"
	if (ledState) {
		on = "on"
	}
	// force refresh of image by using getTime() trick
	image.src = "./web/images/" + model + "/led-gpio17-" + on + ".png?t=" +
		new Date().getTime()
}

function refreshAll() {
	refreshBase()
	refreshLed()
}

function updateStatus(msg) {
	status = msg.Status
	refreshAll()
}

function pause() {
	var cmd = {Type: "pause"}
	conn.send(JSON.stringify(cmd))
}

function resume() {
	var cmd = {Type: "resume"}
	conn.send(JSON.stringify(cmd))
}

function Run(scheme, host, id) {

	conn = new WebSocket(scheme + host + "/ws?id=" + id)

	conn.onopen = function(evt) {
		sendForIdentity()
	}

	conn.onclose = function(evt) {
		location.reload(true)
	}

	conn.onmessage = function(evt) {
		var msg = JSON.parse(evt.data)

		console.log('event', msg)

		switch(msg.Type) {
		case "identity":
			saveIdentity(msg)
			refreshAll()
			break;
		case "ledState":
			saveLedState(msg)
			refreshLed()
			break;
		}
	}
}
