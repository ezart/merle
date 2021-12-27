// Copyright 2021 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

let conn

let id
let model
let name
let status
let ledState
let paused = false

function sendForIdentity() {
	conn.send(JSON.stringify({Msg: "GetIdentity"}))
}

function saveIdentity(msg) {
	status = msg.Status
	id = msg.Id
	model = msg.Model
	name = msg.Name
}

function sendForPaused() {
	conn.send(JSON.stringify({Msg: "GetPaused"}))
}

function savePaused(msg) {
	paused = msg.Paused
}

function saveLedState(msg) {
	ledState = msg.State
}

function refreshBase() {
	var labels = document.getElementById("lables")
	var preId = document.getElementById("id")
	var preModel = document.getElementById("model")
	var preName = document.getElementById("name")

	preId.textContent = id
	preModel.textContent = model
	preName.textContent = name
	labels.className = "labels"
}

function refreshLed() {
	var image = document.getElementById("raspi")
	on = "off"
	if (ledState) {
		on = "on"
	}
	// force refresh of image by using getTime() trick
	image.src = "./web/images/" + model + "/led-gpio17-" + on + ".png?t=" +
		new Date().getTime()
	image.style.visibility = "visible"
}

function refreshButton() {
	var button = document.getElementById("pause")
	if (paused) {
		button.textContent = "Resume"
	} else {
		button.textContent = "Pause"
	}
	button.style.visibility = "visible"
}

function refreshAll() {
	refreshBase()
	refreshLed()
	refreshButton()
}

function updateStatus(msg) {
	status = msg.Status
	refreshAll()
}

function pause() {
	if (paused) {
		conn.send(JSON.stringify({Msg: "CmdResume"}))
	} else {
		conn.send(JSON.stringify({Msg: "CmdPause"}))
	}
	paused = !paused
	refreshButton()
}

function Run(scheme, host, id) {

	conn = new WebSocket(scheme + host + "/ws/" + id)

	conn.onopen = function(evt) {
		sendForIdentity()
	}

	conn.onclose = function(evt) {
		location.reload(true)
	}

	conn.onmessage = function(evt) {
		var msg = JSON.parse(evt.data)

		console.log('event', msg)

		switch(msg.Msg) {
		case "RespIdentity":
			saveIdentity(msg)
			sendForPaused()
			break
		case "RespPaused":
			savePaused(msg)
			refreshAll()
			break
		case "SpamState":
			saveLedState(msg)
			refreshLed()
			break
		case "CmdPause":
			paused = true
			refreshButton()
			break
		case "CmdResume":
			paused = false
			refreshButton()
			break
		case "SpamStatus":
			updateStatus(msg)
			break
		}
	}
}
